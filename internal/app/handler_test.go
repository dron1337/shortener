package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dron1337/shortener/internal/auth"
	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestGetURLHandler_RealStorage(t *testing.T) {
	// 1. Инициализация
	cfg := &config.Config{
		BaseURL: "http://test.example",
	}

	// Создаем реальное in-memory хранилище
	storages := Storages{Memory: store.NewInMemoryStorage()}
	handler := NewURLHandler(cfg, &storages, log.Default())
	router := mux.NewRouter()
	router.HandleFunc("/{key}", handler.GetURL).Methods("GET")

	// 2. Подготовка данных - сохраним тестовый URL
	testURL := "https://example.com"
	userID := "test-user"
	err := storages.Memory.Save(context.Background(), userID, testURL, "abc123")
	assert.NoError(t, err)

	// 3. Тест успешного редиректа
	t.Run("Successful redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/"+"abc123", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		assert.Equal(t, testURL, rr.Header().Get("Location"))
	})

	// 4. Тест несуществующего URL
	t.Run("Non-existent key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGenerateURLHandler_RealStorage(t *testing.T) {
	cfg := &config.Config{
		BaseURL: "http://test.example",
	}

	storages := Storages{Memory: store.NewInMemoryStorage()}
	handler := NewURLHandler(cfg, &storages, log.Default())

	t.Run("Successful URL generation", func(t *testing.T) {
		testURL := "https://example.com"
		req := httptest.NewRequest("POST", "/", strings.NewReader(testURL))
		req = req.WithContext(context.WithValue(req.Context(), auth.UserIDKey, "test-user"))
		rr := httptest.NewRecorder()

		handler.GenerateURL(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Body.String(), cfg.BaseURL)
	})

	t.Run("Duplicate URL", func(t *testing.T) {
		testURL := "https://example.com"

		// Первый запрос - создаем URL
		req1 := httptest.NewRequest("POST", "/", strings.NewReader(testURL))
		req1 = req1.WithContext(context.WithValue(req1.Context(), auth.UserIDKey, "test-user"))
		rr1 := httptest.NewRecorder()
		handler.GenerateURL(rr1, req1)

		// Второй запрос - должен вернуть конфликт
		req2 := httptest.NewRequest("POST", "/", strings.NewReader(testURL))
		req2 = req2.WithContext(context.WithValue(req2.Context(), auth.UserIDKey, "test-user"))
		rr2 := httptest.NewRecorder()
		handler.GenerateURL(rr2, req2)

		assert.Equal(t, http.StatusConflict, rr2.Code)
		assert.Contains(t, rr2.Body.String(), cfg.BaseURL)
	})
}
