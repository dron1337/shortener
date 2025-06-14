package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dron1337/shortener/internal/config"
)

type Server struct {
	Logger     *log.Logger
	HTTPServer *http.Server
	Config     *config.Config
}

func (s *Server) Start() error {
	s.Logger.Println("Starting server on", s.HTTPServer.Addr)

	serverErr := make(chan error, 1)

	go func() {
		if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		s.Logger.Printf("Received signal: %v", sig)
		return s.Stop()
	case err := <-serverErr:
		return err
	}
}

func (s *Server) Stop() error {
	s.Logger.Println("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		s.Logger.Printf("Graceful shutdown failed: %v", err)
		return err
	}

	s.Logger.Println("Server stopped gracefully")
	return nil
}

func NewServer(logger *log.Logger, cfg *config.Config) *Server {
	mux := NewRouter(cfg)
	s := &http.Server{
		Addr:         cfg.ServerAddress,
		Handler:      mux,
		ErrorLog:     logger,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Server{
		Logger:     logger,
		HTTPServer: s,
		Config:     cfg,
	}
}
