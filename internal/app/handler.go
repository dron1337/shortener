package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/service"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

type URLHandler struct {
	store  store.URLStorage
	logger *log.Logger
	config *config.Config
}
type RequestData struct {
	URL string `json:"url"`
}
type ResponseData struct {
	Result string `json:"result"`
}

func NewURLHandler(cfg *config.Config, urlStore store.URLStorage, logger *log.Logger) *URLHandler {
	return &URLHandler{config: cfg, store: urlStore, logger: logger}
}
func (h *URLHandler) GenerateURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("Incoming request: %s %s, Headers: %v", r.Method, r.URL, r.Header)
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
	shortURL := service.GenerateShortKey()
	err = h.store.Save(r.Context(), originalURL, shortURL)
	if err != nil {
		h.logger.Println("error store save:", err)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	h.logger.Printf("Short URL: %s", fullShortURL)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullShortURL))
}
func (h *URLHandler) GenerateJSONURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	h.logger.Printf("Raw body: %q", body)
	var data RequestData
	if err := json.Unmarshal(body, &data); err != nil {
		h.logger.Println("error parse json:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.logger.Println("URL:", data.URL)
	if data.URL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(data.URL); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortURL := service.GenerateShortKey()
	err = h.store.Save(r.Context(), data.URL, shortURL)
	if err != nil {
		h.logger.Println("error store save:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	response := ResponseData{Result: fullShortURL}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("error parse response json:", err)

	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	h.logger.Printf("Sending response: %s", jsonBytes)
	w.Write(jsonBytes)
}
func (h *URLHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("Request: %s %s", r.Method, r.URL.Path)
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.logger.Printf("Key: %s", key)
	url, err := h.store.Get(r.Context(), key)
	h.logger.Println(url)
	if err != nil {
		h.logger.Printf("Key not found: %s", key)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
func (h *URLHandler) CheckDBConnection(w http.ResponseWriter, r *http.Request) {
	if pgStorage, ok := h.store.(*store.PostgresStorage); ok {
		err := pgStorage.CheckConnection(r.Context())
		if err != nil {
			h.logger.Printf("Postgres connection check failed: %v", err)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else if composite, ok := h.store.(*store.CompositeStorage); ok {
		// Это композитное хранилище, ищем Postgres внутри
		hasPostgres := false
		for _, s := range composite.GetStorages() {
			if pg, ok := s.(*store.PostgresStorage); ok {
				hasPostgres = true
				if err := pg.CheckConnection(r.Context()); err != nil {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}
		if !hasPostgres {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	} else {
		// Другой тип хранилища
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
