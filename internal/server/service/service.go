// Package service provides business logic layer for metrics operations.
// It acts as an intermediary between HTTP handlers and repository layer,
// ensuring proper separation of concerns.
package service

import (
	"context"
	"go-metrics-server/internal/models"
	"go-metrics-server/internal/server/repository"
)

// MetricService provides business logic operations for metrics.
// It validates input, handles transactions and coordinates repository operations.
type MetricService struct {
	repo repository.MetricRepository
}

// NewMetricService creates a new MetricService instance with the given repository.
// The repository can be either in-memory or PostgreSQL implementation.
func NewMetricService(repo repository.MetricRepository) *MetricService {
	return &MetricService{repo: repo}
}

// UpdateGauge updates a gauge metric with the given name and value.
// Gauge metrics represent values that can go up and down, like temperature or memory usage.
// Returns error if the update fails.
func (s *MetricService) UpdateGauge(ctx context.Context, name string, value float64) error {
	return s.repo.UpdateGauge(ctx, name, value)
}

// UpdateCounter updates a counter metric with the given name and delta value.
// Counter metrics only increase over time, like request counts.
// The delta value is added to the existing counter value.
// Returns error if the update fails.
func (s *MetricService) UpdateCounter(ctx context.Context, name string, value int64) error {
	return s.repo.UpdateCounter(ctx, name, value)
}

// GetGauge retrieves the current value of a gauge metric by name.
// Returns the metric value and nil if found, or 0 and error if not found.
func (s *MetricService) GetGauge(ctx context.Context, name string) (float64, error) {
	return s.repo.GetGauge(ctx, name)
}

// GetCounter retrieves the current value of a counter metric by name.
// Returns the metric value and nil if found, or 0 and error if not found.
func (s *MetricService) GetCounter(ctx context.Context, name string) (int64, error) {
	return s.repo.GetCounter(ctx, name)
}

// GetAllMetrics retrieves all stored metrics as a map of names to values.
// The map contains both gauge and counter metrics mixed together.
// Returns the metrics map and nil if successful, or nil and error if failed.
func (s *MetricService) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	return s.repo.GetAllMetrics(ctx)
}

// UpdateMetrics performs a batch update of multiple metrics in a single transaction.
// The metrics slice can contain both gauge and counter metrics.
// For counters, the delta values are added to existing values.
// Returns nil if all updates succeeded, or error if any update failed.
func (s *MetricService) UpdateMetrics(ctx context.Context, metrics []models.Metrics) error {
	return s.repo.UpdateMetrics(ctx, metrics)
}

// SaveToFile saves the current metrics to a JSON file.
// If filename is empty, the operation is skipped.
// Returns nil if successful, or error if file operations failed.
func (s *MetricService) SaveToFile(ctx context.Context, filename string) error {
	return s.repo.SaveToFile(ctx, filename)
}

// LoadFromFile loads metrics from a JSON file.
// If filename is empty or file doesn't exist, the operation is skipped.
// Existing metrics are overwritten by values from the file.
// Returns nil if successful, or error if file operations failed.
func (s *MetricService) LoadFromFile(ctx context.Context, filename string) error {
	return s.repo.LoadFromFile(ctx, filename)
}
