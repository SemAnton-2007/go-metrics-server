package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServerAddr     string        `json:"address"`
	PollInterval   time.Duration `json:"poll_interval"`
	ReportInterval time.Duration `json:"report_interval"`
	Key            string        `json:"key"`
	RateLimit      int           `json:"rate_limit"`
	CryptoKey      string        `json:"crypto_key"`
	ConfigFile     string        `json:"-"`
}

func NewConfig() *Config {
	cfg := &Config{}

	configFile := getConfigFile()
	if configFile != "" {
		if err := cfg.loadFromFile(configFile); err != nil {
			fmt.Printf("Warning: Failed to load config file: %v\n", err)
		}
	}

	cfg.setDefaults()

	cfg.applyEnv()

	cfg.parseFlags()

	return cfg
}

func getConfigFile() string {
	fs := flag.NewFlagSet("config", flag.ContinueOnError)
	configFile := fs.String("config", "", "Path to config file")
	fs.StringVar(configFile, "c", "", "Path to config file (shorthand)")

	_ = fs.Parse(filterArgs(os.Args[1:]))
	if *configFile != "" {
		return *configFile
	}

	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	return ""
}

func (cfg *Config) loadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	type fileConfig struct {
		Address        string `json:"address"`
		PollInterval   string `json:"poll_interval"`
		ReportInterval string `json:"report_interval"`
		Key            string `json:"key"`
		RateLimit      int    `json:"rate_limit"`
		CryptoKey      string `json:"crypto_key"`
	}

	var fCfg fileConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fCfg); err != nil {
		return err
	}

	if fCfg.Address != "" {
		cfg.ServerAddr = fCfg.Address
	}
	if fCfg.PollInterval != "" {
		dur, err := time.ParseDuration(fCfg.PollInterval)
		if err != nil {
			return fmt.Errorf("invalid poll_interval: %w", err)
		}
		cfg.PollInterval = dur
	}
	if fCfg.ReportInterval != "" {
		dur, err := time.ParseDuration(fCfg.ReportInterval)
		if err != nil {
			return fmt.Errorf("invalid report_interval: %w", err)
		}
		cfg.ReportInterval = dur
	}
	if fCfg.Key != "" {
		cfg.Key = fCfg.Key
	}
	if fCfg.RateLimit > 0 {
		cfg.RateLimit = fCfg.RateLimit
	}
	if fCfg.CryptoKey != "" {
		cfg.CryptoKey = fCfg.CryptoKey
	}

	return nil
}

func (cfg *Config) setDefaults() {
	defaultServerAddr := "localhost:8080"
	defaultPollInterval := 2
	defaultReportInterval := 10
	defaultRateLimit := 1

	if cfg.ServerAddr == "" {
		cfg.ServerAddr = defaultServerAddr
	}
	if cfg.PollInterval == 0 {
		cfg.PollInterval = time.Duration(defaultPollInterval) * time.Second
	}
	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = time.Duration(defaultReportInterval) * time.Second
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = defaultRateLimit
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
}

func (cfg *Config) parseFlags() {
	fs := flag.NewFlagSet("agent-flags", flag.ContinueOnError)
	fs.StringVar(&cfg.ServerAddr, "a", cfg.ServerAddr, "Server address")
	pollInterval := fs.Int("p", int(cfg.PollInterval.Seconds()), "Poll interval (seconds)")
	reportInterval := fs.Int("r", int(cfg.ReportInterval.Seconds()), "Report interval (seconds)")
	fs.StringVar(&cfg.Key, "k", cfg.Key, "Key for hash")
	fs.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "Rate limit")
	fs.StringVar(&cfg.CryptoKey, "crypto-key", cfg.CryptoKey, "Path to public key")

	_ = fs.Parse(filterArgs(os.Args[1:]))

	cfg.PollInterval = time.Duration(*pollInterval) * time.Second
	cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
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
