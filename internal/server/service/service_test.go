// internal/server/service/service_test.go
package service

import (
	"context"
	"testing"

	"go-metrics-server/internal/server/repository"
)

type mockRepo struct {
	repository.MetricRepository
}

func (m *mockRepo) UpdateGauge(ctx context.Context, name string, value float64) error {
	return nil
}

func TestMetricService(t *testing.T) {
	ctx := context.Background()
	repo := &mockRepo{}
	service := NewMetricService(repo)

	tests := []struct {
		name    string
		metric  string
		value   float64
		wantErr bool
	}{
		{"simple gauge", "test_gauge", 1.23, false},
		{"negative value", "negative", -1.23, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.UpdateGauge(ctx, tt.metric, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
