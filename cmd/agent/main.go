package main

import (
	"go-metrics-server/internal/agent/config"
	"go-metrics-server/internal/agent/metrics"
	"go-metrics-server/internal/agent/sender"
	"log"
	"sync"
	"time"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	log.Printf("Build version: %s\n", buildVersion)
	log.Printf("Build date: %s\n", buildDate)
	log.Printf("Build commit: %s\n", buildCommit)

	RunAgent(config.NewConfig())
}

func RunAgent(cfg *config.Config) {
	metricsCollector := metrics.NewMetrics()
	sender := sender.New(cfg.ServerAddr, cfg.Key)

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
