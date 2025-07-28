// internal/server/service/service_test.go
package service

import (
	"context"
	"errors"
	"testing"

	"go-metrics-server/internal/models"

	"github.com/stretchr/testify/assert"
)

type mockRepo struct {
	updateGaugeErr   error
	updateCounterErr error
	getGaugeErr      error
	getCounterErr    error
	updateMetricsErr error
	saveToFileErr    error
	loadFromFileErr  error
}

func (m *mockRepo) UpdateGauge(ctx context.Context, name string, value float64) error {
	return m.updateGaugeErr
}

func (m *mockRepo) UpdateCounter(ctx context.Context, name string, value int64) error {
	return m.updateCounterErr
}

func (m *mockRepo) GetGauge(ctx context.Context, name string) (float64, error) {
	return 0, m.getGaugeErr
}

func (m *mockRepo) GetCounter(ctx context.Context, name string) (int64, error) {
	return 0, m.getCounterErr
}

func (m *mockRepo) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockRepo) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	return m.updateMetricsErr
}

func (m *mockRepo) SaveToFile(ctx context.Context, filename string) error {
	return m.saveToFileErr
}

func (m *mockRepo) LoadFromFile(ctx context.Context, filename string) error {
	return m.loadFromFileErr
}

func TestMetricService(t *testing.T) {
	t.Run("UpdateGauge success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		err := service.UpdateGauge(context.Background(), "test", 1.23)
		assert.NoError(t, err)
	})

	t.Run("UpdateGauge error", func(t *testing.T) {
		repo := &mockRepo{updateGaugeErr: errors.New("error")}
		service := NewMetricService(repo)
		err := service.UpdateGauge(context.Background(), "test", 1.23)
		assert.Error(t, err)
	})

	t.Run("UpdateCounter success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		err := service.UpdateCounter(context.Background(), "test", 10)
		assert.NoError(t, err)
	})

	t.Run("UpdateCounter error", func(t *testing.T) {
		repo := &mockRepo{updateCounterErr: errors.New("error")}
		service := NewMetricService(repo)
		err := service.UpdateCounter(context.Background(), "test", 10)
		assert.Error(t, err)
	})

	t.Run("GetGauge success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		_, err := service.GetGauge(context.Background(), "test")
		assert.NoError(t, err)
	})

	t.Run("GetGauge error", func(t *testing.T) {
		repo := &mockRepo{getGaugeErr: errors.New("error")}
		service := NewMetricService(repo)
		_, err := service.GetGauge(context.Background(), "test")
		assert.Error(t, err)
	})

	t.Run("GetCounter success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		_, err := service.GetCounter(context.Background(), "test")
		assert.NoError(t, err)
	})

	t.Run("GetCounter error", func(t *testing.T) {
		repo := &mockRepo{getCounterErr: errors.New("error")}
		service := NewMetricService(repo)
		_, err := service.GetCounter(context.Background(), "test")
		assert.Error(t, err)
	})

	t.Run("UpdateMetrics success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		err := service.UpdateMetrics(context.Background(), []models.Metrics{})
		assert.NoError(t, err)
	})

	t.Run("UpdateMetrics error", func(t *testing.T) {
		repo := &mockRepo{updateMetricsErr: errors.New("error")}
		service := NewMetricService(repo)
		err := service.UpdateMetrics(context.Background(), []models.Metrics{})
		assert.Error(t, err)
	})

	t.Run("SaveToFile success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		err := service.SaveToFile(context.Background(), "test")
		assert.NoError(t, err)
	})

	t.Run("SaveToFile error", func(t *testing.T) {
		repo := &mockRepo{saveToFileErr: errors.New("error")}
		service := NewMetricService(repo)
		err := service.SaveToFile(context.Background(), "test")
		assert.Error(t, err)
	})

	t.Run("LoadFromFile success", func(t *testing.T) {
		repo := &mockRepo{}
		service := NewMetricService(repo)
		err := service.LoadFromFile(context.Background(), "test")
		assert.NoError(t, err)
	})

	t.Run("LoadFromFile error", func(t *testing.T) {
		repo := &mockRepo{loadFromFileErr: errors.New("error")}
		service := NewMetricService(repo)
		err := service.LoadFromFile(context.Background(), "test")
		assert.Error(t, err)
	})
}
