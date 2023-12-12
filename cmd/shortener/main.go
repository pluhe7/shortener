package main

import (
	"github.com/labstack/echo/v4"
	"github.com/pluhe7/shortener/config"
	"github.com/pluhe7/shortener/internal/handlers"
)

func main() {
	config.InitConfig()

	cfg := config.GetConfig()
	cfg.ParseFlags()

	e := echo.New()

	e.GET(`/:id`, handlers.ExpandHandler)
	e.POST(`/`, handlers.ShortenHandler)

	e.Logger.Fatal(e.Start(cfg.Address))
}
