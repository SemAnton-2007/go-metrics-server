package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Драйвер PostgreSQL
)

// DB представляет соединение с PostgreSQL
type DB struct {
	*sql.DB
}

// New создает новое соединение с PostgreSQL
func New(dsn string) (*DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// Ping проверяет соединение с БД
func (db *DB) Ping(ctx context.Context) error {
	return db.DB.PingContext(ctx)
}

// Close закрывает соединение с БД
func (db *DB) Close() error {
	return db.DB.Close()
}
