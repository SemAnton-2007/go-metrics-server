package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerAddr     string        // Адрес сервера
	PollInterval   time.Duration // Интервал опроса метрик
	ReportInterval time.Duration // Интервал отправки метрик
}

func NewConfig() *Config {
	cfg := &Config{}

	// Значения по умолчанию
	defaultServerAddr := "localhost:8080"
	defaultPollInterval := 2
	defaultReportInterval := 10

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

	// Определяем флаги
	flag.StringVar(&cfg.ServerAddr, "a", defaultServerAddr, "Адрес HTTP-сервера")
	pollInterval := flag.Int("p", defaultPollInterval, "Интервал опроса метрик (в секундах)")
	reportInterval := flag.Int("r", defaultReportInterval, "Интервал отправки метрик (в секундах)")

	// Парсим флаги
	flag.Parse()

	// Проверяем, что нет неизвестных флагов
	if flag.NArg() > 0 {
		fmt.Println("Ошибка: неизвестные флаги или аргументы")
		flag.Usage()
		panic("неизвестные флаги")
	}

	// Преобразуем интервалы в time.Duration
	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second

	return cfg
}
