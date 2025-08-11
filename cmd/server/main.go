package main

import (
	"context"
	"go-metrics-server/internal/buildinfo"
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
	buildinfo.Print()
	RunServer(config.NewConfig())
}

func RunServer(cfg *config.Config) {
	go func() {
		log.Println("Debug server running on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil && err != http.ErrServerClosed {
			log.Printf("Debug server error: %v\n", err)
		}
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
		defer func() {
			if err := db.Close(); err != nil {
				log.Printf("Failed to close database connection: %v\n", err)
			}
		}()
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
			repo = newSyncSaveRepository(memRepo, cfg.FileStorage, log.Default())
			defer func() {
				if syncRepo, ok := repo.(*syncSaveRepository); ok {
					if err := syncRepo.Close(); err != nil {
						log.Printf("Failed to close sync repository: %v\n", err)
					}
				}
			}()
		} else {
			repo = memRepo
		}
	}

	srv := webservers.NewServer(cfg, repo, db)
	if cfg.GRPCAddress == "" {
		log.Printf("HTTP server running on http://%s", cfg.ServerAddr)
	} else {
		log.Printf("HTTP server running on http://%s, gRPC on %s", cfg.ServerAddr, cfg.GRPCAddress)
	}
	log.Printf("Server is running on http://%s\n", cfg.ServerAddr)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case sig := <-sigChan:
		log.Printf("Received signal: %v. Shutting down...", sig)
	case err := <-serverErr:
		log.Fatalf("Server error: %v\n", err)
	}

	if cfg.DatabaseDSN == "" && cfg.FileStorage != "" {
		log.Println("Saving metrics before shutdown...")
		saveCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := repo.SaveToFile(saveCtx, cfg.FileStorage); err != nil {
			log.Printf("Failed to save metrics on shutdown: %v\n", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v\n", err)
	} else {
		log.Println("Server stopped gracefully")
	}
}

type syncSaveRepository struct {
	repository.MetricRepository
	filePath string
	mu       sync.Mutex
	logger   *log.Logger
}

func newSyncSaveRepository(repo repository.MetricRepository, filePath string, logger *log.Logger) *syncSaveRepository {
	return &syncSaveRepository{
		MetricRepository: repo,
		filePath:         filePath,
		logger:           logger,
	}
}

func (s *syncSaveRepository) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.SaveToFile(context.Background(), s.filePath); err != nil {
		s.logger.Printf("Failed to save metrics on close: %v\n", err)
		return err
	}
	return nil
}
