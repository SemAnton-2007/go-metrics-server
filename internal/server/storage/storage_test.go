// cmd/server/storage/storage_test.go
package storage

import (
	"os"
	"path/filepath"
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

	// Новые тесты для сохранения/загрузки
	// Тест 4: Сохранение и загрузка метрик
	tmpFile := filepath.Join(os.TempDir(), "metrics_test.json")
	defer os.Remove(tmpFile)

	storage.UpdateGauge("cpu", 75.5)
	storage.UpdateCounter("requests", 42)
	err = storage.SaveToFile(tmpFile)
	assert.NoError(t, err)

	newStorage := NewMemStorage()
	err = newStorage.LoadFromFile(tmpFile)
	assert.NoError(t, err)

	val, err := newStorage.GetGauge("cpu")
	assert.NoError(t, err)
	assert.Equal(t, 75.5, val)

	valInt, err := newStorage.GetCounter("requests")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), valInt)

	// Тест 5: Загрузка из несуществующего файла
	err = newStorage.LoadFromFile("non_existent.json")
	assert.NoError(t, err) // Должно просто игнорироваться

	// Тест 6: Сохранение с пустым именем файла
	err = storage.SaveToFile("")
	assert.NoError(t, err) // Должно просто игнорироваться
}
