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
	srv.Echo.POST(`/api/shorten/batch`, srvHandler.APIBatchShortenHandler)
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

	originalURL := string(bodyBytes)
	respStatus := http.StatusCreated

	shortURL, err := s.ShortenURL(originalURL)
	if err != nil {
		err = fmt.Errorf("shorten url error: %w", err)

		if errors.Is(err, app.ErrEmptyURL) {
			return c.String(http.StatusBadRequest, err.Error())

		} else if errors.Is(err, storage.ErrDuplicateRecord) {
			shortURL, err = s.GetExistingShortURL(originalURL)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			respStatus = http.StatusConflict

		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextPlain)

	return c.String(respStatus, shortURL)
}

func (s *SrvHandler) APIShortenHandler(c echo.Context) error {
	var req models.ShortenRequest

	requestDecoder := json.NewDecoder(c.Request().Body)
	err := requestDecoder.Decode(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("decode request error: %w", err).Error())
	}

	respStatus := http.StatusCreated

	shortURL, err := s.ShortenURL(req.URL)
	if err != nil {
		err = fmt.Errorf("shorten url error: %w", err)

		if errors.Is(err, app.ErrEmptyURL) {
			return c.String(http.StatusBadRequest, err.Error())

		} else if errors.Is(err, storage.ErrDuplicateRecord) {
			shortURL, err = s.GetExistingShortURL(req.URL)
			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			respStatus = http.StatusConflict

		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	resp := models.ShortenResponse{
		Result: shortURL,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return c.JSON(respStatus, resp)
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

func (s *SrvHandler) APIBatchShortenHandler(c echo.Context) error {
	var req []models.OriginalURLWithID

	requestDecoder := json.NewDecoder(c.Request().Body)
	err := requestDecoder.Decode(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("decode request error: %w", err).Error())
	}

	shortURLs, err := s.BatchShortenURLs(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("shorten url error: %w", err).Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return c.JSON(http.StatusCreated, shortURLs)
}
