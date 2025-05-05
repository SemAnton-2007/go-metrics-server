package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-metrics-server/internal/models"
	"go-metrics-server/internal/server/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type MetricHandler struct {
	service *service.MetricService
}

func NewMetricHandler(service *service.MetricService) *MetricHandler {
	return &MetricHandler{service: service}
}

func (h *MetricHandler) UpdateMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	switch metricType {
	case "gauge":
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			http.Error(w, "Invalid gauge value", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateGauge(ctx, metricName, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case "counter":
		value, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			http.Error(w, "Invalid counter value", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounter(ctx, metricName, value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h *MetricHandler) GetMetricValue(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	var value interface{}
	var err error

	ctx := r.Context()
	switch metricType {
	case "gauge":
		value, err = h.service.GetGauge(ctx, metricName)
	case "counter":
		value, err = h.service.GetCounter(ctx, metricName)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	if err != nil {
		http.Error(w, "Metric not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("%v", value)))
}

func (h *MetricHandler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metrics, err := h.service.GetAllMetrics(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	var buf bytes.Buffer
	buf.WriteString("<h1>Metrics</h1>\n<ul>\n")
	for name, value := range metrics {
		buf.WriteString(fmt.Sprintf("<li>%s: %v</li>", name, value))
	}
	buf.WriteString("</ul>")

	w.Write(buf.Bytes())
}

func (h *MetricHandler) UpdateMetricJSON(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			http.Error(w, "Value is required for gauge", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateGauge(ctx, metric.ID, *metric.Value); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		value, _ := h.service.GetGauge(ctx, metric.ID)
		metric.Value = &value

	case "counter":
		if metric.Delta == nil {
			http.Error(w, "Delta is required for counter", http.StatusBadRequest)
			return
		}
		if err := h.service.UpdateCounter(ctx, metric.ID, *metric.Delta); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		value, _ := h.service.GetCounter(ctx, metric.ID)
		metric.Delta = &value

	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Del("Content-Length")

	if err := json.NewEncoder(w).Encode(metric); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *MetricHandler) GetMetricValueJSON(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var metric models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metric); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if metric.ID == "" {
		http.Error(w, "Metric name is required", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	var err error
	switch metric.MType {
	case "gauge":
		value, e := h.service.GetGauge(ctx, metric.ID)
		err = e
		metric.Value = &value
	case "counter":
		value, e := h.service.GetCounter(ctx, metric.ID)
		err = e
		metric.Delta = &value
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(metric)
}

func (h *MetricHandler) BatchUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusBadRequest)
		return
	}

	var metrics []models.Metrics
	if err := json.NewDecoder(r.Body).Decode(&metrics); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		http.Error(w, "Empty metrics batch", http.StatusBadRequest)
		return
	}

	for _, metric := range metrics {
		if metric.ID == "" {
			http.Error(w, "Metric name is required", http.StatusBadRequest)
			return
		}
	}

	ctx := r.Context()
	if err := h.service.UpdateMetrics(ctx, metrics); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}
