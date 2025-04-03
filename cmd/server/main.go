package main

import (
	"context"
	"go-metrics-server/cmd/server/config"
	"go-metrics-server/cmd/server/storage"
	"go-metrics-server/cmd/server/webservers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := config.NewConfig()
	baseStorage := storage.NewMemStorage()

	// Загрузка данных при старте, если разрешено
	if cfg.Restore && cfg.FileStorage != "" {
		if err := baseStorage.LoadFromFile(cfg.FileStorage); err != nil {
			log.Printf("Failed to load metrics from file: %v\n", err)
		} else {
			log.Println("Metrics loaded successfully from file")
		}
	}

	var store storage.MemStorage
	var saveTicker *time.Ticker

	if cfg.StoreInterval > 0 && cfg.FileStorage != "" {
		// Периодическое сохранение
		store = baseStorage
		saveTicker = time.NewTicker(cfg.StoreInterval)
		go func() {
			for range saveTicker.C {
				if err := store.SaveToFile(cfg.FileStorage); err != nil {
					log.Printf("Failed to save metrics: %v\n", err)
				} else {
					log.Println("Metrics saved successfully")
				}
			}
		}()
	} else if cfg.FileStorage != "" {
		// Синхронное сохранение
		store = newSyncSaveStorage(baseStorage, cfg.FileStorage)
	} else {
		// Сохранение отключено
		store = baseStorage
	}

	srv := webservers.NewServer(cfg, store)
	log.Printf("Server is running on http://%s\n", cfg.ServerAddr)

	// Обработка graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")

	// Остановка тикера сохранения
	if saveTicker != nil {
		saveTicker.Stop()
	}

	// Сохранение данных перед выходом
	if cfg.FileStorage != "" {
		if err := store.SaveToFile(cfg.FileStorage); err != nil {
			log.Printf("Failed to save metrics on shutdown: %v\n", err)
		} else {
			log.Println("Metrics saved successfully on shutdown")
		}
	}

	// Graceful shutdown сервера
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}
	log.Println("Server stopped")
}

// syncSaveStorage обертка для синхронного сохранения
type syncSaveStorage struct {
	storage.MemStorage
	filePath string
}

func newSyncSaveStorage(s storage.MemStorage, filePath string) *syncSaveStorage {
	return &syncSaveStorage{
		MemStorage: s,
		filePath:   filePath,
	}
}

func (s *syncSaveStorage) UpdateGauge(name string, value float64) {
	s.MemStorage.UpdateGauge(name, value)
	s.save()
}

func (s *syncSaveStorage) UpdateCounter(name string, value int64) {
	s.MemStorage.UpdateCounter(name, value)
	s.save()
}

func (s *syncSaveStorage) save() {
	if err := s.MemStorage.SaveToFile(s.filePath); err != nil {
		log.Printf("Failed to save metrics synchronously: %v\n", err)
	}
}
