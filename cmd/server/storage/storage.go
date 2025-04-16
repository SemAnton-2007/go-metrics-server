package storage

import (
	"encoding/json"
	"errors"
	"go-metrics-server/internal/models"
	"os"
	"sync"
)

// MemStorage — интерфейс для работы с хранилищем метрик.
type MemStorage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, error)
	GetCounter(name string) (int64, error)
	GetAllMetrics() (map[string]interface{}, error) // Измененная сигнатура
	SaveToFile(filename string) error
	LoadFromFile(filename string) error
	UpdateMetrics(metrics []models.Metrics) error
}

// memStorage — реализация MemStorage.
type memStorage struct {
	gauges   map[string]float64
	counters map[string]int64
	mu       sync.Mutex
}

// NewMemStorage — конструктор для memStorage.
func NewMemStorage() *memStorage {
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

func (s *memStorage) GetAllMetrics() (map[string]interface{}, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	metrics := make(map[string]interface{})
	for name, value := range s.gauges {
		metrics[name] = value
	}
	for name, value := range s.counters {
		metrics[name] = value
	}
	return metrics, nil
}

func (s *memStorage) SaveToFile(filename string) error {
	if filename == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data := struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}{
		Gauges:   s.gauges,
		Counters: s.counters,
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func (s *memStorage) LoadFromFile(filename string) error {
	if filename == "" {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Файл не существует - это не ошибка
		}
		return err
	}
	defer file.Close()

	var data struct {
		Gauges   map[string]float64 `json:"gauges"`
		Counters map[string]int64   `json:"counters"`
	}

	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	s.gauges = data.Gauges
	s.counters = data.Counters

	return nil
}

func (s *memStorage) UpdateMetrics(metrics []models.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value != nil {
				s.gauges[metric.ID] = *metric.Value
			}
		case "counter":
			if metric.Delta != nil {
				s.counters[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}
