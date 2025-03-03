package metrics

import (
	"math/rand"
	"runtime"
	"sync"
)

type Metrics struct {
	PollCount   int64
	RandomValue float64
	Runtime     runtime.MemStats
	mu          sync.Mutex
}

func NewMetrics() *Metrics {
	return &Metrics{}
}

func (m *Metrics) Update() {
	m.mu.Lock()
	defer m.mu.Unlock()

	runtime.ReadMemStats(&m.Runtime)

	m.PollCount++

	m.RandomValue = rand.Float64()
}

func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return map[string]interface{}{
		"Alloc":         float64(m.Runtime.Alloc),
		"BuckHashSys":   float64(m.Runtime.BuckHashSys),
		"Frees":         float64(m.Runtime.Frees),
		"GCCPUFraction": m.Runtime.GCCPUFraction,
		"GCSys":         float64(m.Runtime.GCSys),
		"HeapAlloc":     float64(m.Runtime.HeapAlloc),
		"HeapIdle":      float64(m.Runtime.HeapIdle),
		"HeapInuse":     float64(m.Runtime.HeapInuse),
		"HeapObjects":   float64(m.Runtime.HeapObjects),
		"HeapReleased":  float64(m.Runtime.HeapReleased),
		"HeapSys":       float64(m.Runtime.HeapSys),
		"LastGC":        float64(m.Runtime.LastGC),
		"Lookups":       float64(m.Runtime.Lookups),
		"MCacheInuse":   float64(m.Runtime.MCacheInuse),
		"MCacheSys":     float64(m.Runtime.MCacheSys),
		"MSpanInuse":    float64(m.Runtime.MSpanInuse),
		"MSpanSys":      float64(m.Runtime.MSpanSys),
		"Mallocs":       float64(m.Runtime.Mallocs),
		"NextGC":        float64(m.Runtime.NextGC),
		"NumForcedGC":   float64(m.Runtime.NumForcedGC),
		"NumGC":         float64(m.Runtime.NumGC),
		"OtherSys":      float64(m.Runtime.OtherSys),
		"PauseTotalNs":  float64(m.Runtime.PauseTotalNs),
		"StackInuse":    float64(m.Runtime.StackInuse),
		"StackSys":      float64(m.Runtime.StackSys),
		"Sys":           float64(m.Runtime.Sys),
		"TotalAlloc":    float64(m.Runtime.TotalAlloc),
		"PollCount":     m.PollCount,
		"RandomValue":   m.RandomValue,
	}
}
