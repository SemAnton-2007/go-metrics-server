package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// TestLoggerMiddleware проверяет, что middleware корректно логирует запросы и ответы.
func TestLoggerMiddleware(t *testing.T) {
	// Создаем буфер для записи логов
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	// Создаем тестовый обработчик, который возвращает статус 200 и тело "OK"
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Оборачиваем обработчик в middleware
	middleware := LoggerMiddleware(logger)
	wrappedHandler := middleware(handler)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	wrappedHandler.ServeHTTP(rec, req)

	// Проверяем, что статус ответа и тело корректны
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	// Проверяем, что лог содержит ожидаемые поля
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"uri":"/test"`)
	assert.Contains(t, logOutput, `"status":200`)
	assert.Contains(t, logOutput, `"size":2`)    // Размер тела ответа "OK" — 2 байта
	assert.Contains(t, logOutput, `"duration":`) // Поле duration должно присутствовать
}

// TestLoggerMiddlewareWithError проверяет, что middleware корректно логирует ошибки.
func TestLoggerMiddlewareWithError(t *testing.T) {
	// Создаем буфер для записи логов
	var buf bytes.Buffer
	logger := zerolog.New(&buf).With().Timestamp().Logger()

	// Создаем тестовый обработчик, который возвращает статус 404
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	})

	// Оборачиваем обработчик в middleware
	middleware := LoggerMiddleware(logger)
	wrappedHandler := middleware(handler)

	// Создаем тестовый запрос
	req := httptest.NewRequest("GET", "/not-found", nil)
	rec := httptest.NewRecorder()

	// Выполняем запрос
	wrappedHandler.ServeHTTP(rec, req)

	// Проверяем, что статус ответа и тело корректны
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Equal(t, "Not Found", rec.Body.String())

	// Проверяем, что лог содержит ожидаемые поля
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"uri":"/not-found"`)
	assert.Contains(t, logOutput, `"status":404`)
	assert.Contains(t, logOutput, `"size":9`)    // Размер тела ответа "Not Found" — 9 байт
	assert.Contains(t, logOutput, `"duration":`) // Поле duration должно присутствовать
}

// TestLoggingResponseWriter проверяет, что loggingResponseWriter корректно перехватывает статус и размер ответа.
func TestLoggingResponseWriter(t *testing.T) {
	// Создаем mock ResponseWriter
	rec := httptest.NewRecorder()

	// Оборачиваем его в loggingResponseWriter
	lw := &loggingResponseWriter{ResponseWriter: rec}

	// Устанавливаем статус и пишем тело ответа
	lw.WriteHeader(http.StatusOK)
	size, err := lw.Write([]byte("OK"))

	// Проверяем, что размер и ошибка корректны
	assert.Equal(t, 2, size)
	assert.NoError(t, err)

	// Проверяем, что статус и размер записаны
	assert.Equal(t, http.StatusOK, lw.statusCode)
	assert.Equal(t, 2, lw.size)
}
