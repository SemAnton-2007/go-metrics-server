package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-metrics-server/internal/models"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	maxRetries = 3
	retryDelay = time.Second
)

type PostgresStorage struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresStorage(ctx context.Context, db *sql.DB) (*PostgresStorage, error) {
	storage := &PostgresStorage{db: db, ctx: ctx}

	// Проверяем соединение и права
	if err := storage.checkConnection(); err != nil {
		return nil, fmt.Errorf("database connection check failed: %w", err)
	}

	if err := storage.createTablesWithRetry(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return storage, nil
}

func (s *PostgresStorage) checkConnection() error {
	// Проверяем, что у пользователя есть права на создание таблиц
	_, err := s.db.ExecContext(s.ctx, `SELECT 1`)
	if err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	return nil
}

func (s *PostgresStorage) createTables() error {
	// Создаем таблицы напрямую в схеме public (по умолчанию)
	_, err := s.db.ExecContext(s.ctx, `
		CREATE TABLE IF NOT EXISTS gauges (
			name TEXT PRIMARY KEY,
			value DOUBLE PRECISION NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create gauges table: %w", err)
	}

	_, err = s.db.ExecContext(s.ctx, `
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

func (s *PostgresStorage) createTablesWithRetry() error {
	var lastErr error
	delays := []time.Duration{retryDelay, 3 * retryDelay, 5 * retryDelay}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := s.createTables()
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

func isRetryableDBError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Connection exceptions (Class 08)
		return pgErr.Code == pgerrcode.ConnectionException ||
			pgErr.Code == pgerrcode.ConnectionDoesNotExist ||
			pgErr.Code == pgerrcode.ConnectionFailure ||
			pgErr.Code == pgerrcode.SQLClientUnableToEstablishSQLConnection ||
			pgErr.Code == pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection ||
			pgErr.Code == pgerrcode.TransactionResolutionUnknown
	}
	return false
}

func (s *PostgresStorage) UpdateGauge(name string, value float64) {
	if err := s.updateWithRetry(func() error {
		_, err := s.db.ExecContext(s.ctx, `
			INSERT INTO gauges (name, value)
			VALUES ($1, $2)
			ON CONFLICT (name)
			DO UPDATE SET value = EXCLUDED.value
		`, name, value)
		return err
	}); err != nil {
		fmt.Printf("Failed to update gauge: %v\n", err)
	}
}

func (s *PostgresStorage) UpdateCounter(name string, value int64) {
	if err := s.updateWithRetry(func() error {
		_, err := s.db.ExecContext(s.ctx, `
			INSERT INTO counters (name, value)
			VALUES ($1, $2)
			ON CONFLICT (name)
			DO UPDATE SET value = counters.value + EXCLUDED.value
		`, name, value)
		return err
	}); err != nil {
		fmt.Printf("Failed to update counter: %v\n", err)
	}
}

func (s *PostgresStorage) GetGauge(name string) (float64, error) {
	var value float64
	err := s.db.QueryRowContext(s.ctx, `
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

func (s *PostgresStorage) GetCounter(name string) (int64, error) {
	var value int64
	err := s.db.QueryRowContext(s.ctx, `
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

func (s *PostgresStorage) GetAllMetrics() (map[string]interface{}, error) {
	metrics := make(map[string]interface{})
	var errs []error

	// Получаем gauge метрики
	rows, err := s.db.QueryContext(s.ctx, `SELECT name, value FROM gauges`)
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
	rows, err = s.db.QueryContext(s.ctx, `SELECT name, value FROM counters`)
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

	// Если были ошибки, возвращаем частичный результат с ошибкой
	if len(errs) > 0 {
		return metrics, fmt.Errorf("partial metrics retrieved with errors: %v", errors.Join(errs...))
	}

	return metrics, nil
}

// DatabaseRunner интерфейс для выполнения SQL запросов
type DatabaseRunner interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

func (s *PostgresStorage) UpdateMetrics(metrics []models.Metrics) error {
	tx, err := s.db.BeginTx(s.ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := s.updateMetrics(tx, metrics); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *PostgresStorage) updateMetrics(runner DatabaseRunner, metrics []models.Metrics) error {
	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				continue
			}
			_, err := runner.ExecContext(s.ctx, `
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
			_, err := runner.ExecContext(s.ctx, `
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
	return nil
}

func (s *PostgresStorage) updateWithRetry(fn func() error) error {
	var lastErr error
	delays := []time.Duration{retryDelay, 3 * retryDelay, 5 * retryDelay}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := fn()
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

func (s *PostgresStorage) SaveToFile(filename string) error {
	return nil
}

func (s *PostgresStorage) LoadFromFile(filename string) error {
	return nil
}
