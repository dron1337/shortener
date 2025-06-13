package app

import (
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func NewRouter(cfg *config.Config, store *store.URLStorage) *mux.Router {
	urlHandler := NewURLHandler(store)

	router := mux.NewRouter()

	router.HandleFunc(cfg.Paths.CreateURL, urlHandler.GenerateURL).Methods("POST")
	router.HandleFunc(cfg.Paths.GetURL, urlHandler.GetURL).Methods("GET")

	return router

}
