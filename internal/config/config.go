package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server struct {
		Host         string        // Например, "" (для всех интерфейсов)
		Port         string        // ":8080"
		ReadTimeout  time.Duration // 5 * time.Second
		WriteTimeout time.Duration // 10 * time.Second
		IdleTimeout  time.Duration // 15 * time.Second
		ShutdownWait time.Duration // Таймаут graceful shutdown (1 * time.Second)
	}

	Storage struct {
		Filepath string // Путь к файлу для сохранения данных (если нужно)
	}

	Security struct {
		MaxURLLength      int           // Максимальная длина URL (для валидации)
		RateLimit         int           // Лимит запросов в секунду
		RateLimitInterval time.Duration // Интервал для rate-limiting
	}

	Paths struct {
		CreateURL string // "/"
		GetURL    string // "/{key}"
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Загружаем значения из переменных окружения или используем значения по умолчанию
	cfg.Server.Host = getEnv("SERVER_HOST", "")
	cfg.Server.Port = getEnv("SERVER_PORT", ":8080")
	cfg.Server.ReadTimeout = parseDuration(getEnv("SERVER_READ_TIMEOUT", "5s"))
	cfg.Server.WriteTimeout = parseDuration(getEnv("SERVER_WRITE_TIMEOUT", "10s"))
	cfg.Server.IdleTimeout = parseDuration(getEnv("SERVER_IDLE_TIMEOUT", "15s"))
	cfg.Server.ShutdownWait = parseDuration(getEnv("SERVER_SHUTDOWN_WAIT", "1s"))

	cfg.Storage.Filepath = getEnv("STORAGE_FILEPATH", "")

	cfg.Security.MaxURLLength = parseInt(getEnv("MAX_URL_LENGTH", "2048"))
	cfg.Security.RateLimit = parseInt(getEnv("RATE_LIMIT", "100"))
	cfg.Security.RateLimitInterval = parseDuration(getEnv("RATE_LIMIT_INTERVAL", "1s"))

	cfg.Paths.CreateURL = getEnv("PATH_CREATE_URL", "/")
	cfg.Paths.GetURL = getEnv("PATH_GET_URL", "/{key}")

	return cfg, nil
}

// Вспомогательные функции для парсинга переменных окружения
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Printf("Invalid duration format for value '%s', using default", value)
		return 0
	}
	return duration
}

func parseInt(value string) int {
	intValue, err := strconv.Atoi(value)
	if err != nil {
		log.Printf("Invalid integer format for value '%s', using default", value)
		return 0
	}
	return intValue
}
