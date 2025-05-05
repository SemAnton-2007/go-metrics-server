package service

import (
	"context"
	"go-metrics-server/internal/models"
	"go-metrics-server/internal/server/repository"
)

type MetricService struct {
	repo repository.MetricRepository
}

func NewMetricService(repo repository.MetricRepository) *MetricService {
	return &MetricService{repo: repo}
}

func (s *MetricService) UpdateGauge(ctx context.Context, name string, value float64) error {
	return s.repo.UpdateGauge(ctx, name, value)
}

func (s *MetricService) UpdateCounter(ctx context.Context, name string, value int64) error {
	return s.repo.UpdateCounter(ctx, name, value)
}

func (s *MetricService) GetGauge(ctx context.Context, name string) (float64, error) {
	return s.repo.GetGauge(ctx, name)
}

func (s *MetricService) GetCounter(ctx context.Context, name string) (int64, error) {
	return s.repo.GetCounter(ctx, name)
}

func (s *MetricService) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetAllMetrics(ctx)
}

func (s *MetricService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	return s.repo.UpdateMetrics(ctx, metrics)
}

func (s *MetricService) SaveToFile(ctx context.Context, filename string) error {
	return s.repo.SaveToFile(ctx, filename)
}

func (s *MetricService) LoadFromFile(ctx context.Context, filename string) error {
	return s.repo.LoadFromFile(ctx, filename)
}
