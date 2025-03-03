package storage

import (
	"errors"
	"sync"
)

type MemStorage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
}

type memStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
}

func NewMemStorage() MemStorage {
	return &memStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *memStorage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
}

func (s *memStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
}

func (s *memStorage) GetGauge(name string) (float64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value, ok := s.gauges[name]; ok {
		return value, nil
	}
	return 0, errors.New("gauge not found")
}

func (s *memStorage) GetCounter(name string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value, ok := s.counters[name]; ok {
		return value, nil
	}
	return 0, errors.New("counter not found")
}
