package store

import (
	"context"
	"database/sql"
)

type URLRepository interface {
	SaveURL(ctx context.Context, originalURL string) (string, error)
	GetURL(ctx context.Context, shortKey string) (string, error)
}

type PostgresURLRepository struct {
	db *sql.DB
}

func NewPostgresURLRepository(db *sql.DB) *PostgresURLRepository {
	return &PostgresURLRepository{db: db}
}

func (r *PostgresURLRepository) SaveURL(ctx context.Context, originalURL string) (string, error) {
	return "", nil
}

func (r *PostgresURLRepository) GetURL(ctx context.Context, shortKey string) (string, error) {
	return "", nil
}
