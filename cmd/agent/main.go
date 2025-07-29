package main

import (
	"context"
	"go-metrics-server/internal/agent/config"
	"go-metrics-server/internal/agent/metrics"
	"go-metrics-server/internal/agent/sender"
	"go-metrics-server/internal/buildinfo"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v. Shutting down...", sig)
		cancel()
	}()

	for i := 0; i < cfg.RateLimit; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case m, ok := <-metricsChan:
					if !ok {
						return
					}
					if err := sender.SendMetricsBatch(m); err != nil {
						log.Printf("Failed to send metrics batch: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	pollTicker := time.NewTicker(cfg.PollInterval)
	defer pollTicker.Stop()

	go func() {
		for {
			select {
			case <-pollTicker.C:
				metricsCollector.Update()
			case <-ctx.Done():
				return
			}
		}
	}()

	reportTicker := time.NewTicker(cfg.ReportInterval)
	defer reportTicker.Stop()

	go func() {
		for {
			select {
			case <-reportTicker.C:
				select {
				case metricsChan <- metricsCollector.GetMetrics():
				default:
					log.Println("Rate limit exceeded, skipping metrics send")
				}
			case <-ctx.Done():
				log.Println("Sending final metrics before shutdown")
				metricsChan <- metricsCollector.GetMetrics()
				close(metricsChan)
				return
			}
		}
	}()

	wg.Wait()
	log.Println("Agent shutdown complete")
}
