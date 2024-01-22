package main

import (
	"go.uber.org/zap"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/handlers"
	"github.com/pluhe7/shortener/internal/logger"
)

func main() {
	cfg := config.InitConfig()
	logger.InitLogger(cfg.LogLevel)

	server := app.NewServer(cfg)
	handlers.InitHandlers(server)

	err := server.Start()
	if err != nil {
		logger.Log.Fatal("start server", zap.Error(err))
	}
}
