package store

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type FileStorage struct {
	mu       sync.Mutex
	filePath string
}

func NewFileStorage(filePath string) *FileStorage {
	return &FileStorage{
		filePath: filePath,
	}
}

func (s *FileStorage) Save(ctx context.Context, originalURL, shortKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	record := struct {
		ShortKey    string `json:"short_key"`
		OriginalURL string `json:"original_url"`
	}{
		ShortKey:    shortKey,
		OriginalURL: originalURL,
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(s.filePath), 0755); err != nil {
		return err
	}

	file, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data = append(data, '\n')
	if _, err := file.Write(data); err != nil {
		return err
	}

	return nil
}

func (s *FileStorage) GetOriginalURL(ctx context.Context, shortKey string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	file, err := os.Open(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file IsNotExist")
		}
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		// Проверяем контекст на каждой итерации
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		var record struct {
			ShortKey    string `json:"short_key"`
			OriginalURL string `json:"original_url"`
		}

		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			continue // Пропускаем некорректные записи
		}

		if record.ShortKey == shortKey {
			return record.OriginalURL, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("scanner error: %w", err)
	}

	return "", fmt.Errorf("URL not found")
}
func (s *FileStorage) GetShortKey(ctx context.Context, originalURL string) string {
	return ""
}
