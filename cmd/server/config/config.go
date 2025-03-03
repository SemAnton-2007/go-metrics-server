package config

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	ServerAddr string // Адрес сервера
}

func NewConfig() *Config {
	cfg := &Config{}

	// Значение по умолчанию
	defaultServerAddr := "localhost:8080"

	// Получаем значение из переменной окружения
	if addr := os.Getenv("ADDRESS"); addr != "" {
		defaultServerAddr = addr
	}

	// Определяем флаг
	flag.StringVar(&cfg.ServerAddr, "a", defaultServerAddr, "Адрес HTTP-сервера")

	// Парсим флаги
	flag.Parse()

	// Проверяем, что нет неизвестных флагов
	if flag.NArg() > 0 {
		fmt.Println("Ошибка: неизвестные флаги или аргументы")
		flag.Usage()
		panic("неизвестные флаги")
	}

	return cfg
}
