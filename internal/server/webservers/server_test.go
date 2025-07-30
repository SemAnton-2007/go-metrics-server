package webservers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	commonconfig "go-metrics-server/internal/config"
	serverconfig "go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	"go-metrics-server/internal/server/repository"

	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	cfg := &serverconfig.Config{
		CommonConfig: commonconfig.CommonConfig{
			ServerAddr: "localhost:8080",
		},
	}
	repo := repository.NewMemoryRepository()
	srv := NewServer(cfg, repo, nil)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()

	t.Run("update gauge metric", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/update/gauge/test/123.45", "text/plain", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("update counter metric", func(t *testing.T) {
		resp, err := http.Post(ts.URL+"/update/counter/test/10", "text/plain", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("ping endpoint with nil DB", func(t *testing.T) {
		resp, err := http.Get(ts.URL + "/ping")
		assert.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestPingHandler(t *testing.T) {
	cfg := &serverconfig.Config{
		CommonConfig: commonconfig.CommonConfig{
			ServerAddr: "localhost:8080",
		},
	}
	repo := repository.NewMemoryRepository()
	mockDB := &database.DB{}
	srv := NewServer(cfg, repo, mockDB)

	ts := httptest.NewServer(srv.Handler)
	defer ts.Close()
}
