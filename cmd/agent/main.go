package main

import (
	"go-metrics-server/cmd/agent/client"
	"go-metrics-server/cmd/agent/config"
	"go-metrics-server/cmd/agent/metrics"
	"log"
	"sync"
	"time"
)

func main() {
	cfg := config.NewConfig()
	metricsCollector := metrics.NewMetrics()
	client := client.NewClient(cfg.ServerAddr, cfg.Key)

	// Каналы для обмена данными между горутинами
	metricsChan := make(chan map[string]interface{})
	done := make(chan struct{})
	var wg sync.WaitGroup

	// Worker pool для ограничения количества одновременных запросов
	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for metrics := range metricsChan {
				if err := client.SendMetricsBatch(metrics); err != nil {
					log.Printf("Failed to send metrics batch: %v", err)
				}
			}
		}()
	}

	// Запускаем сбор метрик в отдельной горутине
	go func() {
		for range time.Tick(cfg.PollInterval) {
			metricsCollector.Update()
		}
	}()

	// Запускаем отправку метрик в отдельной горутине
	go func() {
		for range time.Tick(cfg.ReportInterval) {
			metricsSnapshot := metricsCollector.GetMetrics()
			select {
			case metricsChan <- metricsSnapshot:
				// Метрики отправлены в worker pool
			default:
				log.Println("Rate limit exceeded, skipping metrics send")
			}
		}
	}()

	// Ожидаем сигнала завершения
	<-done

	// Завершаем работу
	close(metricsChan)
	wg.Wait()
}
