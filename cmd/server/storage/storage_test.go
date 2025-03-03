package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage(t *testing.T) {
	storage := NewMemStorage()

	// Тест 1: Обновление и получение gauge
	storage.UpdateGauge("test", 123.45)
	value, err := storage.GetGauge("test")
	assert.NoError(t, err)
	assert.Equal(t, 123.45, value)

	// Тест 2: Обновление и получение counter
	storage.UpdateCounter("test", 10)
	valueInt, err := storage.GetCounter("test")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), valueInt)

	// Тест 3: Получение несуществующей метрики
	_, err = storage.GetGauge("unknown")
	assert.Error(t, err)
	_, err = storage.GetCounter("unknown")
	assert.Error(t, err)
}
