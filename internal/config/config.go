package config

import (
	"os"
)

type CommonConfig struct {
	ServerAddr  string `json:"address"`
	CryptoKey   string `json:"crypto_key"`
	GRPCAddress string `json:"grpc_address"`
}

func GetConfigFile() string {
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		return envConfig
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		if args[i] == "-config" || args[i] == "-c" {
			if i+1 < len(args) {
				return args[i+1]
			}
		}
	}

	return ""
}
