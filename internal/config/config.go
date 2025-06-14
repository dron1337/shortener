package config

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func LoadConfig() (*Config, error) {
	// Значения по умолчанию
	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
	}
	flagAddr := flag.String("a", "", "HTTP server address")
	flagBase := flag.String("b", "", "Base URL for shortened URLs")
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	if *flagAddr != "" {
		cfg.ServerAddress = *flagAddr
	} else if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}

	if *flagBase != "" {
		cfg.BaseURL = *flagBase
	} else if envBase := os.Getenv("BASE_URL"); envBase != "" {
		cfg.BaseURL = envBase
	}
	// Валидация
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	return &cfg, nil
}
