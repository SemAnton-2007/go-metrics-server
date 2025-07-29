package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultServerAddr    = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultFileStorage   = "/tmp/metrics-db.json"
	defaultRestore       = true
	defaultDatabaseDSN   = ""
	defaultKey           = ""
	defaultCryptoKey     = ""
	defaultConfigFile    = ""
)

type Config struct {
	ServerAddr    string        `json:"address"`
	StoreInterval time.Duration `json:"store_interval"`
	FileStorage   string        `json:"store_file"`
	Restore       bool          `json:"restore"`
	DatabaseDSN   string        `json:"database_dsn"`
	Key           string        `json:"key"`
	CryptoKey     string        `json:"crypto_key"`
	ConfigFile    string        `json:"-"`
}

func NewConfig() *Config {
	cfg := &Config{}

	configFile := getConfigFile()
	if configFile != "" {
		if err := cfg.loadFromFile(configFile); err != nil {
			log.Printf("Warning: Failed to load config file: %v", err)
		}
	}

	cfg.setDefaults()

	cfg.applyEnv()

	cfg.parseFlags()

	return cfg
}

func getConfigFile() string {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	configFile := fs.String("config", defaultConfigFile, "Path to config file")
	fs.StringVar(configFile, "c", defaultConfigFile, "Path to config file (shorthand)")

	_ = fs.Parse(filterArgs(os.Args[1:]))
	if *configFile != "" {
		return *configFile
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	return defaultConfigFile
}

func (cfg *Config) loadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	type fileConfig struct {
		Address       string `json:"address"`
		Restore       bool   `json:"restore"`
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		DatabaseDSN   string `json:"database_dsn"`
		Key           string `json:"key"`
		CryptoKey     string `json:"crypto_key"`
	}

	var fCfg fileConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fCfg); err != nil {
		return err
	}

	if fCfg.Address != "" {
		cfg.ServerAddr = fCfg.Address
	}
	cfg.Restore = fCfg.Restore
	if fCfg.StoreInterval != "" {
		dur, err := time.ParseDuration(fCfg.StoreInterval)
		if err != nil {
			return fmt.Errorf("invalid store_interval: %w", err)
		}
		cfg.StoreInterval = dur
	}
	if fCfg.StoreFile != "" {
		cfg.FileStorage = fCfg.StoreFile
	}
	if fCfg.DatabaseDSN != "" {
		cfg.DatabaseDSN = fCfg.DatabaseDSN
	}
	if fCfg.Key != "" {
		cfg.Key = fCfg.Key
	}
	if fCfg.CryptoKey != "" {
		cfg.CryptoKey = fCfg.CryptoKey
	}

	return nil
}

func (cfg *Config) setDefaults() {
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.StoreInterval == 0 {
		cfg.StoreInterval = defaultStoreInterval
	}
	if cfg.FileStorage == "" {
		cfg.FileStorage = defaultFileStorage
	}
	if !cfg.Restore {
		cfg.Restore = defaultRestore
	}
}

func (cfg *Config) applyEnv() {
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		cfg.ServerAddr = envAddr
	}
	if envRestore := os.Getenv("RESTORE"); envRestore != "" {
		if restore, err := strconv.ParseBool(envRestore); err == nil {
			cfg.Restore = restore
		}
	}
	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		if dur, err := time.ParseDuration(envStoreInterval); err == nil {
			cfg.StoreInterval = dur
		}
	}
	if envStoreFile := os.Getenv("FILE_STORAGE_PATH"); envStoreFile != "" {
		cfg.FileStorage = envStoreFile
	}
	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}
	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.Key = envKey
	}
	if envCryptoKey := os.Getenv("CRYPTO_KEY"); envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
	}
}

func (cfg *Config) parseFlags() {
	fs := flag.NewFlagSet("server-flags", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Server address")
	fs.DurationVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "Store interval")
	fs.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "Store file")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "Key for hash")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to private key")

	_ = fs.Parse(filterArgs(os.Args[1:]))
}

func filterArgs(args []string) []string {
	var filtered []string
	for i := 0; i < len(args); i++ {
		if !strings.HasPrefix(args[i], "-test.") {
			filtered = append(filtered, args[i])
		}
	}
	return filtered
}
