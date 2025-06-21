package store

import (
	"context"
	"fmt"
	"sync"
)

type InMemoryStorage struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) Save(ctx context.Context, originalURL, shortKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[shortKey] = originalURL
	return nil
}

func (s *InMemoryStorage) Get(ctx context.Context, shortKey string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	fmt.Println(s.data)
	if url, exists := s.data[shortKey]; exists {
		return url, nil
	}
	return "", fmt.Errorf("URL not found")
}
