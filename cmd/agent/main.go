package main

import (
	"go-metrics-server/internal/agent/config"
	"go-metrics-server/internal/agent/metrics"
	"go-metrics-server/internal/agent/sender"
	"log"
	"sync"
	"time"
)

func main() {
	cfg := config.NewConfig()
	metricsCollector := metrics.NewMetrics()
	sender := sender.New(cfg.ServerAddr, cfg.Key)

	metricsChan := make(chan map[string]interface{})
	done := make(chan struct{})
	var wg sync.WaitGroup

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for metrics := range metricsChan {
				if err := sender.SendMetricsBatch(metrics); err != nil {
					log.Printf("Failed to send metrics batch: %v", err)
				}
			}
		}()
	}

	go func() {
		for range time.Tick(cfg.PollInterval) {
			metricsCollector.Update()
		}
	}()

	go func() {
		for range time.Tick(cfg.ReportInterval) {
			metricsSnapshot := metricsCollector.GetMetrics()
			select {
			case metricsChan <- metricsSnapshot:
			default:
				log.Println("Rate limit exceeded, skipping metrics send")
			}
		}
	}()

	<-done

	close(metricsChan)
	wg.Wait()
}
