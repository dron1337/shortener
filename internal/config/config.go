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
	FileName      string
}

func LoadConfig() (*Config, error) {
	// Значения по умолчанию
	cfg := Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		FileName:      "tmp/short-url-db.json",
	}
	flagAddr := flag.String("a", "", "HTTP server address")
	flagBase := flag.String("b", "", "Base URL for shortened URLs")
	flagFileName := flag.String("f", "", "File name")
	flag.Parse()
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
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
	if *flagFileName != "" {
		cfg.FileName = *flagFileName
	} else {
		cfg.FileName = os.Getenv("FILE_STORAGE_PATH")
	}
	log.Println("File Name = ",cfg.FileName)
	// Валидация
	if _, err := url.ParseRequestURI(cfg.BaseURL); err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	return &cfg, nil
}
