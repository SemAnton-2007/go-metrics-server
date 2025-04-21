package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr     string        // Адрес сервера
	PollInterval   time.Duration // Интервал опроса метрик
	ReportInterval time.Duration // Интервал отправки метрик
	Key            string        // Ключ для подписи данных
}

func NewConfig() *Config {
	cfg := &Config{}

	// Значения по умолчанию
	defaultServerAddr := "localhost:8080"
	defaultPollInterval := 2
	defaultReportInterval := 10
	defaultKey := ""

	// Получаем значения из переменных окружения
	if addr := os.Getenv("ADDRESS"); addr != "" {
		defaultServerAddr = addr
	}
	if pollIntervalStr := os.Getenv("POLL_INTERVAL"); pollIntervalStr != "" {
		if pollInterval, err := strconv.Atoi(pollIntervalStr); err == nil {
			defaultPollInterval = pollInterval
		}
	}
	if reportIntervalStr := os.Getenv("REPORT_INTERVAL"); reportIntervalStr != "" {
		if reportInterval, err := strconv.Atoi(reportIntervalStr); err == nil {
			defaultReportInterval = reportInterval
		}
	}
	if key := os.Getenv("KEY"); key != "" {
		defaultKey = key
	}

	// Используем локальный FlagSet для изоляции флагов
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", defaultServerAddr, "Адрес HTTP-сервера")
	pollInterval := fs.Int("p", defaultPollInterval, "Интервал опроса метрик (в секундах)")
	reportInterval := fs.Int("r", defaultReportInterval, "Интервал отправки метрик (в секундах)")
	fs.StringVar(&cfg.Key, "k", defaultKey, "Ключ для подписи данных")

	// Фильтруем аргументы, чтобы игнорировать флаги go test
	args := filterArgs(os.Args[1:]) // Игнорируем первый аргумент (имя программы)

	// Парсим только отфильтрованные аргументы
	if err := fs.Parse(args); err != nil {
		fmt.Println("Ошибка при парсинге флагов:", err)
		os.Exit(1)
	}

	// Преобразуем интервалы в time.Duration
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second

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
