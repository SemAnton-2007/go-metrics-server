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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	setupSignalHandler(cancel)

	wg := startWorkers(ctx, sender, metricsChan, cfg.RateLimit)

	pollTicker := startPollingTicker(ctx, metricsCollector, cfg.PollInterval)
	defer pollTicker.Stop()

	reportTicker := startReportingTicker(ctx, metricsCollector, metricsChan, cfg.ReportInterval)
	defer reportTicker.Stop()

	wg.Wait()
	log.Println("Agent shutdown complete")
}

func setupSignalHandler(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v. Shutting down...", sig)
		cancel()
	}()
}

func startWorkers(ctx context.Context, sender *sender.Sender, metricsChan chan map[string]interface{}, rateLimit int) *sync.WaitGroup {
	var wg sync.WaitGroup

	for i := 0; i < rateLimit; i++ {
		wg.Add(1)
		go worker(ctx, &wg, sender, metricsChan)
	}

	return &wg
}

func worker(ctx context.Context, wg *sync.WaitGroup, sender *sender.Sender, metricsChan chan map[string]interface{}) {
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
}

func startPollingTicker(ctx context.Context, metricsCollector *metrics.Metrics, interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				metricsCollector.Update()
			case <-ctx.Done():
				return
			}
		}
	}()

	return ticker
}

func startReportingTicker(ctx context.Context, metricsCollector *metrics.Metrics, metricsChan chan map[string]interface{}, interval time.Duration) *time.Ticker {
	ticker := time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-ticker.C:
				sendMetrics(metricsCollector, metricsChan)
			case <-ctx.Done():
				log.Println("Sending final metrics before shutdown")
				sendMetrics(metricsCollector, metricsChan)
				close(metricsChan)
				return
			}
		}
	}()

	return ticker
}

func sendMetrics(metricsCollector *metrics.Metrics, metricsChan chan map[string]interface{}) {
	select {
	case metricsChan <- metricsCollector.GetMetrics():
	default:
		log.Println("Rate limit exceeded, skipping metrics send")
	}
}
