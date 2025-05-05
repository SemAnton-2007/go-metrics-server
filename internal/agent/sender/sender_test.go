package sender

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSender_SendMetric(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	s := New(ts.URL, "")
	err := s.SendMetric("gauge", "test", 123.45)
	assert.NoError(t, err)

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	s = New(ts.URL, "")
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

	s := New(ts.URL, "")

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
		s := New(ts.URL, "")
		err := s.SendMetric("gauge", "test", 1.23)
		assert.NoError(t, err)
		assert.Empty(t, receivedHash)
	})

	t.Run("with key", func(t *testing.T) {
		s := New(ts.URL, "testkey")
		err := s.SendMetric("gauge", "test", 1.23)
		assert.NoError(t, err)
		assert.NotEmpty(t, receivedHash)
	})
}
