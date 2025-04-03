package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-metrics-server/internal/models"
)

type PostgresStorage struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresStorage(db *sql.DB) (*PostgresStorage, error) {
	ctx := context.Background()
	storage := &PostgresStorage{db: db, ctx: ctx}

	// Проверяем соединение и права
	if err := storage.checkConnection(); err != nil {
		return nil, fmt.Errorf("database connection check failed: %w", err)
	}

	if err := storage.createTables(); err != nil {
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

func (s *PostgresStorage) UpdateGauge(name string, value float64) {
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO gauges (name, value) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET value = EXCLUDED.value
	`, name, value)
	if err != nil {
		fmt.Printf("Failed to update gauge: %v\n", err)
	}
}

func (s *PostgresStorage) UpdateCounter(name string, value int64) {
	_, err := s.db.ExecContext(s.ctx, `
		INSERT INTO counters (name, value) 
		VALUES ($1, $2)
		ON CONFLICT (name) 
		DO UPDATE SET value = counters.value + EXCLUDED.value
	`, name, value)
	if err != nil {
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

func (s *PostgresStorage) GetAllMetrics() map[string]interface{} {
	metrics := make(map[string]interface{})

	// Получаем gauge метрики
	rows, err := s.db.QueryContext(s.ctx, `SELECT name, value FROM gauges`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value float64
			if err := rows.Scan(&name, &value); err == nil {
				metrics[name] = value
			}
		}
		// Добавляем проверку ошибок после итерации
		if err := rows.Err(); err != nil {
			fmt.Printf("Error after iterating gauge rows: %v\n", err)
		}
	}

	// Получаем counter метрики
	rows, err = s.db.QueryContext(s.ctx, `SELECT name, value FROM counters`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var name string
			var value int64
			if err := rows.Scan(&name, &value); err == nil {
				metrics[name] = value
			}
		}
		// Добавляем проверку ошибок после итерации
		if err := rows.Err(); err != nil {
			fmt.Printf("Error after iterating counter rows: %v\n", err)
		}
	}

	return metrics
}

func (s *PostgresStorage) SaveToFile(filename string) error {
	// Не реализовано для PostgresStorage
	return nil
}

func (s *PostgresStorage) LoadFromFile(filename string) error {
	// Не реализовано для PostgresStorage
	return nil
}

func (s *PostgresStorage) UpdateMetrics(metrics []models.Metrics) error {
	tx, err := s.db.BeginTx(s.ctx, nil)
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
			_, err = tx.ExecContext(s.ctx, `
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
			_, err = tx.ExecContext(s.ctx, `
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
