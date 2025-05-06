package main

import (
	"context"
	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/webservers"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	// Создаем основной контекст приложения
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.NewConfig()

	var db *database.DB
	var err error
	var repo repository.MetricRepository

	// Инициализируем соединение с БД, если указан DSN
	if cfg.DatabaseDSN != "" {
		db, err = database.New(cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v\n", err)
		}
		defer db.Close()
		log.Println("Connected to PostgreSQL database")

		pgRepo, err := repository.NewPostgresRepository(db.DB)
		if err != nil {
			log.Fatalf("Failed to initialize Postgres repository: %v\n", err)
		}
		repo = pgRepo
	} else {
		memRepo := repository.NewMemoryRepository()

		if cfg.Restore && cfg.FileStorage != "" {
			if err := memRepo.LoadFromFile(ctx, cfg.FileStorage); err != nil {
				log.Printf("Failed to load metrics from file: %v\n", err)
			} else {
				log.Println("Metrics loaded successfully from file")
			}
		}

		var saveTicker *time.Ticker

		if cfg.StoreInterval > 0 && cfg.FileStorage != "" {
			repo = memRepo
			saveTicker = time.NewTicker(cfg.StoreInterval)
			go func() {
				for range saveTicker.C {
					if err := repo.SaveToFile(ctx, cfg.FileStorage); err != nil {
						log.Printf("Failed to save metrics: %v\n", err)
					} else {
						log.Println("Metrics saved successfully")
					}
				}
			}()
			defer saveTicker.Stop()
		} else if cfg.FileStorage != "" {
			syncRepo := newSyncSaveRepository(memRepo, cfg.FileStorage)
			repo = syncRepo
			defer syncRepo.Close()
		} else {
			repo = memRepo
		}
	}

	srv := webservers.NewServer(cfg, repo, db)
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

	// Сохранение данных перед выходом (если не используется БД)
	if cfg.DatabaseDSN == "" && cfg.FileStorage != "" {
		if err := repo.SaveToFile(ctx, cfg.FileStorage); err != nil {
			log.Printf("Failed to save metrics on shutdown: %v\n", err)
		} else {
			log.Println("Metrics saved successfully on shutdown")
		}
	}

	// Graceful shutdown сервера
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}
	log.Println("Server stopped")
}

type syncSaveRepository struct {
	repository.MetricRepository
	filePath    string
	buffer      map[string]interface{}
	bufferSize  int
	maxBuffer   int
	flushTicker *time.Ticker
	mu          sync.Mutex
	stopChan    chan struct{}
}

func newSyncSaveRepository(repo repository.MetricRepository, filePath string) *syncSaveRepository {
	storage := &syncSaveRepository{
		MetricRepository: repo,
		filePath:         filePath,
		buffer:           make(map[string]interface{}),
		maxBuffer:        10,
		stopChan:         make(chan struct{}),
	}

	// Запускаем фоновую горутину для периодического сброса
	storage.flushTicker = time.NewTicker(5 * time.Second)
	go storage.backgroundFlush()

	return storage
}

func (s *syncSaveRepository) backgroundFlush() {
	for {
		select {
		case <-s.flushTicker.C:
			s.flush()
		case <-s.stopChan:
			s.flushTicker.Stop()
			return
		}
	}
}

func (s *syncSaveRepository) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bufferSize == 0 {
		return
	}

	if err := s.MetricRepository.SaveToFile(context.Background(), s.filePath); err != nil {
		log.Printf("Failed to save metrics synchronously: %v\n", err)
	}

	// Очищаем буфер
	s.buffer = make(map[string]interface{})
	s.bufferSize = 0
}

func (s *syncSaveRepository) UpdateGauge(ctx context.Context, name string, value float64) error {
	err := s.MetricRepository.UpdateGauge(ctx, name, value)
	s.bufferUpdate(name, value)
	return err
}

func (s *syncSaveRepository) UpdateCounter(ctx context.Context, name string, value int64) error {
	err := s.MetricRepository.UpdateCounter(ctx, name, value)
	s.bufferUpdate(name, value)
	return err
}

func (s *syncSaveRepository) bufferUpdate(name string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buffer[name] = value
	s.bufferSize++

	// Если буфер заполнен, сбрасываем на диск
	if s.bufferSize >= s.maxBuffer {
		go s.flush()
	}
}

func (s *syncSaveRepository) Close() {
	close(s.stopChan)
	s.flush()
}
