package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-metrics-server/cmd/server/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
	storage := storage.NewMemStorage()
	handler := UpdateMetricHandler(storage)

	// Тест 1: Успешное обновление gauge
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123.45", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "123.45")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Тест 2: Успешное обновление counter
	req = httptest.NewRequest(http.MethodPost, "/update/counter/test/10", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "counter")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "10")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "OK", w.Body.String())

	// Тест 3: Некорректный тип метрики
	req = httptest.NewRequest(http.MethodPost, "/update/invalid/test/10", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "invalid")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "10")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 4: Некорректное значение gauge
	req = httptest.NewRequest(http.MethodPost, "/update/gauge/test/invalid", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 5: Некорректное значение counter
	req = httptest.NewRequest(http.MethodPost, "/update/counter/test/invalid", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "counter")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 6: Отсутствует имя метрики
	req = httptest.NewRequest(http.MethodPost, "/update/gauge//10", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "")
	rctx.URLParams.Add("value", "10")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// Новые тесты для GetMetricValueHandler и GetAllMetricsHandler

func TestGetMetricValueHandler(t *testing.T) {
	storage := storage.NewMemStorage()
	handler := GetMetricValueHandler(storage)

	// Тест 1: Метрика не найдена
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Тест 2: Успешное получение метрики gauge
	storage.UpdateGauge("test", 123.45)
	req = httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "123.45", w.Body.String())

	// Тест 3: Успешное получение метрики counter
	storage.UpdateCounter("test", 10)
	req = httptest.NewRequest(http.MethodGet, "/value/counter/test", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "counter")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10", w.Body.String())

	// Тест 4: Некорректный тип метрики
	req = httptest.NewRequest(http.MethodGet, "/value/invalid/test", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "invalid")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllMetricsHandler(t *testing.T) {
	storage := storage.NewMemStorage()
	handler := GetAllMetricsHandler(storage)

	// Тест 1: Успешное получение всех метрик
	storage.UpdateGauge("test_gauge", 123.45)
	storage.UpdateCounter("test_counter", 10)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test_gauge: 123.45")
	assert.Contains(t, w.Body.String(), "test_counter: 10")
}
