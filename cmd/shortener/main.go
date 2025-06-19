package main

import (
	"log"

	"github.com/dron1337/shortener/internal/app"
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
)

func main() {
	logger := log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	urlStore := store.New()

	server, err := app.NewServer(logger, cfg, urlStore)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		server.Logger.Fatalf("Error starting server: %s", err)
	}
}
