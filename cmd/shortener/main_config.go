package main

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

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		config.RunAddr = envRunAddr
	} else {
		flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.RunAddr = envBaseURL
	} else {
		flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "default schema, host and port in compressed URL")
	}

	flag.Parse()

	return config
}
