package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"go-metrics-server/internal/config"
)

const (
	defaultServerAddr    = "localhost:8080"
	defaultStoreInterval = 300 * time.Second
	defaultFileStorage   = "/tmp/metrics-db.json"
	defaultRestore       = true
	defaultDatabaseDSN   = ""
	defaultKey           = ""
	defaultGRPCAddress   = "localhost:3200"
)

type Config struct {
	config.CommonConfig

	StoreInterval time.Duration `json:"store_interval"`
	FileStorage   string        `json:"store_file"`
	Restore       bool          `json:"restore"`
	DatabaseDSN   string        `json:"database_dsn"`
	Key           string        `json:"key"`
	CryptoKey     string        `json:"crypto_key"`
	TrustedSubnet string        `json:"trusted_subnet"`
}

func NewConfig() *Config {
	cfg := &Config{}

	if configFile := config.GetConfigFile(); configFile != "" {
		if err := cfg.loadFromFile(configFile); err != nil {
			log.Printf("Warning: Failed to load config file: %v", err)
		}
	}

	cfg.setDefaults()
	cfg.applyEnv()
	cfg.parseFlags()

	return cfg
}

func (cfg *Config) loadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var fileCfg struct {
		config.CommonConfig
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		Restore       bool   `json:"restore"`
		DatabaseDSN   string `json:"database_dsn"`
		Key           string `json:"key"`
		TrustedSubnet string `json:"trusted_subnet"`
	}

	if err := json.NewDecoder(file).Decode(&fileCfg); err != nil {
		return err
	}

	cfg.CommonConfig = fileCfg.CommonConfig

	if fileCfg.StoreInterval != "" {
		if dur, err := time.ParseDuration(fileCfg.StoreInterval); err == nil {
			cfg.StoreInterval = dur
		}
	}

	cfg.FileStorage = fileCfg.StoreFile
	cfg.Restore = fileCfg.Restore
	cfg.DatabaseDSN = fileCfg.DatabaseDSN
	cfg.Key = fileCfg.Key
	cfg.TrustedSubnet = fileCfg.TrustedSubnet

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
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = defaultGRPCAddress
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
	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		cfg.TrustedSubnet = envTrustedSubnet
	}
	if envGRPCAddress := os.Getenv("GRPC_ADDRESS"); envGRPCAddress != "" {
		cfg.GRPCAddress = envGRPCAddress
	}
}

func (cfg *Config) parseFlags() {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	fs.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Server address")
	fs.DurationVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "Store interval")
	fs.StringVar(&cfg.FileStorage, "f", cfg.FileStorage, "Store file")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "Restore")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "Database DSN")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "Key for hash")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to private key")
	fs.StringVar(&cfg.TrustedSubnet, "t", cfg.TrustedSubnet, "Trusted subnet (CIDR)")
	fs.StringVar(&cfg.GRPCAddress, "g", cfg.GRPCAddress, "gRPC server address")

	fs.Parse(os.Args[1:])
}
