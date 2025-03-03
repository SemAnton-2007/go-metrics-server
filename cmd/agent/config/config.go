package config

import (
	"flag"
	"fmt"
	"time"
)

type Config struct {
	ServerAddr     string        // Адрес сервера
	PollInterval   time.Duration // Интервал опроса метрик
	ReportInterval time.Duration // Интервал отправки метрик
}

func NewConfig() *Config {
	cfg := &Config{}

	// Определяем флаги
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "Адрес HTTP-сервера")
	pollInterval := flag.Int("p", 2, "Интервал опроса метрик (в секундах)")
	reportInterval := flag.Int("r", 10, "Интервал отправки метрик (в секундах)")

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
