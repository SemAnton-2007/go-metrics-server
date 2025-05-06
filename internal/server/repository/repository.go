package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-server/internal/models"
	"os"
	"sync"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	maxRetries = 3
	retryDelay = time.Second
)

type MetricRepository interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	GetGauge(ctx context.Context, name string) (float64, error)
	GetCounter(ctx context.Context, name string) (int64, error)
	GetAllMetrics(ctx context.Context) (map[string]interface{}, error)
	UpdateMetrics(ctx context.Context, metrics []models.Metrics) error
	SaveToFile(ctx context.Context, filename string) error
	LoadFromFile(ctx context.Context, filename string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) (*PostgresRepository, error) {
	repo := &PostgresRepository{db: db}
	if err := repo.createTablesWithRetry(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}
	return repo, nil
}

func (r *PostgresRepository) createTables(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS gauges (
			name TEXT PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create gauges table: %w", err)
	}

	_, err = r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS counters (
			name TEXT PRIMARY KEY,
			value BIGINT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create counters table: %w", err)
	}
	return nil
}

func (r *PostgresRepository) createTablesWithRetry(ctx context.Context) error {
	var lastErr error
	delays := []time.Duration{retryDelay, 3 * retryDelay, 5 * retryDelay}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := r.createTables(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetryableDBError(err) {
			break
		}
	}

	return fmt.Errorf("after %d attempts, last error: %w", maxRetries+1, lastErr)
}

func (r *PostgresRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = EXCLUDED.value
	`, name, value)
	return err
}

func (r *PostgresRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = counters.value + EXCLUDED.value
	`, name, value)
	return err
}

func (r *PostgresRepository) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := r.db.QueryRowContext(ctx, `
		SELECT value FROM gauges WHERE name = $1
	`, name).Scan(&value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("gauge not found")
		}
		return 0, fmt.Errorf("failed to get gauge: %w", err)
	}
	return value, nil
}

func (r *PostgresRepository) GetCounter(ctx context.Context, name string) (int64, error) {
	var value int64
	err := r.db.QueryRowContext(ctx, `
		SELECT value FROM counters WHERE name = $1
	`, name).Scan(&value)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errors.New("counter not found")
		}
		return 0, fmt.Errorf("failed to get counter: %w", err)
	}
	return value, nil
}

func (r *PostgresRepository) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	var errs []error

	// Получаем gauge метрики
	rows, err := r.db.QueryContext(ctx, `SELECT name, value FROM gauges`)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get gauges: %w", err))
	} else {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err != nil {
				errs = append(errs, fmt.Errorf("failed to scan gauge row: %w", err))
				continue
			}
			metrics[name] = value
		}
		if err := rows.Err(); err != nil {
			errs = append(errs, fmt.Errorf("error after iterating gauge rows: %w", err))
		}
	}

	// Получаем counter метрики
	rows, err = r.db.QueryContext(ctx, `SELECT name, value FROM counters`)
	if err != nil {
		errs = append(errs, fmt.Errorf("failed to get counters: %w", err))
	} else {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value int64
			if err := rows.Scan(&name, &value); err != nil {
				errs = append(errs, fmt.Errorf("failed to scan counter row: %w", err))
				continue
			}
			metrics[name] = value
		}
		if err := rows.Err(); err != nil {
			errs = append(errs, fmt.Errorf("error after iterating counter rows: %w", err))
		}
	}

	if len(errs) > 0 {
		return metrics, fmt.Errorf("partial metrics retrieved with errors: %v", errors.Join(errs...))
	}
	return metrics, nil
}

func (r *PostgresRepository) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				continue
			}
			_, err := tx.ExecContext(ctx, `
				INSERT INTO gauges (name, value)
				VALUES ($1, $2)
				ON CONFLICT (name)
				DO UPDATE SET value = EXCLUDED.value
			`, metric.ID, *metric.Value)
			if err != nil {
				return fmt.Errorf("failed to update gauge %s: %w", metric.ID, err)
			}

		case "counter":
			if metric.Delta == nil {
				continue
			}
			_, err := tx.ExecContext(ctx, `
				INSERT INTO counters (name, value)
				VALUES ($1, $2)
				ON CONFLICT (name)
				DO UPDATE SET value = counters.value + EXCLUDED.value
			`, metric.ID, *metric.Delta)
			if err != nil {
				return fmt.Errorf("failed to update counter %s: %w", metric.ID, err)
			}
		}
	}
	return tx.Commit()
}

func (r *PostgresRepository) SaveToFile(ctx context.Context, filename string) error {
	return nil
}

func (r *PostgresRepository) LoadFromFile(ctx context.Context, filename string) error {
	return nil
}

type MemoryRepository struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (r *MemoryRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gauges[name] = value
	return nil
}

func (r *MemoryRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.counters[name] += value
	return nil
}

func (r *MemoryRepository) GetGauge(ctx context.Context, name string) (float64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if value, ok := r.gauges[name]; ok {
		return value, nil
	}
	return 0, errors.New("gauge not found")
}

func (r *MemoryRepository) GetCounter(ctx context.Context, name string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if value, ok := r.counters[name]; ok {
		return value, nil
	}
	return 0, errors.New("counter not found")
}

func (r *MemoryRepository) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	metrics := make(map[string]interface{})
	for name, value := range r.gauges {
		metrics[name] = value
	}
	for name, value := range r.counters {
		metrics[name] = value
	}
	return metrics, nil
}

func (r *MemoryRepository) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value != nil {
				r.gauges[metric.ID] = *metric.Value
			}
		case "counter":
			if metric.Delta != nil {
				r.counters[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}

func (r *MemoryRepository) SaveToFile(ctx context.Context, filename string) error {
	if filename == "" {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	data := struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{
		Gauges:   r.gauges,
		Counters: r.counters,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (r *MemoryRepository) LoadFromFile(ctx context.Context, filename string) error {
	if filename == "" {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	var data struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	r.gauges = data.Gauges
	r.counters = data.Counters
	return nil
}

func isRetryableDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == pgerrcode.ConnectionException ||
			pgErr.Code == pgerrcode.ConnectionDoesNotExist ||
			pgErr.Code == pgerrcode.ConnectionFailure ||
			pgErr.Code == pgerrcode.SQLClientUnableToEstablishSQLConnection ||
			pgErr.Code == pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection ||
			pgErr.Code == pgerrcode.TransactionResolutionUnknown
	}
	return false
}
