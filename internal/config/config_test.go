package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetConfigFile(t *testing.T) {
	t.Run("from env", func(t *testing.T) {
		os.Setenv("CONFIG", "test.json")
		defer os.Unsetenv("CONFIG")
		assert.Equal(t, "test.json", GetConfigFile())
	})

	t.Run("from args", func(t *testing.T) {
		oldArgs := os.Args
		defer func() { os.Args = oldArgs }()
		os.Args = []string{"cmd", "-config", "config.json"}
		assert.Equal(t, "config.json", GetConfigFile())
	})

	t.Run("not set", func(t *testing.T) {
		assert.Empty(t, GetConfigFile())
	})
}
