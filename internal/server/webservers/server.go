package webservers

import (
	"compress/gzip"
	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	handler "go-metrics-server/internal/server/handlers"
	"go-metrics-server/internal/server/middleware"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/service"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func NewServer(cfg *config.Config, repo repository.MetricRepository, db *database.DB) *http.Server {
	r := chi.NewRouter()

	logger := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	r.Use(middleware.LoggerMiddleware(logger))
	r.Use(gzipMiddleware)
	r.Use(jsonContentTypeMiddleware)

	metricService := service.NewMetricService(repo)
	metricHandler := handler.NewMetricHandler(metricService)

	r.Post("/update/{type}/{name}/{value}", metricHandler.UpdateMetric)
	r.Get("/value/{type}/{name}", metricHandler.GetMetricValue)
	r.Get("/", metricHandler.GetAllMetrics)
	r.Post("/update/", metricHandler.UpdateMetricJSON)
	r.Post("/value/", metricHandler.GetMetricValueJSON)
	r.Post("/updates/", metricHandler.BatchUpdate)

	if db != nil {
		pingHandler := handler.NewPingHandler(db)
		r.Get("/ping", pingHandler.Ping)
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

		if isGzipped {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		if supportsGzip && isCompressibleContent(r) {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Transfer-Encoding", "chunked")
			w.Header().Del("Content-Length")

			gz := gzip.NewWriter(w)
			defer gz.Close()

			if r.Header.Get("Accept") == "text/html" {
				w.Header().Set("Content-Type", "text/html")
			}

			next.ServeHTTP(gzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			switch r.URL.Path {
			case "/update/", "/value/", "/updates/":
				next.ServeHTTP(w, r)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func isCompressibleContent(r *http.Request) bool {
	contentType := r.Header.Get("Content-Type")
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") ||
		r.Header.Get("Accept") == "text/html"
}

type gzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}
