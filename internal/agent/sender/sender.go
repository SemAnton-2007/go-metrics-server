package sender

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
	initialRetryDelay = time.Second
	secondRetryDelay  = 3 * time.Second
	thirdRetryDelay   = 5 * time.Second
)

var retryableErrors = []error{
	errors.New("connection refused"),
	errors.New("connection reset by peer"),
	errors.New("i/o timeout"),
}

type Sender struct {
	ServerURL string
	Client    *http.Client
	Key       string
}

func New(serverURL, key string) *Sender {
	if !strings.HasPrefix(serverURL, httpScheme) && !strings.HasPrefix(serverURL, httpsScheme) {
		serverURL = httpScheme + serverURL
	}
	return &Sender{
		ServerURL: serverURL,
		Client:    &http.Client{Timeout: 10 * time.Second},
		Key:       key,
	}
}

func (s *Sender) SendMetric(metricType, name string, value interface{}) error {
	metric, err := s.createMetric(metricType, name, value)
	if err != nil {
		return err
	}
	return s.sendWithRetry("/update/", []models.Metrics{metric})
}

func (s *Sender) SendMetricsBatch(metrics map[string]interface{}) error {
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

	return s.sendRequest("/updates/", batch)
}

func (s *Sender) createMetric(metricType, name string, value interface{}) (models.Metrics, error) {
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

func (s *Sender) sendWithRetry(endpoint string, metrics []models.Metrics) error {
	var lastErr error
	delays := []time.Duration{initialRetryDelay, secondRetryDelay, thirdRetryDelay}

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(delays[attempt-1])
		}

		err := s.sendRequest(endpoint, metrics)
		if err == nil {
			return nil
		}

		lastErr = err
		if !s.isRetryableError(err) {
			break
		}
	}

	return fmt.Errorf("after %d attempts, last error: %w", maxRetries+1, lastErr)
}

func (s *Sender) isRetryableError(err error) bool {
	errStr := err.Error()
	for _, retryableErr := range retryableErrors {
		if strings.Contains(errStr, retryableErr.Error()) {
			return true
		}
	}
	return false
}

func (s *Sender) sendRequest(endpoint string, metrics []models.Metrics) error {
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
		fmt.Sprintf("%s%s", s.ServerURL, endpoint),
		&buf,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Accept-Encoding", "gzip")

	if s.Key != "" {
		h := hmac.New(sha256.New, []byte(s.Key))
		h.Write(jsonData)
		hash := hex.EncodeToString(h.Sum(nil))
		req.Header.Set("HashSHA256", hash)
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}
