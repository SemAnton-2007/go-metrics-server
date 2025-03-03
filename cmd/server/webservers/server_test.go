package webservers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/storage"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{ServerAddr: "localhost:8080"}
	storage := storage.NewMemStorage()
	srv := NewServer(cfg, storage)

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
}
