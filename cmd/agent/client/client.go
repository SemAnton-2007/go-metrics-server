package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"go-metrics-server/internal/models"
	"net/http"
	"strings"
	"time"
)

const (
	httpScheme  = "http://"
	httpsScheme = "https://"
)

// Client — HTTP-клиент для отправки метрик.
type Client struct {
	ServerURL string
	Client    *http.Client
}

// NewClient — конструктор для Client.
func NewClient(serverURL string) *Client {
	if !strings.HasPrefix(serverURL, httpScheme) && !strings.HasPrefix(serverURL, httpsScheme) {
		serverURL = httpScheme + serverURL
	}
	return &Client{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// SendMetric — отправляет метрику на сервер в формате JSON.
func (c *Client) SendMetric(metricType, name string, value interface{}) error {
	metric := c.createMetric(metricType, name, value)
	return c.sendMetricsBatch([]models.Metrics{metric})
}

// SendMetricsBatch - отправляет метрики батчем
func (c *Client) SendMetricsBatch(metrics map[string]interface{}) error {
	var batch []models.Metrics

	for name, value := range metrics {
		var metricType string
		switch value.(type) {
		case float64:
			metricType = "gauge"
		case int64:
			metricType = "counter"
		default:
			continue
		}
		batch = append(batch, c.createMetric(metricType, name, value))
	}

	if len(batch) == 0 {
		return nil // Не отправляем пустые батчи
	}

	return c.sendMetricsBatch(batch)
}

func (c *Client) createMetric(metricType, name string, value interface{}) models.Metrics {
	metric := models.Metrics{
		ID:    name,
		MType: metricType,
	}

	switch v := value.(type) {
	case float64:
		metric.Value = &v
	case int64:
		metric.Delta = &v
	}

	return metric
}

func (c *Client) sendMetricsBatch(metrics []models.Metrics) error {
	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(jsonData); err != nil {
		return fmt.Errorf("compression error: %w", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("compression close error: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/updates/", c.ServerURL),
		&buf,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
