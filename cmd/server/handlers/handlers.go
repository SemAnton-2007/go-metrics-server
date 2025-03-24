package handlers

import (
	"encoding/json"
	"fmt"
	"go-metrics-server/cmd/server/storage"
	"go-metrics-server/internal/models"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// UpdateMetricHandler — обработчик для обновления метрик.
func UpdateMetricHandler(storage storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")
		metricValue := chi.URLParam(r, "value")

		if metricName == "" {
			http.Error(w, "Metric name is required", http.StatusNotFound)
			return
		}

		switch metricType {
		case "gauge":
			value, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Invalid gauge value", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metricName, value)
		case "counter":
			value, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Invalid counter value", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metricName, value)
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

// GetMetricValueHandler — обработчик для получения значения метрики.
func GetMetricValueHandler(storage storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metricType := chi.URLParam(r, "type")
		metricName := chi.URLParam(r, "name")

		var value interface{}
		var err error

		switch metricType {
		case "gauge":
			value, err = storage.GetGauge(metricName)
		case "counter":
			value, err = storage.GetCounter(metricName)
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
}

// GetAllMetricsHandler — обработчик для получения всех метрик в HTML
func GetAllMetricsHandler(storage storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics := storage.GetAllMetrics()

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "<h1>Metrics</h1>")
		fmt.Fprintln(w, "<ul>")
		for name, value := range metrics {
			fmt.Fprintf(w, "<li>%s: %v</li>", name, value)
		}
		fmt.Fprintln(w, "</ul>")
	}
}

// UpdateMetricJSONHandler — обработчик для обновления метрик через JSON
func UpdateMetricJSONHandler(storage storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		switch metric.MType {
		case "gauge":
			if metric.Value == nil {
				http.Error(w, "Value is required for gauge", http.StatusBadRequest)
				return
			}
			storage.UpdateGauge(metric.ID, *metric.Value)
			value, _ := storage.GetGauge(metric.ID)
			metric.Value = &value
		case "counter":
			if metric.Delta == nil {
				http.Error(w, "Delta is required for counter", http.StatusBadRequest)
				return
			}
			storage.UpdateCounter(metric.ID, *metric.Delta)
			value, _ := storage.GetCounter(metric.ID)
			metric.Delta = &value
		default:
			http.Error(w, "Invalid metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metric)
	}
}

// GetMetricValueJSONHandler — обработчик для получения значения метрики через JSON
func GetMetricValueJSONHandler(storage storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		var err error
		switch metric.MType {
		case "gauge":
			value, e := storage.GetGauge(metric.ID)
			err = e
			metric.Value = &value
		case "counter":
			value, e := storage.GetCounter(metric.ID)
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
}
