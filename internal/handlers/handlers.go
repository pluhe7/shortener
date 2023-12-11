package handlers

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/pluhe7/shortener/internal/app"
	"io"
	"net/http"
)

func ExpandHandler(c echo.Context) error {
	id := c.Param("id")

	expandedURL, err := app.ExpandURL(id)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("expand url error: %w", err).Error())
	}

	return c.Redirect(http.StatusTemporaryRedirect, expandedURL)
}

func ShortenHandler(c echo.Context) error {
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("read request body error: %w", err).Error())
	}

	shortURL, err := app.ShortenURL(string(bodyBytes))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("shorten url error: %w", err).Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)

	return c.String(http.StatusCreated, shortURL)
}
