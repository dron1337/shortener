package store

import (
	"context"
	"database/sql"
	"fmt"
)

type CompositeStorage struct {
	storages []URLStorage
}
type URLStorage interface {
	Save(ctx context.Context, userId, originalURL, shortKey string) error
	GetOriginalURL(ctx context.Context, shortKey string) (string, error)
	GetShortKey(ctx context.Context, originalURL string) string
}

func NewCompositeStorage(storages ...URLStorage) *CompositeStorage {
	return &CompositeStorage{
		storages: storages,
	}
}

func (s *CompositeStorage) Save(ctx context.Context, userId, originalURL, shortKey string) error {
	for _, storage := range s.storages {
		if err := storage.Save(ctx, userId, originalURL, shortKey); err != nil {
			return err
		}
	}
	return nil
}

func (s *CompositeStorage) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	for _, storage := range s.storages {
		if url, err := storage.GetOriginalURL(ctx, shortKey); err == nil {
			return url, nil
		}
	}
	return "", fmt.Errorf("URL not found")
}
func (s *CompositeStorage) GetShortKey(ctx context.Context, originalURL string) string {
	shortKey := ""
	for _, storage := range s.storages {
		shortKey = storage.GetShortKey(ctx, originalURL)
		if shortKey != "" {
			return shortKey
		}

	}
	return shortKey
}
func (s *PostgresStorage) DB() *sql.DB {
	return s.db
}
func (s *CompositeStorage) GetStorages() []URLStorage {
	return s.storages
}
