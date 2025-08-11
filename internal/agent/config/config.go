package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"go-metrics-server/internal/config"
)

const defaultGRPCAddress = "localhost:3200"

type Config struct {
	config.CommonConfig

	PollInterval   time.Duration `json:"poll_interval"`
	ReportInterval time.Duration `json:"report_interval"`
	Key            string        `json:"key"`
	RateLimit      int           `json:"rate_limit"`
	GRPCAddress    string        `json:"grpc_address"`
}

func NewConfig() *Config {
	cfg := &Config{}

	if configFile := config.GetConfigFile(); configFile != "" {
		if err := cfg.loadFromFile(configFile); err != nil {
			fmt.Printf("Warning: Failed to load config file: %v\n", err)
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
		PollInterval   string `json:"poll_interval"`
		ReportInterval string `json:"report_interval"`
		Key            string `json:"key"`
		RateLimit      int    `json:"rate_limit"`
	}

	if err := json.NewDecoder(file).Decode(&fileCfg); err != nil {
		return err
	}

	cfg.CommonConfig = fileCfg.CommonConfig

	if fileCfg.PollInterval != "" {
		if dur, err := time.ParseDuration(fileCfg.PollInterval); err == nil {
			cfg.PollInterval = dur
		}
	}

	if fileCfg.ReportInterval != "" {
		if dur, err := time.ParseDuration(fileCfg.ReportInterval); err == nil {
			cfg.ReportInterval = dur
		}
	}

	cfg.Key = fileCfg.Key
	cfg.RateLimit = fileCfg.RateLimit

	return nil
}

func (cfg *Config) setDefaults() {
	if cfg.ServerAddr == "" {
		cfg.ServerAddr = "localhost:8080"
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = 2 * time.Second
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = 10 * time.Second
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = 1
	}
	if cfg.GRPCAddress == "" {
		cfg.GRPCAddress = defaultGRPCAddress
	}
}

func (cfg *Config) applyEnv() {
	if addr := os.Getenv("ADDRESS"); addr != "" {
		cfg.ServerAddr = addr
	}
	if pollIntervalStr := os.Getenv("POLL_INTERVAL"); pollIntervalStr != "" {
		if pollInterval, err := strconv.Atoi(pollIntervalStr); err == nil {
			cfg.PollInterval = time.Duration(pollInterval) * time.Second
		}
	}
	if reportIntervalStr := os.Getenv("REPORT_INTERVAL"); reportIntervalStr != "" {
		if reportInterval, err := strconv.Atoi(reportIntervalStr); err == nil {
			cfg.ReportInterval = time.Duration(reportInterval) * time.Second
		}
	}
	if key := os.Getenv("KEY"); key != "" {
		cfg.Key = key
	}
	if rateLimitStr := os.Getenv("RATE_LIMIT"); rateLimitStr != "" {
		if rateLimit, err := strconv.Atoi(rateLimitStr); err == nil {
			cfg.RateLimit = rateLimit
		}
	}
	if cryptoKey := os.Getenv("CRYPTO_KEY"); cryptoKey != "" {
		cfg.CryptoKey = cryptoKey
	}
	if envGRPCAddress := os.Getenv("GRPC_ADDRESS"); envGRPCAddress != "" {
		cfg.GRPCAddress = envGRPCAddress
	}
}

func (cfg *Config) parseFlags() {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	fs.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Server address")
	pollInterval := fs.Int("p", int(cfg.PollInterval.Seconds()), "Poll interval (seconds)")
	reportInterval := fs.Int("r", int(cfg.ReportInterval.Seconds()), "Report interval (seconds)")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "Key for hash")
	fs.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "Rate limit")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to public key")
	fs.StringVar(&cfg.GRPCAddress, "g", cfg.GRPCAddress, "gRPC server address")

	fs.Parse(os.Args[1:])

	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
}
