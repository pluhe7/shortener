package main

import (
	"github.com/labstack/echo/v4"
	"github.com/pluhe7/shortener/internal/handlers"
)

func main() {
	e := echo.New()

	e.GET(`/:id`, handlers.ExpandHandler)
	e.POST(`/`, handlers.ShortenHandler)

	e.Logger.Fatal(e.Start(":8080"))
}
