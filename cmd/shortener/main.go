package main

import (
	"log"

	"github.com/dron1337/shortener/internal/app"
)

func main() {
	logger := log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	server, err := app.NewServer(logger)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}
	if err := server.Start(); err != nil {
		server.Logger.Fatalf("Error starting server: %s", err)
	}
}
