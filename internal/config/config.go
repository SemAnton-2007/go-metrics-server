package config

import (
	"flag"
	"os"
	"strings"
)

// CommonConfig содержит общие поля конфигурации
type CommonConfig struct {
	ServerAddr string `json:"address"`
	CryptoKey  string `json:"crypto_key"`
}

// GetConfigFile возвращает путь к конфигурационному файлу
func GetConfigFile() string {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	configFile := fs.String("config", "", "Path to config file")
	fs.StringVar(configFile, "c", "", "Path to config file (shorthand)")

	_ = fs.Parse(FilterArgs(os.Args[1:]))
	if *configFile != "" {
		return *configFile
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	return ""
}

// FilterArgs удаляет тестовые флаги из аргументов
func FilterArgs(args []string) []string {
	var filtered []string
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-test.") {
			filtered = append(filtered, args[i])
		}
	}
	return filtered
}
