package config

import (
	"flag"
	"fmt"
	"log"
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
	defaultKey           = ""
	defaultCryptoKey     = ""
)

// Config holds application configuration parameters.
// Fields can be set via command line flags or environment variables.
type Config struct {
	ServerAddr    string        // Server address in host:port format (env: ADDRESS, flag: -a)
	StoreInterval time.Duration // Interval for saving metrics to disk (env: STORE_INTERVAL, flag: -i)
	FileStorage   string        // Path to metrics storage file (env: FILE_STORAGE_PATH, flag: -f)
	Restore       bool          // Whether to load metrics at startup (env: RESTORE, flag: -r)
	DatabaseDSN   string        // PostgreSQL connection DSN (env: DATABASE_DSN, flag: -d)
	Key           string        // Secret key for hash verification (env: KEY, flag: -k)
	CryptoKey     string        // Path to private key file (env: CRYPTO_KEY, flag: -crypto-key)
}

// NewConfig creates a new Config instance by parsing command line flags and environment variables.
// Environment variables take precedence over default values but are overridden by command line flags.
func NewConfig() *Config {
	cfg := &Config{}

	// Получаем значения из переменных окружения или используем значения по умолчанию
	serverAddr := getEnvOrDefault("ADDRESS", defaultServerAddr)
	storeInterval := parseDuration(getEnvOrDefault("STORE_INTERVAL", defaultStoreInterval.String()))
	fileStorage := getEnvOrDefault("FILE_STORAGE_PATH", defaultFileStorage)
	restore := parseBool(getEnvOrDefault("RESTORE", strconv.FormatBool(defaultRestore)))
	databaseDSN := getEnvOrDefault("DATABASE_DSN", defaultDatabaseDSN)
	key := getEnvOrDefault("KEY", defaultKey)
	cryptoKey := getEnvOrDefault("CRYPTO_KEY", defaultCryptoKey)

	// Используем локальный FlagSet для изоляции флагов
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", serverAddr, "Адрес HTTP-сервера")
	fs.DurationVar(&cfg.StoreInterval, "i", storeInterval, "Интервал сохранения на диск (секунды)")
	fs.StringVar(&cfg.FileStorage, "f", fileStorage, "Файл для сохранения метрик")
	fs.BoolVar(&cfg.Restore, "r", restore, "Загружать данные при старте")
	fs.StringVar(&cfg.DatabaseDSN, "d", databaseDSN, "DSN")
	fs.StringVar(&cfg.Key, "k", key, "Ключ для подписи данных")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cryptoKey, "Путь к файлу с приватным ключом")

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
	dur, err := time.ParseDuration(value)
	if err == nil {
		return dur
	}

	// Пробуем интерпретировать как число секунд
	sec, err := strconv.Atoi(value)
	if err == nil {
		return time.Duration(sec) * time.Second
	}

	log.Printf("Warning: Invalid STORE_INTERVAL value: %s, using default %v", value, defaultStoreInterval)
	return defaultStoreInterval
}

func parseBool(value string) bool {
	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Warning: Invalid RESTORE value: %s, using default %v", value, defaultRestore)
		return defaultRestore
	}
	return b
}
