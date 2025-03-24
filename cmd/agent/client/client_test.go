package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_SendMetric(t *testing.T) {
	// Тестовый сервер
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Тест 1: Успешная отправка
	client := NewClient(ts.URL)
	err := client.SendMetric("gauge", "test", 123.45)
	assert.NoError(t, err)

	// Тест 2: Ошибка сервера
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	client = NewClient(ts.URL)
	err = client.SendMetric("gauge", "test", 123.45)
	assert.Error(t, err)
}

// Добавляем в конец файла

func TestSendMetricJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/update/", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewClient(ts.URL)

	t.Run("send gauge", func(t *testing.T) {
		err := c.SendMetric("gauge", "test", 1.23)
		assert.NoError(t, err)
	})

	t.Run("send counter", func(t *testing.T) {
		err := c.SendMetric("counter", "test", int64(10))
		assert.NoError(t, err)
	})
}
