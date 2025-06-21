package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"os"
)

type Config struct {
	ServerPort  string
	DatabaseURL string
	GinMode     string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		logrus.Error("Failed to load .env file: ", err)
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	config := &Config{
		ServerPort:  os.Getenv("SERVER_PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		GinMode:     os.Getenv("GIN_MODE"),
	}

	//Set default values
	if config.ServerPort == "" {
		config.ServerPort = "8080"
	}
	if config.DatabaseURL == "" {
		logrus.Error("DATABASE_URL is not set")
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	if config.GinMode == "" {
		logrus.Error("GIN_MODE is not set")
		return nil, fmt.Errorf("GIN_MODE is not set")
	}

	return config, nil
}
