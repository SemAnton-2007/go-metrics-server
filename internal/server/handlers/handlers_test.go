package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-metrics-server/internal/models"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/service"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestUpdateMetricHandler(t *testing.T) {
	repo := repository.NewMemoryRepository()
	handler := NewMetricHandler(service.NewMetricService(repo))

	// Тест 1: Успешное обновление gauge
	req := httptest.NewRequest(http.MethodPost, "/update/gauge/test/123.45", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "123.45")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.UpdateMetric(w, req)
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
	handler.UpdateMetric(w, req)
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
	handler.UpdateMetric(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 4: Некорректное значение gauge
	req = httptest.NewRequest(http.MethodPost, "/update/gauge/test/invalid", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.UpdateMetric(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 5: Некорректное значение counter
	req = httptest.NewRequest(http.MethodPost, "/update/counter/test/invalid", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "counter")
	rctx.URLParams.Add("name", "test")
	rctx.URLParams.Add("value", "invalid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.UpdateMetric(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Тест 6: Отсутствует имя метрики
	req = httptest.NewRequest(http.MethodPost, "/update/gauge//10", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "")
	rctx.URLParams.Add("value", "10")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.UpdateMetric(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMetricValueHandler(t *testing.T) {
	repo := repository.NewMemoryRepository()
	handler := NewMetricHandler(service.NewMetricService(repo))

	// Тест 1: Метрика не найдена
	req := httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.GetMetricValue(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Тест 2: Успешное получение метрики gauge
	repo.UpdateGauge(context.Background(), "test", 123.45)
	req = httptest.NewRequest(http.MethodGet, "/value/gauge/test", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "gauge")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.GetMetricValue(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "123.45", w.Body.String())

	// Тест 3: Успешное получение метрики counter
	repo.UpdateCounter(context.Background(), "test", 10)
	req = httptest.NewRequest(http.MethodGet, "/value/counter/test", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "counter")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.GetMetricValue(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "10", w.Body.String())

	// Тест 4: Некорректный тип метрики
	req = httptest.NewRequest(http.MethodGet, "/value/invalid/test", nil)
	rctx = chi.NewRouteContext()
	rctx.URLParams.Add("type", "invalid")
	rctx.URLParams.Add("name", "test")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w = httptest.NewRecorder()
	handler.GetMetricValue(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetAllMetricsHandler(t *testing.T) {
	repo := repository.NewMemoryRepository()
	handler := NewMetricHandler(service.NewMetricService(repo))

	// Тест 1: Успешное получение всех метрик
	repo.UpdateGauge(context.Background(), "test_gauge", 123.45)
	repo.UpdateCounter(context.Background(), "test_counter", 10)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.GetAllMetrics(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "test_gauge: 123.45")
	assert.Contains(t, w.Body.String(), "test_counter: 10")
}

func TestUpdateMetricJSONHandler(t *testing.T) {
	repo := repository.NewMemoryRepository()
	handler := NewMetricHandler(service.NewMetricService(repo))

	t.Run("success gauge update", func(t *testing.T) {
		body := `{"id":"test","type":"gauge","value":1.23}`
		req := httptest.NewRequest("POST", "/update/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateMetricJSON(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.Metrics
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 1.23, *resp.Value)
	})

	t.Run("invalid content type", func(t *testing.T) {
		body := `{"id":"test","type":"gauge","value":1.23}`
		req := httptest.NewRequest("POST", "/update/", strings.NewReader(body))
		w := httptest.NewRecorder()

		handler.UpdateMetricJSON(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestGetMetricValueJSONHandler(t *testing.T) {
	repo := repository.NewMemoryRepository()
	handler := NewMetricHandler(service.NewMetricService(repo))
	repo.UpdateGauge(context.Background(), "test", 1.23)

	t.Run("success get gauge", func(t *testing.T) {
		body := `{"id":"test","type":"gauge"}`
		req := httptest.NewRequest("POST", "/value/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.GetMetricValueJSON(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp models.Metrics
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, 1.23, *resp.Value)
	})

	t.Run("metric not found", func(t *testing.T) {
		body := `{"id":"unknown","type":"gauge"}`
		req := httptest.NewRequest("POST", "/value/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.GetMetricValueJSON(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
