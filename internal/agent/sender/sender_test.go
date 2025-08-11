package sender

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/internal/agent/config"
	"go-metrics-server/internal/models"

	"github.com/stretchr/testify/assert"
)

func getTestConfig() *config.Config {
	return &config.Config{}
}

func TestSender_SendMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := New(ts.URL, "", getTestConfig())
	err := s.SendMetric("gauge", "test", 123.45)
	assert.NoError(t, err)

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s = New(ts.URL, "", getTestConfig())
	err = s.SendMetric("gauge", "test", 123.45)
	assert.Error(t, err)
}

func TestSendMetricJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := New(ts.URL, "", getTestConfig())

	t.Run("send gauge", func(t *testing.T) {
		err := s.SendMetric("gauge", "test", 1.23)
		assert.NoError(t, err)
	})

	t.Run("send counter", func(t *testing.T) {
		err := s.SendMetric("counter", "test", int64(10))
		assert.NoError(t, err)
	})
}

func TestHashHeader(t *testing.T) {
	var receivedHash string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHash = r.Header.Get("HashSHA256")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	t.Run("without key", func(t *testing.T) {
		s := New(ts.URL, "", getTestConfig())
		err := s.SendMetric("gauge", "test", 1.23)
		assert.NoError(t, err)
		assert.Empty(t, receivedHash)
	})

	t.Run("with key", func(t *testing.T) {
		s := New(ts.URL, "testkey", getTestConfig())
		err := s.SendMetric("gauge", "test", 1.23)
		assert.NoError(t, err)
		assert.NotEmpty(t, receivedHash)
	})
}

func TestSendMetricsBatch(t *testing.T) {
	t.Run("successful batch send", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/updates/", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		s := New(ts.URL, "", getTestConfig())
		err := s.SendMetricsBatch(map[string]interface{}{
			"gauge1":   1.23,
			"counter1": int64(10),
		})
		assert.NoError(t, err)
	})

	t.Run("empty batch", func(t *testing.T) {
		s := New("http://localhost:8080", "", getTestConfig())
		err := s.SendMetricsBatch(map[string]interface{}{})
		assert.NoError(t, err)
	})
}

func TestSendRequestErrors(t *testing.T) {
	t.Run("request creation error", func(t *testing.T) {
		s := New("invalid_url", "", getTestConfig())
		err := s.sendRequest("/update/", []models.Metrics{{}})
		assert.Error(t, err)
	})

	t.Run("non-retryable error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer ts.Close()

		s := New(ts.URL, "", getTestConfig())
		err := s.sendWithRetry("/update/", []models.Metrics{{}})
		assert.Error(t, err)
	})
}
