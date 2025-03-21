package client

import (
	"bytes"
	"fmt"
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

// SendMetric — отправляет метрику на сервер.
func (c *Client) SendMetric(metricType, name string, value interface{}) error {
	url := fmt.Sprintf("%s/update/%s/%s/%v", c.ServerURL, metricType, name, value)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}
