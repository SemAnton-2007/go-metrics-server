package handlers

import (
	"go-metrics-server/cmd/server/storage"
	"net/http"
	"strconv"
	"strings"
)

func UpdateMetricHandler(storage storage.MemStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Разбираем путь на части
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) != 5 {
			http.Error(w, "Invalid URL format", http.StatusNotFound)
			return
		}

		metricType := parts[2]
		metricName := parts[3]
		metricValue := parts[4]

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
