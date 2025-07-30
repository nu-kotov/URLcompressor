package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
)

// Config - структура конфигурации проекта.
type Config struct {
	RunAddr            string
	BaseURL            string
	FileStoragePath    string
	DatabaseConnection string
	EnableHTTPS        bool
	ConfigFileName     string
}

// FileConfig - структура конфигурации проекта из файла json.
type JSONFileConfig struct {
	RunAddr            string
	BaseURL            string
	FileStoragePath    string
	DatabaseConnection string
	EnableHTTPS        bool
	ConfigFileName     string
}

// NewConfig - конструктор конфигурации проекта.
func NewConfig() (*Config, error) {
	var config Config
	var jsonConfig JSONFileConfig

	flag.StringVar(&config.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "default schema, host and port in compressed URL")
	flag.StringVar(&config.FileStoragePath, "f", "", "Path to file with saved URLs data")
	flag.StringVar(&config.DatabaseConnection, "d", "", "Database connection string")
	flag.StringVar(&config.ConfigFileName, "c", "", "JSON config file name")
	flag.BoolVar(&config.EnableHTTPS, "s", false, "Enable HTTPS connection")

	if envConfigFileName := os.Getenv("CONFIG"); envConfigFileName != "" {
		config.ConfigFileName = envConfigFileName
	}

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		config.RunAddr = envRunAddr
	}
	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}
	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		config.FileStoragePath = envFileStoragePath
	}
	if envDatabaseConnection := os.Getenv("DATABASE_DSN"); envDatabaseConnection != "" {
		config.DatabaseConnection = envDatabaseConnection
	}
	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS == "true" {
		config.EnableHTTPS = true
	}

	flag.Parse()

	if config.ConfigFileName != "" {

		confJSONFile, err := os.Open(config.ConfigFileName)
		if err != nil {
			return nil, fmt.Errorf("openning config json file error: %w", err)
		}
		defer confJSONFile.Close()

		bytes, _ := io.ReadAll(confJSONFile)
		json.Unmarshal(bytes, &jsonConfig)

		if config.RunAddr == "" {
			config.RunAddr = jsonConfig.RunAddr
		}
		if config.BaseURL == "" {
			config.BaseURL = jsonConfig.BaseURL
		}
		if config.FileStoragePath == "" {
			config.FileStoragePath = jsonConfig.FileStoragePath
		}
		if config.DatabaseConnection == "" {
			config.DatabaseConnection = jsonConfig.DatabaseConnection
		}
		config.EnableHTTPS = jsonConfig.EnableHTTPS
	}

	return &config, nil
}
