package main

import (
	"log"

	"github.com/dron1337/shortener/internal/app"
	"github.com/dron1337/shortener/internal/config"
)

func main() {
	logger := log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	server := app.NewServer(logger, cfg)
	if err := server.Start(); err != nil {
		server.Logger.Fatalf("Error starting server: %s", err)
	}
}
