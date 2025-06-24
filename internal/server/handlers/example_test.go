package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"

	"go-metrics-server/internal/models"
	"go-metrics-server/internal/server/handlers"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/service"
)

func ExampleMetricHandler_UpdateMetric() {
	// Create test server
	repo := repository.NewMemoryRepository()
	handler := handlers.NewMetricHandler(service.NewMetricService(repo))

	// Example 1: Update gauge metric
	gaugeReq := httptest.NewRequest("POST", "/update/gauge/temperature/23.5", nil)
	w := httptest.NewRecorder()
	handler.UpdateMetric(w, gaugeReq)
	fmt.Printf("Gauge update status: %d\n", w.Code)

	// Example 2: Update counter metric
	counterReq := httptest.NewRequest("POST", "/update/counter/requests/1", nil)
	w = httptest.NewRecorder()
	handler.UpdateMetric(w, counterReq)
	fmt.Printf("Counter update status: %d\n", w.Code)

	// Output:
	// Gauge update status: 200
	// Counter update status: 200
}

func ExampleMetricHandler_UpdateMetricJSON() {
	repo := repository.NewMemoryRepository()
	handler := handlers.NewMetricHandler(service.NewMetricService(repo))

	// Prepare JSON request
	metric := models.Metrics{
		ID:    "cpu_usage",
		MType: "gauge",
		Value: ptrFloat64(75.3),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest("POST", "/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler.UpdateMetricJSON(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	var resp models.Metrics
	json.NewDecoder(w.Body).Decode(&resp)
	fmt.Printf("Updated value: %.1f\n", *resp.Value)

	// Output:
	// Status: 200
	// Updated value: 75.3
}

func ExampleMetricHandler_GetAllMetrics() {
	repo := repository.NewMemoryRepository()
	handler := handlers.NewMetricHandler(service.NewMetricService(repo))

	// Add some test data
	repo.UpdateGauge(context.Background(), "temperature", 23.5)
	repo.UpdateCounter(context.Background(), "requests", 42)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.GetAllMetrics(w, req)

	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))
	fmt.Printf("Body contains temperature: %v\n", bytes.Contains(w.Body.Bytes(), []byte("temperature")))

	// Output:
	// Status: 200
	// Content-Type: text/html
	// Body contains temperature: true
}

func ptrFloat64(f float64) *float64 {
	return &f
}
