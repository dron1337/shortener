package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/dron1337/shortener/internal/errors"
)

type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string]map[string]string
}
type ResponseURLs struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]map[string]string),
	}
}

func (s *InMemoryStorage) Save(ctx context.Context, userID, originalURL, shortKey string) error {
	fmt.Println("Save InMemoryStorage")
	s.mu.Lock()
	defer s.mu.Unlock()
	if user, exists := s.data[userID]; exists {
		user[shortKey] = originalURL
	} else {
		s.data[userID] = make(map[string]string)
		s.data[userID][shortKey] = originalURL
	}
	return nil
}

func (s *InMemoryStorage) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, properties := range s.data {
		if url, exists := properties[shortKey]; exists {
			return url, nil
		}
	}
	return "", errors.ErrURLNotFound
}
func (s *InMemoryStorage) GetShortKey(ctx context.Context, originalURL string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, properties := range s.data {
		fmt.Println(properties)
		for key, val := range properties {
			if val == originalURL {
				return key
			}
		}
	}
	return ""
}

func (s *InMemoryStorage) GetURLsByUser(ctx context.Context, userID, baseURL string) []ResponseURLs {
	var result []ResponseURLs
	userData, exists := s.data[userID]
	if !exists {
		return result
	}

	for shortKey, originalURL := range userData {
		result = append(result, ResponseURLs{
			OriginalURL: originalURL,
			ShortURL:    fmt.Sprintf("%s/%s", baseURL, shortKey),
		})
	}

	return result
}
