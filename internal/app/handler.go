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
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

type URLHandler struct {
	store  *store.URLStorage
	config *config.Config
}
type RequestData struct {
	URL string `json:"url"`
}
type ResponseData struct {
	Result string `json:"result"`
}

func NewURLHandler(store *store.URLStorage, cfg *config.Config) *URLHandler {
	return &URLHandler{store: store, config: cfg}
}
func (h *URLHandler) GenerateURL(w http.ResponseWriter, r *http.Request) {
	log.Printf("Incoming request: %s %s, Headers: %v", r.Method, r.URL, r.Header)
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
	shortURL := h.store.Save(originalURL,h.config.FileName)
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	log.Printf("Short URL: %s", fullShortURL)
	w.Header().Set("Content-Type", "text/plain")
	//w.Header().Set("Content-Length", fmt.Sprint(len(fullShortURL)))
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
	log.Printf("Raw body: %q", body)
	var data RequestData
	if err := json.Unmarshal(body, &data); err != nil {
		log.Println("error parse json:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("URL:", data.URL)
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
	shortURL := h.store.Save(data.URL, h.config.FileName)
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	response := ResponseData{Result: fullShortURL}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		log.Println("error parse response json:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	//w.Header().Set("Content-Length", fmt.Sprint(len(fullShortURL)))
	//w.Header().Set("Content-Length", fmt.Sprint(len(jsonBytes)))
	w.WriteHeader(http.StatusCreated)
	log.Printf("Sending response: %s", jsonBytes)
	w.Write(jsonBytes)
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
