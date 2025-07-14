package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	oldAddr := os.Getenv("ADDRESS")
	oldPoll := os.Getenv("POLL_INTERVAL")
	oldReport := os.Getenv("REPORT_INTERVAL")
	defer func() {
		err := os.Setenv("ADDRESS", oldAddr)
		require.NoError(t, err, "Failed to restore ADDRESS env var")

		err = os.Setenv("POLL_INTERVAL", oldPoll)
		require.NoError(t, err, "Failed to restore POLL_INTERVAL env var")

		err = os.Setenv("REPORT_INTERVAL", oldReport)
		require.NoError(t, err, "Failed to restore REPORT_INTERVAL env var")
	}()

	t.Run("default values", func(t *testing.T) {
		err := os.Unsetenv("ADDRESS")
		require.NoError(t, err, "Failed to unset ADDRESS env var")

		err = os.Unsetenv("POLL_INTERVAL")
		require.NoError(t, err, "Failed to unset POLL_INTERVAL env var")

		err = os.Unsetenv("REPORT_INTERVAL")
		require.NoError(t, err, "Failed to unset REPORT_INTERVAL env var")

		cfg := NewConfig()
		assert.Equal(t, "localhost:8080", cfg.ServerAddr)
		assert.Equal(t, 2*time.Second, cfg.PollInterval)
		assert.Equal(t, 10*time.Second, cfg.ReportInterval)
	})

	t.Run("environment variables", func(t *testing.T) {
		err := os.Setenv("ADDRESS", "127.0.0.1:9090")
		require.NoError(t, err, "Failed to set ADDRESS env var")

		err = os.Setenv("POLL_INTERVAL", "5")
		require.NoError(t, err, "Failed to set POLL_INTERVAL env var")

		err = os.Setenv("REPORT_INTERVAL", "15")
		require.NoError(t, err, "Failed to set REPORT_INTERVAL env var")

		cfg := NewConfig()
		assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)
		assert.Equal(t, 5*time.Second, cfg.PollInterval)
		assert.Equal(t, 15*time.Second, cfg.ReportInterval)
	})

	t.Run("command line flags", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()

		os.Args = []string{"cmd", "-a=127.0.0.1:9090", "-p=5", "-r=15"}
		cfg := NewConfig()
		assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)
		assert.Equal(t, 5*time.Second, cfg.PollInterval)
		assert.Equal(t, 15*time.Second, cfg.ReportInterval)
	})
}
