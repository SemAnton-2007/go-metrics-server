package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/cmd/server/storage"

	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
	storage := storage.NewMemStorage()
	handler := UpdateMetricHandler(storage)

	// Тест 1: Успешное обновление gauge
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123.45", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Тест 2: Успешное обновление counter
	req = httptest.NewRequest(http.MethodPost, "/update/counter/test/10", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Тест 3: Некорректный тип метрики
	req = httptest.NewRequest(http.MethodPost, "/update/invalid/test/10", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 4: Некорректное значение gauge
	req = httptest.NewRequest(http.MethodPost, "/update/gauge/test/invalid", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 5: Некорректное значение counter
	req = httptest.NewRequest(http.MethodPost, "/update/counter/test/invalid", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 6: Отсутствует имя метрики
	req = httptest.NewRequest(http.MethodPost, "/update/gauge//10", nil)
	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}
