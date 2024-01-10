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
	storage *storage.ShortURLStorage
	Config  *config.Config
	Echo    *echo.Echo
}

func NewServer(cfg *config.Config) *Server {
	s, err := storage.NewShortURLStorage(cfg.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("create new storage", zap.Error(err))
	}

	e := echo.New()

	server := &Server{
		storage: s,
		Config:  cfg,
		Echo:    e,
	}

	return server
}

func (s *Server) Start() error {
	err := s.Echo.Start(s.Config.Address)
	if err != nil {
		return fmt.Errorf("echo start server: %w", err)
	}

	return nil
}
