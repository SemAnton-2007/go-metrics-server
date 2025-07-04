package repository

import (
	"context"
	"testing"
)

func BenchmarkMemoryRepository_UpdateGauge(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.UpdateGauge(ctx, "test", float64(i))
	}
}

func BenchmarkMemoryRepository_UpdateCounter(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = repo.UpdateCounter(ctx, "test", int64(i))
	}
}

func BenchmarkMemoryRepository_GetAllMetrics(b *testing.B) {
	repo := NewMemoryRepository()
	ctx := context.Background()

	// Заполняем данными
	for i := 0; i < 1000; i++ {
		_ = repo.UpdateGauge(ctx, "test", float64(i))
		_ = repo.UpdateCounter(ctx, "test", int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetAllMetrics(ctx)
	}
}
