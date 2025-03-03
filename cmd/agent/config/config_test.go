package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// Сохраняем текущие переменные окружения
	oldAddr := os.Getenv("ADDRESS")
	oldPoll := os.Getenv("POLL_INTERVAL")
	oldReport := os.Getenv("REPORT_INTERVAL")
	defer func() {
		os.Setenv("ADDRESS", oldAddr)
		os.Setenv("POLL_INTERVAL", oldPoll)
		os.Setenv("REPORT_INTERVAL", oldReport)
	}()

	// Тест 1: Значения по умолчанию
	os.Unsetenv("ADDRESS")
	os.Unsetenv("POLL_INTERVAL")
	os.Unsetenv("REPORT_INTERVAL")
	cfg := NewConfig()
	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.Equal(t, 2*time.Second, cfg.PollInterval)
	assert.Equal(t, 10*time.Second, cfg.ReportInterval)

	// Тест 2: Переменные окружения
	os.Setenv("ADDRESS", "127.0.0.1:9090")
	os.Setenv("POLL_INTERVAL", "5")
	os.Setenv("REPORT_INTERVAL", "15")
	cfg = NewConfig()
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 15*time.Second, cfg.ReportInterval)

	// Тест 3: Флаги
	// Сохраняем оригинальные os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Устанавливаем аргументы для теста
	os.Args = []string{"cmd", "-a=127.0.0.1:9090", "-p=5", "-r=15"}
	cfg = NewConfig()
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 15*time.Second, cfg.ReportInterval)
}
