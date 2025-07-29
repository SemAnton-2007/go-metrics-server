package main

import (
	"go-metrics-server/internal/agent/config"
	"go-metrics-server/internal/agent/metrics"
	"go-metrics-server/internal/agent/sender"
	"go-metrics-server/internal/buildinfo"
	"log"
	"sync"
	"time"
)

func main() {
	buildinfo.Print()

	RunAgent(config.NewConfig())
}

func RunAgent(cfg *config.Config) {
	metricsCollector := metrics.NewMetrics()
	sender := sender.New(cfg.ServerAddr, cfg.Key, cfg)

	metricsChan := make(chan map[string]interface{})
	var wg sync.WaitGroup

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range metricsChan {
				if err := sender.SendMetricsBatch(m); err != nil {
					log.Printf("Failed to send metrics batch: %v", err)
				}
			}
		}()
	}

	pollTicker := time.NewTicker(cfg.PollInterval)
	defer pollTicker.Stop()

	go func() {
		for range pollTicker.C {
			metricsCollector.Update()
		}
	}()

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()

	go func() {
		for range reportTicker.C {
			select {
			case metricsChan <- metricsCollector.GetMetrics():
			default:
				log.Println("Rate limit exceeded, skipping metrics send")
			}
		}
	}()

	<-make(chan struct{})
	close(metricsChan)
	wg.Wait()
}
