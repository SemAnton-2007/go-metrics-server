package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPinger реализует интерфейс для тестирования PingHandler
type mockPinger struct {
	pingErr error
}

func (m *mockPinger) Ping(ctx context.Context) error {
	return m.pingErr
}

func TestPingHandler(t *testing.T) {
	tests := []struct {
		name       string
		pingErr    error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "successful ping",
			pingErr:    nil,
			wantStatus: http.StatusOK,
			wantBody:   "OK",
		},
		{
			name:       "failed ping",
			pingErr:    assert.AnError,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "Database connection failed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем mock DB
			db := &mockPinger{pingErr: tt.pingErr}

			// Создаем запрос
			req, err := http.NewRequest("GET", "/ping", nil)
			require.NoError(t, err)

			// Создаем ResponseRecorder для записи ответа
			rr := httptest.NewRecorder()

			// Вызываем обработчик
			handler := PingHandler(db)
			handler.ServeHTTP(rr, req)

			// Проверяем статус код
			assert.Equal(t, tt.wantStatus, rr.Code)

			// Проверяем тело ответа
			assert.Equal(t, tt.wantBody, rr.Body.String())
		})
	}
}

func TestPingHandler_ContextCancel(t *testing.T) {
	// Создаем mock DB с долгим ping
	db := &mockPinger{pingErr: context.Canceled}

	// Создаем запрос с отмененным контекстом
	req, err := http.NewRequest("GET", "/ping", nil)
	require.NoError(t, err)
	ctx, cancel := context.WithCancel(req.Context())
	cancel()
	req = req.WithContext(ctx)

	// Создаем ResponseRecorder
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler := PingHandler(db)
	handler.ServeHTTP(rr, req)

	// Проверяем что получили 500 при отмене контекста
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
