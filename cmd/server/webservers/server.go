package webservers

import (
	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/handlers"
	"go-metrics-server/cmd/server/storage"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewServer(cfg *config.Config, storage storage.MemStorage) *http.Server {
	r := chi.NewRouter()

	// Регистрируем обработчики
	r.Post("/update/{type}/{name}/{value}", handlers.UpdateMetricHandler(storage))
	r.Get("/value/{type}/{name}", handlers.GetMetricValueHandler(storage))
	r.Get("/", handlers.GetAllMetricsHandler(storage))

	return &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: r,
	}
}
