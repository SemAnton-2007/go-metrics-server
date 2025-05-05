package main

import (
	"context"
	"go-metrics-server/internal/server/config"
	"go-metrics-server/internal/server/database"
	"go-metrics-server/internal/server/storage"
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
	var store storage.MemStorage

	// Инициализируем соединение с БД, если указан DSN
	if cfg.DatabaseDSN != "" {
		db, err = database.New(cfg.DatabaseDSN)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v\n", err)
		}
		defer db.Close()
		log.Println("Connected to PostgreSQL database")

		// Используем PostgresStorage как основное хранилище
		pgStorage, err := storage.NewPostgresStorage(ctx, db.DB)
		if err != nil {
			log.Fatalf("Failed to initialize Postgres storage: %v\n", err)
		}
		store = pgStorage
	} else {
		// Старая логика с файловым или in-memory хранилищем
		baseStorage := storage.NewMemStorage()

		// Загрузка данных при старте, если разрешено
		if cfg.Restore && cfg.FileStorage != "" {
			if err := baseStorage.LoadFromFile(cfg.FileStorage); err != nil {
				log.Printf("Failed to load metrics from file: %v\n", err)
			} else {
				log.Println("Metrics loaded successfully from file")
			}
		}

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
			defer saveTicker.Stop()
		} else if cfg.FileStorage != "" {
			// Синхронное сохранение с буферизацией
			syncStorage := newSyncSaveStorage(baseStorage, cfg.FileStorage)
			store = syncStorage
			defer syncStorage.Close()
		} else {
			// Сохранение отключено
			store = baseStorage
		}
	}

	srv := webservers.NewServer(cfg, store, db)
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
		if err := store.SaveToFile(cfg.FileStorage); err != nil {
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

// syncSaveStorage обертка для синхронного сохранения с буферизацией
type syncSaveStorage struct {
	storage.MemStorage
	filePath    string
	buffer      map[string]interface{} // буфер для накопления изменений
	bufferSize  int                    // текущий размер буфера
	maxBuffer   int                    // максимальный размер буфера перед сбросом на диск
	flushTicker *time.Ticker           // тикер для периодического сброса
	mu          sync.Mutex             // мьютекс для защиты буфера
	stopChan    chan struct{}          // канал для остановки фоновой горутины
}

func newSyncSaveStorage(s storage.MemStorage, filePath string) *syncSaveStorage {
	storage := &syncSaveStorage{
		MemStorage: s,
		filePath:   filePath,
		buffer:     make(map[string]interface{}),
		maxBuffer:  10, // сбрасываем на диск каждые 10 изменений
		stopChan:   make(chan struct{}),
	}

	// Запускаем фоновую горутину для периодического сброса
	storage.flushTicker = time.NewTicker(5 * time.Second)
	go storage.backgroundFlush()

	return storage
}

func (s *syncSaveStorage) backgroundFlush() {
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

func (s *syncSaveStorage) flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.bufferSize == 0 {
		return
	}

	// Сохраняем все метрики, а не только буферизованные
	if err := s.MemStorage.SaveToFile(s.filePath); err != nil {
		log.Printf("Failed to save metrics synchronously: %v\n", err)
	}

	// Очищаем буфер
	s.buffer = make(map[string]interface{})
	s.bufferSize = 0
}

func (s *syncSaveStorage) UpdateGauge(name string, value float64) {
	s.MemStorage.UpdateGauge(name, value)
	s.bufferUpdate(name, value)
}

func (s *syncSaveStorage) UpdateCounter(name string, value int64) {
	s.MemStorage.UpdateCounter(name, value)
	s.bufferUpdate(name, value)
}

func (s *syncSaveStorage) bufferUpdate(name string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.buffer[name] = value
	s.bufferSize++

	// Если буфер заполнен, сбрасываем на диск
	if s.bufferSize >= s.maxBuffer {
		go s.flush()
	}
}

func (s *syncSaveStorage) Close() {
	close(s.stopChan)
	s.flush() // финальный сброс перед закрытием
}
