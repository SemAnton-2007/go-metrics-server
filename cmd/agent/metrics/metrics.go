package metrics

import (
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Metrics struct {
	PollCount   int64
	RandomValue float64
	Runtime     runtime.MemStats
	mu          sync.Mutex
	lastPoll    time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		lastPoll: time.Now(),
	}
}

func (m *Metrics) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.PollCount++

	m.RandomValue = rand.Float64()

	runtime.ReadMemStats(&m.Runtime)

	m.lastPoll = time.Now()
}

func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Создаем копию метрик для безопасного доступа
	metrics := map[string]interface{}{
		"PollCount":   m.PollCount,
		"RandomValue": m.RandomValue,
	}

	// Добавляем runtime метрики
	metrics["Alloc"] = float64(m.Runtime.Alloc)
	metrics["BuckHashSys"] = float64(m.Runtime.BuckHashSys)
	metrics["Frees"] = float64(m.Runtime.Frees)
	metrics["GCCPUFraction"] = m.Runtime.GCCPUFraction
	metrics["GCSys"] = float64(m.Runtime.GCSys)
	metrics["HeapAlloc"] = float64(m.Runtime.HeapAlloc)
	metrics["HeapIdle"] = float64(m.Runtime.HeapIdle)
	metrics["HeapInuse"] = float64(m.Runtime.HeapInuse)
	metrics["HeapObjects"] = float64(m.Runtime.HeapObjects)
	metrics["HeapReleased"] = float64(m.Runtime.HeapReleased)
	metrics["HeapSys"] = float64(m.Runtime.HeapSys)
	metrics["LastGC"] = float64(m.Runtime.LastGC)
	metrics["Lookups"] = float64(m.Runtime.Lookups)
	metrics["MCacheInuse"] = float64(m.Runtime.MCacheInuse)
	metrics["MCacheSys"] = float64(m.Runtime.MCacheSys)
	metrics["MSpanInuse"] = float64(m.Runtime.MSpanInuse)
	metrics["MSpanSys"] = float64(m.Runtime.MSpanSys)
	metrics["Mallocs"] = float64(m.Runtime.Mallocs)
	metrics["NextGC"] = float64(m.Runtime.NextGC)
	metrics["NumForcedGC"] = float64(m.Runtime.NumForcedGC)
	metrics["NumGC"] = float64(m.Runtime.NumGC)
	metrics["OtherSys"] = float64(m.Runtime.OtherSys)
	metrics["PauseTotalNs"] = float64(m.Runtime.PauseTotalNs)
	metrics["StackInuse"] = float64(m.Runtime.StackInuse)
	metrics["StackSys"] = float64(m.Runtime.StackSys)
	metrics["Sys"] = float64(m.Runtime.Sys)
	metrics["TotalAlloc"] = float64(m.Runtime.TotalAlloc)

	return metrics
}
