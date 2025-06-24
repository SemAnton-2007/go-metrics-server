package main

import (
	"context"
	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	"go-metrics-server/internal/server/repository"
	"go-metrics-server/internal/server/webservers"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	RunServer(config.NewConfig())
}

func RunServer(cfg *config.Config) {
	go func() {
		log.Println("Debug server running on :6060")
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()

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
			if err := memRepo.LoadFromFile(context.Background(), cfg.FileStorage); err != nil {
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
					if err := repo.SaveToFile(context.Background(), cfg.FileStorage); err != nil {
						log.Printf("Failed to save metrics: %v\n", err)
					}
				}
			}()
			defer saveTicker.Stop()
		} else if cfg.FileStorage != "" {
			repo = newSyncSaveRepository(memRepo, cfg.FileStorage)
			defer repo.(*syncSaveRepository).Close()
		} else {
			repo = memRepo
		}
	}

	srv := webservers.NewServer(cfg, repo, db)
	log.Printf("Server is running on http://%s\n", cfg.ServerAddr)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v\n", err)
		}
	}()

	<-stop
	log.Println("Server is shutting down...")

	if cfg.DatabaseDSN == "" && cfg.FileStorage != "" {
		if err := repo.SaveToFile(context.Background(), cfg.FileStorage); err != nil {
			log.Printf("Failed to save metrics on shutdown: %v\n", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	}
	log.Println("Server stopped")
}

type syncSaveRepository struct {
	repository.MetricRepository
	filePath string
	mu       sync.Mutex
}

func newSyncSaveRepository(repo repository.MetricRepository, filePath string) *syncSaveRepository {
	return &syncSaveRepository{
		MetricRepository: repo,
		filePath:         filePath,
	}
}

func (s *syncSaveRepository) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.SaveToFile(context.Background(), s.filePath); err != nil {
		log.Printf("Failed to save metrics on close: %v\n", err)
	}
}
