package app

import (
	"database/sql"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/logger"
	"github.com/dron1337/shortener/internal/service"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func NewRouter(cfg *config.Config, db *sql.DB, store *store.URLStorage) *mux.Router {
	r := mux.NewRouter()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	r.Use(logger.LoggingMiddleware)
	r.Use(service.GzipHandle)

	handler := NewURLHandler(store, cfg, db)
	r.HandleFunc("/ping", handler.CheckDBConnection).Methods("GET")
	r.HandleFunc("/{key}", handler.GetURL).Methods("GET")
	r.HandleFunc("/", handler.GenerateURL).Methods("POST")
	r.HandleFunc("/api/shorten", handler.GenerateJSONURL).Methods("POST")

	return r
}
