package client

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-server/internal/models"
	"net/http"
	"strings"
	"time"
)

const (
	httpScheme  = "http://"
	httpsScheme = "https://"
	maxRetries  = 3
	retryDelay  = time.Second
)

var retryableErrors = []error{
	errors.New("connection refused"),
	errors.New("connection reset by peer"),
	errors.New("i/o timeout"),
}

type Client struct {
	ServerURL string
	Client    *http.Client
}

func NewClient(serverURL string) *Client {
	if !strings.HasPrefix(serverURL, httpScheme) && !strings.HasPrefix(serverURL, httpsScheme) {
		serverURL = httpScheme + serverURL
	}
	return &Client{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// SendMetric - отправляет одну метрику (сохраняем старый функционал для совместимости)
func (c *Client) SendMetric(metricType, name string, value interface{}) error {
	metric := c.createMetric(metricType, name, value)
	return c.sendWithRetry("/update/", []models.Metrics{metric})
}

// SendMetricsBatch - отправляет метрики батчем (новый функционал)
func (c *Client) SendMetricsBatch(metrics map[string]interface{}) error {
	var batch []models.Metrics

	for name, value := range metrics {
		var metric models.Metrics
		metric.ID = name

		switch v := value.(type) {
		case float64:
			metric.MType = "gauge"
			metric.Value = &v
		case int64:
			metric.MType = "counter"
			metric.Delta = &v
		default:
			continue
		}
		batch = append(batch, metric)
	}

	if len(batch) == 0 {
		return nil
	}

	return c.sendRequest("/updates/", batch)
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

func (c *Client) sendWithRetry(endpoint string, metrics []models.Metrics) error {
	var lastErr error
	delays := []time.Duration{retryDelay, 3 * retryDelay, 5 * retryDelay}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := c.sendRequest(endpoint, metrics)
		if err == nil {
			return nil
		}

		lastErr = err
		if !c.isRetryableError(err) {
			break
		}
	}

	return fmt.Errorf("after %d attempts, last error: %w", maxRetries+1, lastErr)
}

func (c *Client) isRetryableError(err error) bool {
	errStr := err.Error()
	for _, retryableErr := range retryableErrors {
		if strings.Contains(errStr, retryableErr.Error()) {
			return true
		}
	}
	return false
}

func (c *Client) sendRequest(endpoint string, metrics []models.Metrics) error {
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
		fmt.Sprintf("%s%s", c.ServerURL, endpoint),
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
