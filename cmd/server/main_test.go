package main

import (
	"testing"
	"time"

	commonconfig "go-metrics-server/internal/config"
	serverconfig "go-metrics-server/internal/server/config"
)

func TestRunServer(t *testing.T) {
	cfg := &serverconfig.Config{
		CommonConfig: commonconfig.CommonConfig{
			ServerAddr: ":8080",
		},
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
