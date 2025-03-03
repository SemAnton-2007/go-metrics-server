package main

import (
	"go-metrics-server/cmd/agent/client"
	"go-metrics-server/cmd/agent/config"
	"go-metrics-server/cmd/agent/metrics"
	"log"
	"time"
)

func main() {
	cfg := config.NewConfig()
	metrics := metrics.NewMetrics()
	client := client.NewClient(cfg.ServerAddr)

	// Запускаем сбор метрик
	go func() {
		for {
			metrics.Update()
			time.Sleep(cfg.PollInterval)
		}
	}()

	// Запускаем отправку метрик
	for {
		time.Sleep(cfg.ReportInterval)
		sendMetrics(metrics, client)
	}
}

// sendMetrics — отправляет метрики на сервер.
func sendMetrics(m *metrics.Metrics, c *client.Client) {
	metrics := m.GetMetrics()
	for name, value := range metrics {
		var metricType string
		switch value.(type) {
		case float64:
			metricType = "gauge"
		case int64:
			metricType = "counter"
		default:
			continue
		}

		if err := c.SendMetric(metricType, name, value); err != nil {
			log.Printf("Failed to send metric %s: %v", name, err)
		}
	}
}
