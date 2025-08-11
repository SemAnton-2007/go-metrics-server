package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBConnection(t *testing.T) {
	dsn := "postgres://praktikum:praktikum@localhost:5432/praktikum?sslmode=disable"

	db, err := New(dsn)
	require.NoError(t, err, "Failed to connect to database")
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database connection: %v", err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	assert.NoError(t, err, "Database ping failed")
}

func TestDB_Ping(t *testing.T) {
	dsn := "postgres://praktikum:praktikum@localhost:5432/praktikum?sslmode=disable"
	db, err := New(dsn)
	require.NoError(t, err, "Failed to connect to database")
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Failed to close database connection: %v", err)
		}
	}()

	t.Run("successful ping", func(t *testing.T) {
		ctx := context.Background()
		err = db.Ping(ctx)
		assert.NoError(t, err, "Ping should succeed with valid context")
	})

	t.Run("ping with canceled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err = db.Ping(ctx)
		assert.Error(t, err, "Ping should fail with canceled context")
	})
}

func TestDBConnectionError(t *testing.T) {
	_, err := New("invalid_dsn")
	assert.Error(t, err)
}
