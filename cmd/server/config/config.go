package config

type Config struct {
	ServerAddr string // Адрес сервера
}

func NewConfig() *Config {
	return &Config{
		ServerAddr: "localhost:8080", // Значение по умолчанию
	}
}
