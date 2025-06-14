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
	err := godotenv.Load() 
	if err != nil {
		log.Fatal("Error loading .env file:", err)
	}
	// Переопределение из env-переменных
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}else{
		flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "HTTP server address")
		flag.Parse()

	}
	if envBase := os.Getenv("BASE_URL"); envBase != "" {
		cfg.BaseURL = envBase
	}else{
		flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "Base URL for shortened URLs")
		flag.Parse()

	}
	fmt.Println(cfg.ServerAddress)

	// Валидация
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	fmt.Println(cfg)
	return &cfg, nil
}
