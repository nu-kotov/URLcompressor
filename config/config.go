package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr         string
	BaseURL         string
	FileStoragePath string
}

func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "default schema, host and port in compressed URL")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/urls_data.json", "Path to file with saved URLs data")

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		config.RunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		config.FileStoragePath = envFileStoragePath
	}

	flag.Parse()

	return config
}
