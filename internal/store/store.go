package store

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/dron1337/shortener/internal/service"
)

type URLStorage struct {
	data map[string]string
}

func New() *URLStorage {
	return &URLStorage{
		data: make(map[string]string),
	}
}
func (s *URLStorage) Save(originalURL string) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	shortKey := service.GenerateShortKey(rand)
	s.data[shortKey] = originalURL
	return fmt.Sprintf("http://localhost:8080/%s", shortKey)
}

func (s *URLStorage) Get(shortKey string) (string, bool) {
	url, exists := s.data[shortKey]
	return url, exists
}
