package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/dron1337/shortener/internal/auth"
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/errors"
	"github.com/dron1337/shortener/internal/service"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

type URLHandler struct {
	storages *Storages
	logger   *log.Logger
	config   *config.Config
}
type RequestData struct {
	URL string `json:"url"`
}
type ResponseData struct {
	Result string `json:"result"`
}
type BatchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchRequest []BatchRequestItem
type BatchResponse []BatchResponseItem
type BatchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewURLHandler(cfg *config.Config, storages *Storages, logger *log.Logger) *URLHandler {
	return &URLHandler{config: cfg, storages: storages, logger: logger}
}
func (h *URLHandler) GenerateURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("Incoming request: %s %s, Headers: %v", r.Method, r.URL, r.Header)
	shortURL := ""
	userID := r.Context().Value(auth.UserIDKey).(string)
	w.Header().Set("Content-Type", "text/plain")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	originalURL := strings.TrimSpace(string(body))
	h.logger.Printf("Original URL: %s", originalURL)
	if originalURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(originalURL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if h.storages.Postgres != nil {
		shortURL = h.storages.Postgres.GetShortKey(r.Context(), originalURL)
	}

	if shortURL == "" && h.storages.FileStorage != nil {
		shortURL = h.storages.FileStorage.GetShortKey(r.Context(), originalURL)
	}

	if shortURL == "" && h.storages.Memory != nil {
		shortURL = h.storages.Memory.GetShortKey(r.Context(), originalURL)
	}
	status := http.StatusConflict
	if shortURL == "" {
		shortURL = service.GenerateShortKey()
		var saveErrors []error
		if h.storages.Postgres != nil {
			if err := h.storages.Postgres.Save(r.Context(), userID, originalURL, shortURL); err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("postgres save failed: %w", err))
				h.logger.Printf("Postgres save error: %v", err)
			}
		}
		if h.storages.FileStorage != nil {
			if err := h.storages.FileStorage.Save(r.Context(), userID, originalURL, shortURL); err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("file save failed: %w", err))
				h.logger.Printf("File storage save error: %v", err)
			}
		}
		if h.storages.Memory != nil {
			if err := h.storages.Memory.Save(r.Context(), userID, originalURL, shortURL); err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("memory save failed: %w", err))
				h.logger.Printf("Memory storage save error: %v", err)
			}
		}
		status = http.StatusCreated
		if len(saveErrors) > 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to save URL to one or more storage backends"))
			return
		}
	}
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	h.logger.Printf("Short URL: %s", fullShortURL)
	w.WriteHeader(status)
	w.Write([]byte(fullShortURL))
}
func (h *URLHandler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	h.logger.Printf("User ID из куки: %s", userID)
	w.Header().Set("Content-Type", "application/json")
	var urls []store.ResponseURLs
	if h.storages.Memory != nil {
		urls = h.storages.Memory.GetURLsByUser(r.Context(), userID, h.config.BaseURL)
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	json.NewEncoder(w).Encode(urls)
}
func (h *URLHandler) GenerateJSONURL(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	shortURL := ""
	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	h.logger.Printf("Raw body: %q", body)
	var data RequestData
	if err := json.Unmarshal(body, &data); err != nil {
		h.logger.Println("error parse json:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.logger.Println("URL:", data.URL)
	if data.URL == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if _, err := url.ParseRequestURI(data.URL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if h.storages.Postgres != nil {
		shortURL = h.storages.Postgres.GetShortKey(r.Context(), data.URL)
	}
	if shortURL == "" && h.storages.FileStorage != nil {
		shortURL = h.storages.FileStorage.GetShortKey(r.Context(), data.URL)
	}
	if shortURL == "" && h.storages.Memory != nil {
		shortURL = h.storages.Memory.GetShortKey(r.Context(), data.URL)
	}
	status := http.StatusConflict
	h.logger.Println("shortURL:", shortURL)
	if shortURL == "" {
		shortURL = service.GenerateShortKey()
		var saveErrors []error
		if h.storages.Postgres != nil {
			if err := h.storages.Postgres.Save(r.Context(), userID, data.URL, shortURL); err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("postgres save failed: %w", err))
				h.logger.Printf("Postgres save error: %v", err)
			}
		}
		if h.storages.FileStorage != nil {
			if err := h.storages.FileStorage.Save(r.Context(), userID, data.URL, shortURL); err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("file save failed: %w", err))
				h.logger.Printf("File storage save error: %v", err)
			}
		}
		if h.storages.Memory != nil {
			if err := h.storages.Memory.Save(r.Context(), userID, data.URL, shortURL); err != nil {
				saveErrors = append(saveErrors, fmt.Errorf("memory save failed: %w", err))
				h.logger.Printf("Memory storage save error: %v", err)
			}
		}
		status = http.StatusCreated
		if len(saveErrors) > 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to save URL to one or more storage backends"))
			return
		}
	}
	fullShortURL := fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL)
	response := ResponseData{Result: fullShortURL}
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		h.logger.Println("error parse response json:", err)
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	h.logger.Printf("Sending response: %s", jsonBytes)
	w.Write(jsonBytes)
}
func (h *URLHandler) GetURL(w http.ResponseWriter, r *http.Request) {
	h.logger.Printf("Request: %s %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "text/plain")
	vars := mux.Vars(r)
	key := vars["key"]
	if key == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.logger.Printf("Key: %s", key)
	var saveErrors []error
	var url string
	var found bool
	if h.storages.Postgres != nil {
		if u, err := h.storages.Postgres.GetOriginalURL(r.Context(), key); err != nil {
			if err == errors.ErrURLDeleted {
				h.logger.Printf("URL deleted in Postgres: %s", key)
				w.WriteHeader(http.StatusGone)
				return
			}
		} else {
			url = u
			found = true
		}
	}
	if !found && h.storages.FileStorage != nil {
		if u, err := h.storages.FileStorage.GetOriginalURL(r.Context(), key); err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("file get failed: %w", err))
			h.logger.Printf("File storage get error: %v", err)
		} else {
			url = u
			found = true
		}
	}
	if !found && h.storages.Memory != nil {
		if u, err := h.storages.Memory.GetOriginalURL(r.Context(), key); err != nil {
			saveErrors = append(saveErrors, fmt.Errorf("memory get failed: %w", err))
			h.logger.Printf("Memory storage get error: %v", err)
		} else {
			url = u
			found = true
		}
	}
	h.logger.Println("URL:", url)
	if len(saveErrors) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)
}
func (h *URLHandler) CheckDBConnection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	if err := h.storages.Postgres.CheckConnection(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
func (h *URLHandler) GenerateBatchJSONURL(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	w.Header().Set("Content-Type", "application/json")
	var batch BatchRequest
	if err := json.NewDecoder(r.Body).Decode(&batch); err != nil {
		h.logger.Println("Failed to decode batch request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if len(batch) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var response BatchResponse
	var shortURL string
	status := http.StatusConflict
	for _, item := range batch {
		shortURL = h.storages.Postgres.GetShortKey(r.Context(), item.OriginalURL)
		if shortURL == "" {
			shortURL = service.GenerateShortKey()
			err := h.storages.Postgres.Save(r.Context(), userID, item.OriginalURL, shortURL)
			status = http.StatusCreated
			if err != nil {
				h.logger.Printf("Unexpected save error: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		response = append(response, BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("%s/%s", h.config.BaseURL, shortURL),
		})
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Printf("Failed to encode response: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
func (h *URLHandler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(auth.UserIDKey).(string)
	var urls []string
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewDecoder(r.Body).Decode(&urls); err != nil {
		h.logger.Println("Failed to decode batch request:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if len(urls) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	h.storages.Postgres.DeleteUserURLs(r.Context(), userID, urls)
	w.WriteHeader(http.StatusAccepted)
}
