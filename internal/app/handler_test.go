package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStorage реализует URLStorage для тестов
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(ctx context.Context, userId, originalURL, shortKey string) error {
	args := m.Called(ctx, userId, originalURL, shortKey)
	return args.Error(0)
}

func (m *MockStorage) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	args := m.Called(ctx, shortKey)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetShortKey(ctx context.Context, originalURL string) string {
	args := m.Called(ctx, originalURL)
	return args.String(0)
}

func TestGetURLHandler(t *testing.T) {
	// 1. Подготовка тестовых данных
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	// 2. Создаем mock хранилище
	mockStore := new(MockStorage)

	// 3. Создаем хендлер с mock хранилищем
	handler := NewURLHandler(cfg, mockStore, log.Default())

	// 4. Создаем тестовый роутер
	router := mux.NewRouter()
	router.HandleFunc("/{key}", handler.GetURL).Methods("GET")

	// 5. Тест кейс 1: Успешное получение URL
	t.Run("Successful redirect", func(t *testing.T) {
		testKey := "test123"
		testURL := "https://example.com"

		// Настраиваем mock
		mockStore.On("GetOriginalURL", mock.Anything, testKey).Return(testURL, nil)

		req := httptest.NewRequest("GET", "/"+testKey, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		assert.Equal(t, testURL, rr.Header().Get("Location"))
		mockStore.AssertExpectations(t)
	})

	// 6. Тест кейс 2: Несуществующий ключ
	t.Run("Non-existent key", func(t *testing.T) {
		testKey := "nonexistent"

		// Настраиваем mock
		mockStore.On("GetOriginalURL", mock.Anything, testKey).Return("", store.ErrURLNotFound)

		req := httptest.NewRequest("GET", "/"+testKey, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockStore.AssertExpectations(t)
	})
}

func TestGenerateURLHandler(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	t.Run("Successful URL generation", func(t *testing.T) {
		mockStore := new(MockStorage)
		testURL := "https://example.com"
		//testKey := "abc123"

		// Настраиваем mock
		mockStore.On("GetShortKey", mock.Anything, testURL).Return("")
		mockStore.On("Save", mock.Anything, mock.Anything, testURL, mock.AnythingOfType("string")).Return(nil)

		handler := NewURLHandler(cfg, mockStore, log.Default())

		req := httptest.NewRequest("POST", "/", strings.NewReader(testURL))
		req = req.WithContext(context.WithValue(req.Context(), "userID", "test-user"))
		rr := httptest.NewRecorder()

		handler.GenerateURL(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		assert.Contains(t, rr.Body.String(), cfg.BaseURL)
		mockStore.AssertExpectations(t)
	})

	t.Run("Duplicate URL", func(t *testing.T) {
		mockStore := new(MockStorage)
		testURL := "https://example.com"
		testKey := "abc123"

		// Настраиваем mock
		mockStore.On("GetShortKey", mock.Anything, testURL).Return(testKey)

		handler := NewURLHandler(cfg, mockStore, log.Default())

		req := httptest.NewRequest("POST", "/", strings.NewReader(testURL))
		req = req.WithContext(context.WithValue(req.Context(), "userID", "test-user"))
		rr := httptest.NewRecorder()

		handler.GenerateURL(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), cfg.BaseURL+"/"+testKey)
		mockStore.AssertExpectations(t)
	})
}

func TestGenerateJSONURLHandler(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	t.Run("Successful JSON URL generation", func(t *testing.T) {
		mockStore := new(MockStorage)
		testURL := "https://example.com"
		//testKey := "abc123"
		requestBody := `{"url":"` + testURL + `"}`

		// Настраиваем mock
		mockStore.On("GetShortKey", mock.Anything, testURL).Return("")
		mockStore.On("Save", mock.Anything, mock.Anything, testURL, mock.AnythingOfType("string")).Return(nil)

		handler := NewURLHandler(cfg, mockStore, log.Default())

		req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(context.WithValue(req.Context(), "userID", "test-user"))
		rr := httptest.NewRecorder()

		handler.GenerateJSONURL(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)

		var resp ResponseData
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Result, cfg.BaseURL)
		mockStore.AssertExpectations(t)
	})
}
