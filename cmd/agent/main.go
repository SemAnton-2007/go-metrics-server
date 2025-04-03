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
	metricsCollector := metrics.NewMetrics()
	client := client.NewClient(cfg.ServerAddr)

	// Канал для синхронизации доступа к метрикам
	metricsChan := make(chan struct{}, 1)

	// Запускаем сбор метрик
	go func() {
		for range time.Tick(cfg.PollInterval) {
			metricsChan <- struct{}{} // Захватываем канал
			metricsCollector.Update()
			<-metricsChan // Освобождаем канал
		}
	}()

	// Запускаем отправку метрик
	for range time.Tick(cfg.ReportInterval) {
		metricsChan <- struct{}{} // Захватываем канал
		metricsSnapshot := metricsCollector.GetMetrics()
		<-metricsChan // Освобождаем канал

		if err := client.SendMetricsBatch(metricsSnapshot); err != nil {
			log.Printf("Failed to send metrics batch: %v", err)
		}
	}
}
