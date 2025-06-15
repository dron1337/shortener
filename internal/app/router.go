package app

import (
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/logger"

	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func NewRouter(cfg *config.Config) *mux.Router {
	store := store.New()
	r := mux.NewRouter()
	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}
	r.Use(logger.LoggingMiddleware)
	handler := NewURLHandler(store, cfg)

	r.HandleFunc("/{key}", handler.GetURL).Methods("GET")
	r.HandleFunc("/", handler.GenerateURL).Methods("POST")
	r.HandleFunc("/api/shorten", handler.GenerateJSONURL).Methods("POST")

	return r
}
