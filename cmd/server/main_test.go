package main

import (
	"context"
	"testing"
	"time"

	commonconfig "go-metrics-server/internal/config"
	serverconfig "go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/repository"

	"github.com/stretchr/testify/assert"
)

func TestRunServerComponents(t *testing.T) {
	cfg := &serverconfig.Config{
		CommonConfig: commonconfig.CommonConfig{
			ServerAddr: ":0",
		},
	}

	t.Run("with file storage", func(t *testing.T) {
		cfg.DatabaseDSN = ""
		cfg.FileStorage = "/tmp/metrics.json"
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		go func() {
			RunServer(cfg)
		}()

		<-ctx.Done()
	})

	t.Run("with database", func(t *testing.T) {
		cfg.DatabaseDSN = "postgres://praktikum:praktikum@localhost:5432/praktikum?sslmode=disable"
		cfg.FileStorage = ""
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		go func() {
			RunServer(cfg)
		}()

		<-ctx.Done()
	})
}

func TestSyncSaveRepository(t *testing.T) {
	t.Run("Close saves metrics", func(t *testing.T) {
		repo := newSyncSaveRepository(repository.NewMemoryRepository(), "/tmp/test.json", nil)
		err := repo.Close()
		assert.NoError(t, err)
	})
}
