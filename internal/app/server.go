package app

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/logger"
	"github.com/pluhe7/shortener/internal/storage"
)

type Server struct {
	Storage storage.Storage
	Config  *config.Config
	Echo    *echo.Echo
}

func NewServer(cfg *config.Config) *Server {
	s, err := storage.NewStorage(cfg.FileStoragePath, cfg.DatabaseDSN)
	if err != nil {
		logger.Log.Fatal("create new storage", zap.Error(err))
	}

	e := echo.New()

	server := &Server{
		Storage: s,
		Config:  cfg,
		Echo:    e,
	}

	return server
}

func (s *Server) Start() error {
	logger.Log.Info("Starting server...", zap.Object("config", s.Config))

	err := s.Echo.Start(s.Config.Address)
	if err != nil {
		return fmt.Errorf("echo start server: %w", err)
	}

	return nil
}

func (s *Server) Stop() {
	logger.Log.Info("Stopping server...")

	s.Storage.Close()

	logger.Log.Info("Server stopped")
}
