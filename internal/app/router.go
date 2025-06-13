package app

import (
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func NewRouter(store *store.URLStorage) *mux.Router {
	r := mux.NewRouter()
	handler := NewURLHandler(store)

	r.HandleFunc("/{key}", handler.GetURL).Methods("GET")
	r.HandleFunc("/", handler.GenerateURL).Methods("POST")

	return r
}
