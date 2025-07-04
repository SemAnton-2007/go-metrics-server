package main

import (
	"testing"
	"time"

	"go-metrics-server/internal/server/config"
)

func TestRunServer(t *testing.T) {
	cfg := &config.Config{
		ServerAddr:    ":8080",
		StoreInterval: time.Second,
		FileStorage:   "",
		Restore:       false,
		DatabaseDSN:   "",
	}

	go func() {
		RunServer(cfg)
	}()

	time.Sleep(100 * time.Millisecond)
}
