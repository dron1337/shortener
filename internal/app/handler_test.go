package app

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
)

func TestPostShortenURL(t *testing.T) {
	// Инициализируем роутер или хендлер
	handler := setupServer()

	// Создаем тестовый запрос
	body := strings.NewReader("https://practicum.yandex.ru/")
	req, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "text/plain")

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем хендлер напрямую
	handler.ServeHTTP(rr, req)

	// Проверяем статус код
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Проверяем Content-Type
	expectedContentType := "text/plain"
	if contentType := rr.Header().Get("Content-Type"); contentType != expectedContentType {
		t.Errorf("handler returned wrong content type: got %v want %v",
			contentType, expectedContentType)
	}

	// Проверяем что вернулся сокращенный URL
	shortURL := rr.Body.String()
	if !strings.HasPrefix(shortURL, "http://localhost:8080/") {
		t.Errorf("handler returned unexpected body: got %v, should start with http://localhost:8080/",
			shortURL)
	}
}

func TestGetRedirectURL(t *testing.T) {
	// Инициализируем роутер или хендлер
	handler := setupServer()

	// Сначала создаем сокращенную ссылку
	body := strings.NewReader("https://practicum.yandex.ru/")
	reqPost, err := http.NewRequest("POST", "/", body)
	if err != nil {
		t.Fatal(err)
	}
	reqPost.Header.Set("Content-Type", "text/plain")

	rrPost := httptest.NewRecorder()
	handler.ServeHTTP(rrPost, reqPost)
	shortURL := rrPost.Body.String()
	shortID := strings.TrimPrefix(shortURL, "http://localhost:8080/")

	// Теперь тестируем редирект
	reqGet, err := http.NewRequest("GET", "/"+shortID, nil)
	if err != nil {
		t.Fatal(err)
	}

	rrGet := httptest.NewRecorder()
	handler.ServeHTTP(rrGet, reqGet)

	// Проверяем статус код
	if status := rrGet.Code; status != http.StatusTemporaryRedirect {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusTemporaryRedirect)
	}

	// Проверяем Location header
	expectedLocation := "https://practicum.yandex.ru/"
	if location := rrGet.Header().Get("Location"); location != expectedLocation {
		t.Errorf("handler returned wrong Location header: got %v want %v",
			location, expectedLocation)
	}
}

func TestInvalidRequests(t *testing.T) {
	handler := setupServer()

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
			wantStatus: http.StatusBadRequest,
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
			var body *strings.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			} else {
				body = strings.NewReader("")
			}

			req, err := http.NewRequest(tt.method, tt.path, body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			t.Logf("Request: %s %s", tt.method, tt.path)
			t.Logf("Response status: %d", rr.Code)
			t.Logf("Response headers: %v", rr.Header())
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.wantStatus)
			}
		})
	}
}
func setupServer() http.Handler {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	store := store.New()
	handler := NewURLHandler(store)
	r := mux.NewRouter()

	// Регистрируем пути из конфига
	r.HandleFunc(cfg.Paths.GetURL, handler.GetURL).Methods("GET")
	r.HandleFunc(cfg.Paths.CreateURL, handler.GenerateURL).Methods("POST")

	// Обработчик для некорректных GET запросов на корень
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		http.NotFound(w, r)
	}).Methods("GET")

	return r
}
