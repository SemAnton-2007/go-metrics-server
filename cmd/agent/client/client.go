package client

import (
	"bytes"
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
	// Добавляем схему "http://", если она отсутствует
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
	var metric models.Metrics
	metric.ID = name
	metric.MType = metricType

	switch v := value.(type) {
	case float64:
		metric.Value = &v
	case int64:
		metric.Delta = &v
	default:
		return fmt.Errorf("unsupported value type")
	}

	jsonData, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/update/", c.ServerURL),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
