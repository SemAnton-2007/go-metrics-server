package main

import (
	"testing"
	"time"

	agentconfig "go-metrics-server/internal/agent/config"
	commonconfig "go-metrics-server/internal/config"
)

func TestRunAgent(t *testing.T) {
	cfg := &agentconfig.Config{
		CommonConfig: commonconfig.CommonConfig{
			ServerAddr: "localhost:8080",
		},
		PollInterval:   50 * time.Millisecond,
		ReportInterval: 50 * time.Millisecond,
		RateLimit:      1,
	}

	go func() {
		RunAgent(cfg)
	}()

	time.Sleep(200 * time.Millisecond)
}
