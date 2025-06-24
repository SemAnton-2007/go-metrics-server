package main

import (
	"testing"
	"time"

	"go-metrics-server/internal/agent/config"
)

func TestRunAgent(t *testing.T) {
	cfg := &config.Config{
		ServerAddr:     "localhost:8080",
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 50 * time.Millisecond,
		RateLimit:      1,
	}

	go func() {
		RunAgent(cfg)
	}()

	time.Sleep(200 * time.Millisecond)
}
