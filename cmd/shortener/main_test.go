package main

/*
import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/dron1337/shortener/internal/app"
	"github.com/dron1337/shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	cfg := &config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://test.example",
	}

	t.Run("TestServerStartStop", func(t *testing.T) {
		//mockStore := new(MockStorage)
		logger := log.Default()

		server, err := app.NewServer(logger, cfg, nil)
		assert.NoError(t, err)

		// Тестируем graceful shutdown
		done := make(chan struct{})
		go func() {
			err := server.Start()
			assert.ErrorIs(t, err, http.ErrServerClosed)
			close(done)
		}()

		err = server.Stop()
		assert.NoError(t, err)
		<-done
	})
}
*/
