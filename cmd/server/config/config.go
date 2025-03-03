package config

import (
	"flag"
	"fmt"
)

type Config struct {
	ServerAddr string // Адрес сервера
}

func NewConfig() *Config {
	cfg := &Config{}

	// Определяем флаг для адреса сервера
	flag.StringVar(&cfg.ServerAddr, "a", "localhost:8080", "Адрес HTTP-сервера")

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
