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
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	}
}

func TestDB_Ping(t *testing.T) {
	dsn := "postgres://praktikum:praktikum@localhost:5432/praktikum?sslmode=disable"
	db, err := New(dsn)
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()
	err = db.Ping(ctx)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = db.Ping(ctx)
	assert.Error(t, err)
}
