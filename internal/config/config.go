package config

import (
	"flag"
	"fmt"
	"net/url"
)

type Config struct {
	ServerAddress string
	BaseURL       string
}

func LoadConfig() (*Config, error) {
	var serverAddress string
	var baseURL string

	flag.StringVar(&serverAddress, "a", "localhost:8080", "HTTP server address")
	flag.StringVar(&baseURL, "b", "http://localhost:8080", "Base URL for shortened URLs")
	flag.Parse()

	_, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	return &Config{
		ServerAddress: serverAddress,
		BaseURL:       baseURL,
	}, nil
}
