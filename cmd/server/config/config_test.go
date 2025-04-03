package config

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig_Defaults(t *testing.T) {
	// Сохраняем и очищаем env
	oldEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range oldEnv {
			parts := strings.SplitN(env, "=", 2)
			os.Setenv(parts[0], parts[1])
		}
	}()
	os.Clearenv()

	// Сохраняем и сбрасываем аргументы
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"} // Только имя программы

	cfg := NewConfig()
	assert.Equal(t, "localhost:8080", cfg.ServerAddr)
	assert.Equal(t, 300*time.Second, cfg.StoreInterval) // Проверяем реальное значение из кода
	assert.Equal(t, "/tmp/metrics-db.json", cfg.FileStorage)
	assert.True(t, cfg.Restore)
	assert.Equal(t, "", cfg.DatabaseDSN) // Добавили проверку нового поля
}

func TestNewConfig_Flags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Устанавливаем тестовые флаги
	os.Args = []string{"cmd", "-a=0.0.0.0:8080", "-i=30s", "-f=/tmp/test.json", "-r=false", "-d=postgres://user:pass@localhost:5432/db"}

	cfg := NewConfig()
	assert.Equal(t, "0.0.0.0:8080", cfg.ServerAddr)
	assert.Equal(t, 30*time.Second, cfg.StoreInterval)
	assert.Equal(t, "/tmp/test.json", cfg.FileStorage)
	assert.False(t, cfg.Restore)
	assert.Equal(t, "postgres://user:pass@localhost:5432/db", cfg.DatabaseDSN)
}

func TestNewConfig_EnvVars(t *testing.T) {
	oldEnv := os.Environ()
	defer func() {
		os.Clearenv()
		for _, env := range oldEnv {
			parts := strings.SplitN(env, "=", 2)
			os.Setenv(parts[0], parts[1])
		}
	}()
	os.Clearenv()

	os.Setenv("ADDRESS", "127.0.0.1:9090")
	os.Setenv("STORE_INTERVAL", "60s")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/env.json")
	os.Setenv("RESTORE", "false")
	os.Setenv("DATABASE_DSN", "postgres://env@localhost:5432/db")

	cfg := NewConfig()
	assert.Equal(t, "127.0.0.1:9090", cfg.ServerAddr)
	assert.Equal(t, 60*time.Second, cfg.StoreInterval)
	assert.Equal(t, "/tmp/env.json", cfg.FileStorage)
	assert.False(t, cfg.Restore)
	assert.Equal(t, "postgres://env@localhost:5432/db", cfg.DatabaseDSN)
}
