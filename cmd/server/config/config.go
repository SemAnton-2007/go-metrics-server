package config

import (
	"flag"
	"fmt"
	"os"
	"strings"
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

	// Используем локальный FlagSet для изоляции флагов
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", defaultServerAddr, "Адрес HTTP-сервера")

	// Фильтруем аргументы, чтобы игнорировать флаги go test
	args := filterArgs(os.Args[1:]) // Игнорируем первый аргумент (имя программы)

	// Парсим только отфильтрованные аргументы
	if err := fs.Parse(args); err != nil {
		fmt.Println("Ошибка при парсинге флагов:", err)
		os.Exit(1)
	}

	// Проверяем, что нет неизвестных флагов
	if fs.NArg() > 0 {
		fmt.Println("Ошибка: неизвестные флаги или аргументы")
		fs.Usage()
		panic("неизвестные флаги")
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
