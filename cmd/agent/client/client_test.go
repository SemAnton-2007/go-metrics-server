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
