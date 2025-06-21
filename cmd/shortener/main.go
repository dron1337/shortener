package main

import (
	"log"

	"github.com/dron1337/shortener/internal/app"
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
)

func main() {
	logger := log.New(log.Writer(), "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	storages := []store.URLStorage{store.NewInMemoryStorage()}
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
		cfg = &config.Config{}
	}
	if cfg.FileName != "" {
		storages = append(storages, store.NewFileStorage(cfg.FileName))
	}

	if cfg.DBConnection != "" {
		db, err := store.CreateDBConnection(cfg.DBConnection)
		if err != nil {
			logger.Printf("WARNING: DB storage disabled: %v", err)
		} else {
			if err := db.Ping(); err != nil {
				logger.Printf("WARNING: DB connection failed: %v", err)
			} else {
				storages = append(storages, store.NewPostgresStorage(db))
			}
		}
	}
	// Создаем композитное хранилище
	urlStore := store.NewCompositeStorage(storages...)
	server, err := app.NewServer(logger, cfg, urlStore)
	if err != nil {
		logger.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		server.Logger.Fatalf("Error starting server: %s", err)
	}
}
