package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerAddr    = "localhost:8080" // Значение по умолчанию для адреса сервера
	defaultStoreInterval = 300 * time.Second
	defaultFileStorage   = "/tmp/metrics-db.json"
	defaultRestore       = true
	defaultDatabaseDSN   = ""
)

type Config struct {
	ServerAddr    string        // Адрес сервера
	StoreInterval time.Duration // Интервал сохранения на диск
	FileStorage   string        // Путь к файлу сохранения
	Restore       bool          // Загружать данные при старте
	DatabaseDSN   string
}

func NewConfig() *Config {
	cfg := &Config{}

	// Получаем значения из переменных окружения или используем значения по умолчанию
	serverAddr := getEnvOrDefault("ADDRESS", defaultServerAddr)
	storeInterval := parseDuration(getEnvOrDefault("STORE_INTERVAL", defaultStoreInterval.String()))
	fileStorage := getEnvOrDefault("FILE_STORAGE_PATH", defaultFileStorage)
	restore := parseBool(getEnvOrDefault("RESTORE", strconv.FormatBool(defaultRestore)))
	databaseDSN := getEnvOrDefault("DATABASE_DSN", defaultDatabaseDSN)

	// Используем локальный FlagSet для изоляции флагов
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", serverAddr, "Адрес HTTP-сервера")
	fs.DurationVar(&cfg.StoreInterval, "i", storeInterval, "Интервал сохранения на диск (секунды)")
	fs.StringVar(&cfg.FileStorage, "f", fileStorage, "Файл для сохранения метрик")
	fs.BoolVar(&cfg.Restore, "r", restore, "Загружать данные при старте")
	fs.StringVar(&cfg.DatabaseDSN, "d", databaseDSN, "DSN")

	// Фильтруем аргументы, чтобы игнорировать флаги go test
	args := filterArgs(os.Args[1:]) // Игнорируем первый аргумент (имя программы)

	// Парсим только отфильтрованные аргументы
	if err := fs.Parse(args); err != nil {
		fmt.Println(fmt.Errorf("ошибка при парсинге флагов: %w", err))
		os.Exit(1)
	}

	// Проверяем, что нет неизвестных флагов
	if fs.NArg() > 0 {
		fmt.Println(fmt.Errorf("ошибка: неизвестные флаги или аргументы"))
		fs.Usage()
		os.Exit(1)
	}

	return cfg
}

// filterArgs удаляет флаги go test из списка аргументов
func filterArgs(args []string) []string {
	var filtered []string
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-test.") {
			filtered = append(filtered, args[i])
		}
	}
	return filtered
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseDuration(value string) time.Duration {
	if dur, err := time.ParseDuration(value); err == nil {
		return dur
	}
	// Попробуем интерпретировать как число секунд
	if sec, err := strconv.Atoi(value); err == nil {
		return time.Duration(sec) * time.Second
	}
	return defaultStoreInterval
}

func parseBool(value string) bool {
	if b, err := strconv.ParseBool(value); err == nil {
		return b
	}
	return defaultRestore
}
