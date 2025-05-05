package webservers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	"go-metrics-server/internal/server/storage"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{ServerAddr: "localhost:8080"}
	store := storage.NewMemStorage()
	srv := NewServer(cfg, store, nil) // nil для DB, так как тестируем без БД

	// Тест 1: Проверка маршрута /update/
	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	resp, err := http.Post(ts.URL+"/update/gauge/test/123.45", "text/plain", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp, err = http.Post(ts.URL+"/update/counter/test/10", "text/plain", nil)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	// Тест 3: Проверка маршрута /ping (должен отсутствовать при nil DB)
	resp, err = http.Get(ts.URL + "/ping")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	resp.Body.Close()
}

func TestPingHandler(t *testing.T) {
	cfg := &config.Config{ServerAddr: "localhost:8080"}
	store := storage.NewMemStorage()
	mockDB := &database.DB{}
	srv := NewServer(cfg, store, mockDB)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

}
