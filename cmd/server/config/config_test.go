package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// Сохраняем текущие переменные окружения
	oldAddr := os.Getenv("ADDRESS")
	defer func() {
		os.Setenv("ADDRESS", oldAddr)
	}()

	// Тест 1: Значение по умолчанию
	os.Unsetenv("ADDRESS")
	cfg := NewConfig()
	assert.Equal(t, "localhost:8080", cfg.ServerAddr)

	// Тест 2: Переменная окружения
	os.Setenv("ADDRESS", "127.0.0.1:9090")
	cfg = NewConfig()
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)

	// Тест 3: Флаг
	os.Args = []string{"cmd", "-a=127.0.0.1:9090"}
	cfg = NewConfig()
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)
}
