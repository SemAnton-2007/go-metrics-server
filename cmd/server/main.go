package main

import (
	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/metrics_server"
	"go-metrics-server/cmd/server/storage"
	"log"
)

func main() {
	cfg := config.NewConfig()
	storage := storage.NewMemStorage()

	srv := metrics_server.NewServer(cfg, storage)
	log.Printf("Server is running on http://%s\n", cfg.ServerAddr)
	log.Fatal(srv.ListenAndServe())
}
