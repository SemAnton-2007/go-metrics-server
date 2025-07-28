package webservers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	"go-metrics-server/internal/server/repository"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{ServerAddr: "localhost:8080"}
	repo := repository.NewMemoryRepository()
	srv := NewServer(cfg, repo, nil) // nil для DB, так как тестируем без БД

	// Тест 1: Проверка маршрута /update/
	ts := httptest.NewServer(srv.Handler)
	defer func() {
		ts.Close()
	}()

	t.Run("update gauge metric", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/update/gauge/test/123.45", "text/plain", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	})

	t.Run("update counter metric", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/update/counter/test/10", "text/plain", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	})

	t.Run("ping endpoint with nil DB", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/ping")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		if err := resp.Body.Close(); err != nil {
			t.Errorf("Failed to close response body: %v", err)
		}
	})
}

func TestPingHandler(t *testing.T) {
	cfg := &config.Config{ServerAddr: "localhost:8080"}
	repo := repository.NewMemoryRepository()
	mockDB := &database.DB{}
	srv := NewServer(cfg, repo, mockDB)

	ts := httptest.NewServer(srv.Handler)
	defer func() {
		ts.Close()
	}()
}
