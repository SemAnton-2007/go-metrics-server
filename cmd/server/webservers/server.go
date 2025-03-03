package webservers

import (
	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/handlers"
	"go-metrics-server/cmd/server/storage"
	"net/http"
)

func NewServer(cfg *config.Config, storage storage.MemStorage) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/update/", handlers.UpdateMetricHandler(storage))

	return &http.Server{
		Addr:    cfg.ServerAddr,
		Handler: mux,
	}
}
