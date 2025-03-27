package webservers

import (
	"compress/gzip"
	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/database"
	"go-metrics-server/cmd/server/handlers"
	"go-metrics-server/cmd/server/middleware"
	"go-metrics-server/cmd/server/storage"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func NewServer(cfg *config.Config, storage storage.MemStorage, db *database.DB) *http.Server {
	r := chi.NewRouter()

	// Инициализация логгера
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()

	// Добавляем middleware для логирования
	r.Use(middleware.LoggerMiddleware(logger))

	r.Use(gzipMiddleware)

	r.Use(jsonContentTypeMiddleware(storage))

	// Регистрируем обработчики
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricHandler(storage))
	r.Get("/value/{type}/{name}", handlers.GetMetricValueHandler(storage))
	r.Get("/", handlers.GetAllMetricsHandler(storage))

	// Новые JSON эндпоинты
	r.Post("/update/", handlers.UpdateMetricJSONHandler(storage))
	r.Post("/value/", handlers.GetMetricValueJSONHandler(storage))

	// Новый batch endpoint
	r.Post("/updates/", handlers.BatchUpdateHandler(storage))

	// Новый batch endpoint
	r.Post("/updates/", handlers.BatchUpdateHandler(storage))

	// Добавляем обработчик для проверки БД, если БД подключена
	if db != nil {
		r.Get("/ping", handlers.PingHandler(db))
	}

	return &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		contentEncoding := r.Header.Get("Content-Encoding")
		isGzipped := strings.Contains(contentEncoding, "gzip")

		// Распаковка входящего запроса
		if isGzipped {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		// Сжатие исходящего ответа
		if supportsGzip && isCompressibleContent(r) {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()

			// Специальная обработка для HTML
			if r.Header.Get("Accept") == "text/html" {
				w.Header().Set("Content-Type", "text/html")
			}

			next.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isCompressibleContent(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		r.Header.Get("Accept") == "text/html" // Добавляем проверку для Accept: text/html
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func jsonContentTypeMiddleware(storage storage.MemStorage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
				switch r.URL.Path {
				case "/update/":
					handlers.UpdateMetricJSONHandler(storage)(w, r)
					return
				case "/value/":
					handlers.GetMetricValueJSONHandler(storage)(w, r)
					return
				case "/updates/":
					handlers.BatchUpdateHandler(storage)(w, r)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
