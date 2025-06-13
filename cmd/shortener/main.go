package main

import (
	"log"
	"os"

	"github.com/dron1337/shortener/internal/app"
	"github.com/dron1337/shortener/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := log.New(os.Stdout, "SERVER: ", log.LstdFlags)
	server := app.NewServer(cfg, logger)

	if err := server.Start(); err != nil {
		logger.Fatal(err)
	}
}
