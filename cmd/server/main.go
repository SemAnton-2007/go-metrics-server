package main

import (
	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/storage"
	"go-metrics-server/cmd/server/webservers"
	"log"
)

func main() {
	cfg := config.NewConfig()
	storage := storage.NewMemStorage()

	srv := webservers.NewServer(cfg, storage)
	log.Printf("Server is running on http://%s\n", cfg.ServerAddr)
	log.Fatal(srv.ListenAndServe())
}
