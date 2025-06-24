package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Save(ctx context.Context, userId, originalURL, shortKey string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx,
		"INSERT INTO short_urls (original_url, short_key) VALUES ($1, $2)",
		originalURL, shortKey)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *PostgresStorage) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	var originalURL string
	err := s.db.QueryRowContext(ctx,
		"SELECT original_url FROM short_urls WHERE short_key = $1", shortKey).
		Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("URL not found")
		}
		return "", fmt.Errorf("db get error: %w", err)
	}
	return originalURL, nil
}
func (s *PostgresStorage) GetShortKey(ctx context.Context, originalURL string) string {
	var existingShortKey string
	s.db.QueryRowContext(ctx,
		"SELECT short_key FROM short_urls WHERE original_url = $1", originalURL).Scan(&existingShortKey)
	return existingShortKey
}
func CreateDBConnection(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening DB connection: %w", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging DB: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS short_urls (
			uuid SERIAL PRIMARY KEY,
			original_url TEXT NOT NULL,
			short_key VARCHAR(10) UNIQUE NOT NULL
	);
`)
	if err != nil {
		return nil, err
	}
	return db, nil
}
func (s *PostgresStorage) CheckConnection(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
