package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_Update(t *testing.T) {
	m := NewMetrics()
	m.Update()

	assert.Equal(t, int64(1), m.PollCount)
	assert.True(t, m.RandomValue >= 0 && m.RandomValue <= 1)
}

func TestMetrics_GetMetrics(t *testing.T) {
	m := NewMetrics()
	m.Update()

	metrics := m.GetMetrics()
	assert.NotEmpty(t, metrics)
	assert.Contains(t, metrics, "PollCount")
	assert.Contains(t, metrics, "RandomValue")
}
