package handlers

import (
	"fmt"
	"go-metrics-server/cmd/server/storage"
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

// GetAllMetricsHandler — обработчик для получения всех метрик.
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
