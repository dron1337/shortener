package store

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dron1337/shortener/internal/errors"
	"github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

func (s *PostgresStorage) Save(ctx context.Context, userID, originalURL, shortKey string) error {
	fmt.Println("Save PostgresStorage")
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx,
		"INSERT INTO short_urls (original_url, short_key,user_id) VALUES ($1, $2, $3)",
		originalURL, shortKey, userID)
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *PostgresStorage) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	var originalURL string
	var isDeleted bool
	err := s.db.QueryRowContext(ctx,
		"SELECT original_url,is_deleted FROM short_urls WHERE short_key = $1", shortKey).
		Scan(&originalURL, &isDeleted)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.ErrURLNotFound
		}
		return "", fmt.Errorf("db get error: %w", err)
	}
	if isDeleted {
		return "", errors.ErrURLDeleted
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
			user_id CHAR(36) NOT NULL,
			original_url TEXT NOT NULL,
			short_key VARCHAR(10) UNIQUE NOT NULL,
			is_deleted BOOLEAN DEFAULT FALSE
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
func (s *PostgresStorage) DeleteUserURLs(ctx context.Context, userID string, urls []string) {
	var wg sync.WaitGroup
	chunks := splitURLs(urls, 4)
	errCh := make(chan error, len(chunks))
	for _, batch := range chunks {
		wg.Add(1)
		go func(batch []string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			default:
				errCh <- s.updateDeleteUserURLs(userID, batch)
			}
		}(batch)
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()
	for err := range errCh {
		if err != nil {
			log.Printf("Ошибка при удалении: %v", err)
		}
	}
}
func splitURLs(urls []string, n int) [][]string {
	var chunks [][]string
	chunkSize := (len(urls) + n - 1) / n

	for i := 0; i < len(urls); i += chunkSize {
		end := min(i+chunkSize, len(urls))
		chunks = append(chunks, urls[i:end])
	}

	return chunks
}
func (s *PostgresStorage) updateDeleteUserURLs(userID string, batch []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	query := "UPDATE short_urls SET is_deleted = true WHERE short_key = ANY($1) AND user_id = $2 AND is_deleted = false"
	_, err = tx.Exec(query, pq.Array(batch), userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
