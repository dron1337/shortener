package store

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/dron1337/shortener/internal/service"
)

type URLStorage struct {
	data map[string]string
}
type Storage struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var (
	counter int
)

func NewStorage(shortURL, originalURL string) Storage {
	counter++
	uuid := counter
	return Storage{
		UUID:        uuid,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}
func New() *URLStorage {
	return &URLStorage{
		data: make(map[string]string),
	}
}
func (s *URLStorage) Save(originalURL string, fileName string) string {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	shortKey := service.GenerateShortKey(rand)
	s.data[shortKey] = originalURL
	if fileName != "" {
		SaveInFile(NewStorage(shortKey, originalURL), fileName)
	}
	return shortKey

}
func SaveInFile(s Storage, fileName string) {
	if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		log.Printf("Failed to create directory: %v", err)
		return
	}
	file, err := os.OpenFile("."+fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("SaveInFile: failed to open file %s: %v", fileName, err)
		return
	}
	defer file.Close()
	data, err := json.Marshal(&s)
	if err != nil {
		log.Printf("SaveInFile: json marshal error: %v", err)
		return
	}
	log.Printf("Trying to save to file: %s", fileName)
	if _, err := os.Stat(fileName); err == nil {
		log.Printf("File already exists")
	} else if os.IsNotExist(err) {
		log.Printf("File does not exist")
	} else {
		log.Printf("File stat error: %v", err)
	}
	data = append(data, '\n')
	_, _ = file.Write(data)
}
func (s *URLStorage) Get(shortKey string) (string, bool) {
	url, exists := s.data[shortKey]
	return url, exists
}
