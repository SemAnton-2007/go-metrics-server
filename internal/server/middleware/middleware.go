package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// LoggerMiddleware создает middleware для логирования запросов и ответов.
func LoggerMiddleware(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Начало отсчета времени выполнения запроса
			start := time.Now()

			// Создаем обертку для ResponseWriter, чтобы перехватывать статус и размер ответа
			lw := &loggingResponseWriter{ResponseWriter: w}

			// Передаем управление следующему обработчику
			next.ServeHTTP(lw, r)

			// Логируем информацию о запросе и ответе
			logger.Info().
				Str("method", r.Method).
				Str("uri", r.RequestURI).
				Int("status", lw.statusCode).
				Int("size", lw.size).
				Dur("duration", time.Since(start)).
				Msg("request processed")
		})
	}
}

// loggingResponseWriter — обертка для http.ResponseWriter, чтобы перехватывать статус и размер ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

// WriteHeader перехватывает статус ответа.
func (lw *loggingResponseWriter) WriteHeader(statusCode int) {
	lw.statusCode = statusCode
	lw.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает размер ответа.
func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := lw.ResponseWriter.Write(b)
	lw.size += size
	return size, err
}
