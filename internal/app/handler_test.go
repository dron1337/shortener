package app

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func TestPostShortenURL(t *testing.T) {
	// Настраиваем тестовую конфигурацию
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	handler := setupTestServer(cfg)
	testURL := "https://practicum.yandex.ru/"

	// Создаем тестовый запрос
	req, err := http.NewRequest("POST", "/", strings.NewReader(testURL))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Проверяем Content-Type
	if contentType := rr.Header().Get("Content-Type"); contentType != "text/plain" {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, "text/plain")
	}

	// Проверяем что вернулся сокращенный URL с правильным базовым адресом
	shortURL := rr.Body.String()
	if !strings.HasPrefix(shortURL, cfg.BaseURL+"/") {
		t.Errorf("handler returned unexpected body: got %v, should start with %v/",
			shortURL, cfg.BaseURL)
	}
}

func TestGetRedirectURL(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}
	handler := setupTestServer(cfg)
	testURL := "https://practicum.yandex.ru/"

	// Сначала создаем сокращенную ссылку
	rrPost := httptest.NewRecorder()
	reqPost, _ := http.NewRequest("POST", "/", strings.NewReader(testURL))
	reqPost.Header.Set("Content-Type", "text/plain")
	handler.ServeHTTP(rrPost, reqPost)

	shortURL := rrPost.Body.String()
	shortID := strings.TrimPrefix(shortURL, cfg.BaseURL+"/")

	// Теперь тестируем редирект
	reqGet, _ := http.NewRequest("GET", "/"+shortID, nil)
	rrGet := httptest.NewRecorder()
	handler.ServeHTTP(rrGet, reqGet)

	// Проверки
	if status := rrGet.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}

	if location := rrGet.Header().Get("Location"); location != testURL {
		t.Errorf("handler returned wrong Location header: got %v want %v",
			location, testURL)
	}
}

func TestInvalidRequests(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}
	handler := setupTestServer(cfg)

	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		contentType string
		wantStatus  int
	}{
		{
			name:       "GET to root",
			method:     "GET",
			path:       "/",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:        "Empty body",
			method:      "POST",
			path:        "/",
			body:        "",
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Invalid URL",
			method:      "POST",
			path:        "/",
			body:        "not-a-valid-url",
			contentType: "text/plain",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:       "Non-existent short URL",
			method:     "GET",
			path:       "/nonexistent",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("%s: handler returned wrong status code: got %v want %v",
					tt.name, status, tt.wantStatus)
			}
		})
	}
}

func TestGenerateJSONURL(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8888",
		BaseURL:       "http://test.example",
	}
	handler := setupTestServer(cfg)

	tests := []struct {
		name           string
		requestBody    string
		contentType    string
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid JSON request",
			requestBody:    `{"url":"https://practicum.yandex.ru/"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				if contentType := rr.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("expected content type application/json, got %s", contentType)
				}

				var resp ResponseData
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}

				if !strings.HasPrefix(resp.Result, cfg.BaseURL+"/") {
					t.Errorf("expected result to start with %s/, got %s", cfg.BaseURL, resp.Result)
				}
			},
		},
		{
			name:           "Empty URL",
			requestBody:    `{"url":""}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url":"https://practicum.yandex.ru/"`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing URL field",
			requestBody:    `{"not_url":"https://practicum.yandex.ru/"}`,
			contentType:    "application/json",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten", strings.NewReader(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", tt.contentType)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}
		})
	}
}

func TestCheckDBConnection(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
		DBConnection:  "postgres://user:pass@localhost/db", // Тестовая строка подключения
	}

	handler := setupTestServer(cfg)

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Так как мы не можем реально подключиться к БД в тестах,
	// ожидаем 500 ошибку (или можно мокать DB соединение)
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusInternalServerError)
	}
}

// setupTestServer создает тестовый сервер с нужной конфигурацией
func setupTestServer(cfg *config.Config) http.Handler {
	store := store.New()
	var db *sql.DB
	if cfg.DBConnection != "" {
		// В тестах мы не создаем реальное подключение к БД
		db = nil
	}

	r := mux.NewRouter()
	handler := NewURLHandler(store, cfg, db)
	r.HandleFunc("/ping", handler.CheckDBConnection).Methods("GET")
	r.HandleFunc("/{key}", handler.GetURL).Methods("GET")
	r.HandleFunc("/", handler.GenerateURL).Methods("POST")
	r.HandleFunc("/api/shorten", handler.GenerateJSONURL).Methods("POST")

	return r
}
