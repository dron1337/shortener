package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/dron1337/shortener/internal/store"
)

type URLHandler struct {
	store *store.URLStorage
}

func NewURLHandler(store *store.URLStorage) *URLHandler {
	return &URLHandler{store: store}
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
	if !strings.HasPrefix(originalURL, "http://") && !strings.HasPrefix(originalURL, "https://") {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL := h.store.Save(originalURL)
	log.Printf("Short URL: %s", shortURL)
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", fmt.Sprint(len(shortURL)))
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
func (h *URLHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request: %s %s", r.Method, r.URL.Path)
	if r.URL.Path == "/" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	key := r.URL.Path[1:]
	if key == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("Key: %s", key)
	url, ok := h.store.Get(key)
	log.Printf("Url: %s exists %t", key, ok)
	if !ok {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
