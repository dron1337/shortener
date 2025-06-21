package app

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dron1337/shortener/internal/config"
	"github.com/dron1337/shortener/internal/store"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/mock"
)

// MockStorage реализует URLStorage для тестов
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(ctx context.Context, originalURL, shortKey string) error {
	args := m.Called(ctx, originalURL, shortKey)
	return args.Error(0)
}

func (m *MockStorage) Get(ctx context.Context, shortKey string) (string, error) {
	args := m.Called(ctx, shortKey)
	return args.String(0), args.Error(1)
}
func TestGetURLHandler_RealStorage(t *testing.T) {
	// 1. Подготовка тестовых данных
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	// 2. Создаем реальное хранилище в памяти
	storage := store.NewInMemoryStorage()

	// 3. Добавляем тестовые данные в хранилище
	testKey := "test123"
	testURL := "https://example.com"
	err := storage.Save(context.Background(), testURL, testKey)
	if err != nil {
		t.Fatalf("Failed to save test data: %v", err)
	}

	// 4. Создаем хендлер с реальным хранилищем
	handler := NewURLHandler(cfg, storage, log.Default())

	// 5. Создаем тестовый роутер
	router := mux.NewRouter()
	router.HandleFunc("/{key}", handler.GetURL).Methods("GET")

	// 6. Тест кейс 1: Успешное получение URL
	t.Run("Successful redirect", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/"+testKey, nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusTemporaryRedirect {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusTemporaryRedirect)
		}

		if location := rr.Header().Get("Location"); location != testURL {
			t.Errorf("handler returned wrong Location header: got %v want %v",
				location, testURL)
		}
	})

	// 7. Тест кейс 2: Несуществующий ключ
	t.Run("Non-existent key", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/nonexistent", nil)
		rr := httptest.NewRecorder()

		router.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code for non-existent key: got %v want %v",
				status, http.StatusBadRequest)
		}
	})
}

/*
func TestURLHandler(t *testing.T) {
	// Общая тестовая конфигурация
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}
	/*
		t.Run("TestGenerateURL", func(t *testing.T) {
			mockStore := new(MockStorage)
			mockStore.On("Save", mock.Anything, "https://example.com", mock.AnythingOfType("string")).Return(nil)

			handler := NewURLHandler(cfg, mockStore, log.Default())
			req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
			req.Header.Set("Content-Type", "text/plain")
			rr := httptest.NewRecorder()

			handler.GenerateURL(rr, req)

			assert.Equal(t, http.StatusCreated, rr.Code)
			assert.Contains(t, rr.Body.String(), cfg.BaseURL)
			mockStore.AssertExpectations(t)
			fmt.Println("TestGenerateURL")
		})

	t.Run("TestGetURL", func(t *testing.T) {
		mockStore := new(MockStorage)
		mockStore.On("Get", mock.Anything, "abc123").Return("https://example.com", nil)

		handler := NewURLHandler(cfg, mockStore, log.Default())
		req := httptest.NewRequest("GET", "/abc123", nil)
		rr := httptest.NewRecorder()

		handler.GetURL(rr, req)

		assert.Equal(t, http.StatusTemporaryRedirect, rr.Code)
		assert.Equal(t, "https://example.com", rr.Header().Get("Location"))
		mockStore.AssertExpectations(t)
		fmt.Println("TestGetURL")
	})
	/*
		t.Run("TestGenerateJSONURL", func(t *testing.T) {
			mockStore := new(MockStorage)
			mockStore.On("Save", mock.Anything, "https://example.com", mock.AnythingOfType("string")).Return(nil)

			handler := NewURLHandler(cfg, mockStore, log.Default())
			reqBody := `{"url":"https://example.com"}`
			req := httptest.NewRequest("POST", "/api/shorten", strings.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.GenerateJSONURL(rr, req)

			assert.Equal(t, http.StatusCreated, rr.Code)

			var resp ResponseData
			err := json.Unmarshal(rr.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Contains(t, resp.Result, cfg.BaseURL)
			mockStore.AssertExpectations(t)
			fmt.Println("TestGenerateJSONURL")
		})

}
*/
