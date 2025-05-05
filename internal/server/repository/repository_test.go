package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryRepository(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryRepository()

	// Тест 1: Обновление и получение gauge
	err := repo.UpdateGauge(ctx, "test", 123.45)
	assert.NoError(t, err)
	value, err := repo.GetGauge(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, 123.45, value)

	// Тест 2: Обновление и получение counter
	err = repo.UpdateCounter(ctx, "test", 10)
	assert.NoError(t, err)
	valueInt, err := repo.GetCounter(ctx, "test")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), valueInt)

	// Тест 3: Получение несуществующей метрики
	_, err = repo.GetGauge(ctx, "unknown")
	assert.Error(t, err)
	_, err = repo.GetCounter(ctx, "unknown")
	assert.Error(t, err)

	// Тест 4: Сохранение и загрузка метрик
	tmpFile := filepath.Join(os.TempDir(), "metrics_test.json")
	defer os.Remove(tmpFile)

	err = repo.UpdateGauge(ctx, "cpu", 75.5)
	assert.NoError(t, err)
	err = repo.UpdateCounter(ctx, "requests", 42)
	assert.NoError(t, err)
	err = repo.SaveToFile(ctx, tmpFile)
	assert.NoError(t, err)

	newRepo := NewMemoryRepository()
	err = newRepo.LoadFromFile(ctx, tmpFile)
	assert.NoError(t, err)

	val, err := newRepo.GetGauge(ctx, "cpu")
	assert.NoError(t, err)
	assert.Equal(t, 75.5, val)

	valInt, err := newRepo.GetCounter(ctx, "requests")
	assert.NoError(t, err)
	assert.Equal(t, int64(42), valInt)

	// Тест 5: Загрузка из несуществующего файла
	err = newRepo.LoadFromFile(ctx, "non_existent.json")
	assert.NoError(t, err) // Должно просто игнорироваться

	// Тест 6: Сохранение с пустым именем файла
	err = repo.SaveToFile(ctx, "")
	assert.NoError(t, err) // Должно просто игнорироваться
}
