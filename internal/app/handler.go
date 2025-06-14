package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

type URLHandler struct {
	store  *store.URLStorage
	config *config.Config
}

func NewURLHandler(store *store.URLStorage, cfg *config.Config) *URLHandler {
	return &URLHandler{store: store, config: cfg}
}
func (h *URLHandler) GenerateURL(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	originalURL := strings.TrimSpace(string(body))
	log.Printf("Original URL: %s", originalURL)
	if originalURL == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL := h.store.Save(originalURL)
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	log.Printf("Short URL: %s", fullShortURL)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprint(len(fullShortURL)))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullShortURL))
}
func (h *URLHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Key: %s", key)
	url, ok := h.store.Get(key)
	log.Printf("Url: %s exists %t", key, ok)
	if !ok {
		log.Printf("Key not found: %s", key)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
