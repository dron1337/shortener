package app
/*
import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dron1337/shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	t.Run("TestValidRoutes", func(t *testing.T) {
		mockStore := new(MockStorage)
		router := NewRouter(cfg, mockStore, log.Default())

		tests := []struct {
			method       string
			path         string
			expectedCode int
		}{
			{"POST", "/", http.StatusCreated},
			{"GET", "/abc123", http.StatusTemporaryRedirect},
			{"POST", "/api/shorten", http.StatusCreated},
			{"GET", "/ping", http.StatusOK},
		}

		for _, tt := range tests {
			t.Run(tt.method+" "+tt.path, func(t *testing.T) {
				var req *http.Request
				if tt.method == "POST" {
					if tt.path == "/api/shorten" {
						req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"url":"http://example.com"}`))
						req.Header.Set("Content-Type", "application/json")
					} else {
						req = httptest.NewRequest(tt.method, tt.path, strings.NewReader("http://example.com"))
						req.Header.Set("Content-Type", "text/plain")
					}
				} else {
					req = httptest.NewRequest(tt.method, tt.path, nil)
				}

				rr := httptest.NewRecorder()
				router.ServeHTTP(rr, req)

				assert.Equal(t, tt.expectedCode, rr.Code)
			})
		}
	})
}
*/