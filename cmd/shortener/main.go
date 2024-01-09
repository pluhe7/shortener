package main

import (
	"log"

	"github.com/labstack/echo/v4"

	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/handlers"
	"github.com/pluhe7/shortener/internal/handlers/middleware"
	"github.com/pluhe7/shortener/internal/logger"
)

func main() {
	config.InitConfig()
	cfg := config.GetConfig()

	err := logger.Initialize(cfg.LogLevel)
	if err != nil {
		log.Fatal("init logger: " + err.Error())
	}

	e := echo.New()

	e.Use(middleware.RequestLogger, middleware.CompressorMiddleware)

	e.GET(`/:id`, handlers.ExpandHandler)
	e.POST(`/`, handlers.ShortenHandler)
	e.POST(`/api/shorten`, handlers.APIShortenHandler)

	e.Logger.Fatal(e.Start(cfg.Address))
}
