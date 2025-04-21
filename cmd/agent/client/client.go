package client

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go-metrics-server/internal/models"
	"net/http"
	"strings"
	"time"
)

const (
	httpScheme        = "http://"
	httpsScheme       = "https://"
	maxRetries        = 3
	initialRetryDelay = time.Second     // Первая задержка - 1s
	secondRetryDelay  = 3 * time.Second // Вторая задержка - 3s
	thirdRetryDelay   = 5 * time.Second // Третья задержка - 5s
)

var retryableErrors = []error{
	errors.New("connection refused"),
	errors.New("connection reset by peer"),
	errors.New("i/o timeout"),
}

type Client struct {
	ServerURL string
	Client    *http.Client
	Key       string // Ключ для подписи данных
}

func NewClient(serverURL, key string) *Client {
	if !strings.HasPrefix(serverURL, httpScheme) && !strings.HasPrefix(serverURL, httpsScheme) {
		serverURL = httpScheme + serverURL
	}
	return &Client{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
		Key:       key,
	}
}

// SendMetric - отправляет одну метрику (сохраняем старый функционал для совместимости)
func (c *Client) SendMetric(metricType, name string, value interface{}) error {
	metric, err := c.createMetric(metricType, name, value)
	if err != nil {
		return err
	}
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

func (c *Client) createMetric(metricType, name string, value interface{}) (models.Metrics, error) {
	metric := models.Metrics{
		ID:    name,
		MType: metricType,
	}

	switch metricType {
	case "gauge":
		if v, ok := value.(float64); ok {
			metric.Value = &v
		} else {
			return models.Metrics{}, fmt.Errorf("invalid value type for gauge metric, expected float64, got %T", value)
		}
	case "counter":
		if v, ok := value.(int64); ok {
			metric.Delta = &v
		} else {
			return models.Metrics{}, fmt.Errorf("invalid value type for counter metric, expected int64, got %T", value)
		}
	default:
		return models.Metrics{}, fmt.Errorf("unknown metric type: %s", metricType)
	}

	return metric, nil
}

func (c *Client) sendWithRetry(endpoint string, metrics []models.Metrics) error {
	var lastErr error
	delays := []time.Duration{initialRetryDelay, secondRetryDelay, thirdRetryDelay}

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

	// Добавляем подпись, если ключ установлен
	if c.Key != "" {
		h := hmac.New(sha256.New, []byte(c.Key))
		h.Write(jsonData)
		hash := hex.EncodeToString(h.Sum(nil))
		req.Header.Set("HashSHA256", hash)
	}

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
