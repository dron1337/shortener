package app

import (
	"log"

	"github.com/dron1337/shortener/internal/auth"
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/logger"
	"github.com/dron1337/shortener/internal/service"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func NewRouter(cfg *config.Config, urlStore store.URLStorage, log *log.Logger) *mux.Router {
	r := mux.NewRouter()

	if err := logger.Initialize("info"); err != nil {
		panic(err)
	}

	r.Use(logger.LoggingMiddleware)
	r.Use(service.GzipHandle)
	r.Use(auth.AuthMiddleware)
	handler := NewURLHandler(cfg, urlStore, log)
	r.HandleFunc("/ping", handler.CheckDBConnection).Methods("GET")
	r.HandleFunc("/{key}", handler.GetURL).Methods("GET")
	r.HandleFunc("/api/user/urls", handler.GetUserURLs).Methods("GET")
	r.HandleFunc("/", handler.GenerateURL).Methods("POST")
	r.HandleFunc("/api/shorten", handler.GenerateJSONURL).Methods("POST")
	r.HandleFunc("/api/shorten/batch", handler.GenerateBatchJSONURL).Methods("POST")
	return r
}
