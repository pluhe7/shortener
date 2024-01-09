package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/models"
)

func ExpandHandler(c echo.Context) error {
	id := c.Param("id")

	expandedURL, err := app.ExpandURL(id)
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, app.ErrURLNotFound) {
			status = http.StatusNotFound
		}

		return c.String(status, fmt.Errorf("expand url error: %w", err).Error())
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

func APIShortenHandler(c echo.Context) error {
	var req models.ShortenRequest

	requestDecoder := json.NewDecoder(c.Request().Body)
	err := requestDecoder.Decode(&req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("decode request error: %w", err).Error())
	}

	shortURL, err := app.ShortenURL(req.URL)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("shorten url error: %w", err).Error())
	}

	resp := models.ShortenResponse{
		Result: shortURL,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return c.JSON(http.StatusCreated, resp)
}
