package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr string
	BaseURL string
}

func ParseConfig() Config {
	var config Config

	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "default schema, host and port in compressed URL")

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		config.RunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	flag.Parse()

	return config
}
