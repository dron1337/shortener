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
	Save(ctx context.Context, originalURL, shortKey string) error
	Get(ctx context.Context, shortKey string) (string, error)
}

func NewCompositeStorage(storages ...URLStorage) *CompositeStorage {
	return &CompositeStorage{
		storages: storages,
	}
}

func (s *CompositeStorage) Save(ctx context.Context, originalURL, shortKey string) error {
	for _, storage := range s.storages {
		fmt.Println("originalURL=", storage, originalURL)
		fmt.Println("shortKey=", storage, shortKey)
		if err := storage.Save(ctx, originalURL, shortKey); err != nil {
			return err
		}
	}
	return nil
}

func (s *CompositeStorage) Get(ctx context.Context, shortKey string) (string, error) {
	fmt.Println(s.storages)
	for _, storage := range s.storages {
		if url, err := storage.Get(ctx, shortKey); err == nil {
			fmt.Println("Возвращаю CompositeStorage url", url)
			return url, nil
		}
	}
	return "", fmt.Errorf("URL not found")
}
func (s *CompositeStorage) GetWorkingPostgres(ctx context.Context) (*PostgresStorage, error) {
	for _, storage := range s.storages {
		if pg, ok := storage.(*PostgresStorage); ok {
			if err := pg.DB().PingContext(ctx); err == nil {
				return pg, nil
			}
		}
	}
	return nil, fmt.Errorf("no working PostgresStorage available")
}

func (s *PostgresStorage) DB() *sql.DB {
	return s.db
}
func (s *CompositeStorage) GetStorages() []URLStorage {
	return s.storages
}
