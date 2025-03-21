package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

const (
	defaultServerAddr = "localhost:8080" // Значение по умолчанию для адреса сервера
)

type Config struct {
	ServerAddr string // Адрес сервера
}

func NewConfig() *Config {
	cfg := &Config{}

	// Получаем значение из переменной окружения
	serverAddr := defaultServerAddr
	if addr := os.Getenv("ADDRESS"); addr != "" {
		serverAddr = addr
	}

	// Используем локальный FlagSet для изоляции флагов
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", serverAddr, "Адрес HTTP-сервера")

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
		fmt.Println(fmt.Errorf("неизвестные флаги"))
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
