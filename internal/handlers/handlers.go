package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/models"
	"github.com/pluhe7/shortener/internal/storage"
)

type SrvHandler struct {
	*app.Server
}

func InitHandlers(srv *app.Server) {
	srvHandler := SrvHandler{srv}

	srv.Echo.Use(RequestLogger, CompressorMiddleware)

	srv.Echo.GET(`/:id`, srvHandler.ExpandHandler)
	srv.Echo.GET(`/ping`, srvHandler.PingDatabaseHandler)
	srv.Echo.POST(`/`, srvHandler.ShortenHandler)
	srv.Echo.POST(`/api/shorten`, srvHandler.APIShortenHandler)
}

func (s *SrvHandler) ExpandHandler(c echo.Context) error {
	id := c.Param("id")

	expandedURL, err := s.ExpandURL(id)
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, storage.ErrURLNotFound) {
			status = http.StatusNotFound
		}

		return c.String(status, fmt.Errorf("expand url error: %w", err).Error())
	}

	return c.Redirect(http.StatusTemporaryRedirect, expandedURL)
}

func (s *SrvHandler) ShortenHandler(c echo.Context) error {
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("read request body error: %w", err).Error())
	}

	shortURL, err := s.ShortenURL(string(bodyBytes))
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("shorten url error: %w", err).Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)

	return c.String(http.StatusCreated, shortURL)
}

func (s *SrvHandler) APIShortenHandler(c echo.Context) error {
	var req models.ShortenRequest

	requestDecoder := json.NewDecoder(c.Request().Body)
	err := requestDecoder.Decode(&req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("decode request error: %w", err).Error())
	}

	shortURL, err := s.ShortenURL(req.URL)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("shorten url error: %w", err).Error())
	}

	resp := models.ShortenResponse{
		Result: shortURL,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return c.JSON(http.StatusCreated, resp)
}

func (s *SrvHandler) PingDatabaseHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := s.Storage.PingContext(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("storage ping: %w", err).Error())
	}

	return c.NoContent(http.StatusOK)
}
