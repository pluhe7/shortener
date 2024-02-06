package handlers

import (
	gocontext "context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/context"
	"github.com/pluhe7/shortener/internal/models"
	"github.com/pluhe7/shortener/internal/storage"
)

func InitHandlers(srv *app.Server) {
	srv.Echo.Use(
		ContextMiddleware(srv),
		RequestLoggerMiddleware,
		CompressorMiddleware,
		AuthMiddleware,
	)

	srv.Echo.GET("/:id", echoHandler(ExpandHandler))
	srv.Echo.GET("/ping", echoHandler(PingDatabaseHandler))
	srv.Echo.POST("/", echoHandler(ShortenHandler))

	apiGroup := srv.Echo.Group("/api")
	apiGroup.GET("/user/urls", echoHandler(GetUserURLs))
	apiGroup.POST("/shorten", echoHandler(APIShortenHandler))
	apiGroup.POST("/shorten/batch", echoHandler(APIBatchShortenHandler))
}

func echoHandler(h func(cc *context.Context) error) echo.HandlerFunc {
	return func(c echo.Context) error {
		cc := c.(*context.Context)
		return h(cc)
	}
}

func ExpandHandler(c *context.Context) error {
	id := c.Param("id")

	expandedURL, err := c.Server.ExpandURL(id)
	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, storage.ErrURLNotFound) {
			status = http.StatusNotFound
		}

		return c.String(status, fmt.Errorf("expand url error: %w", err).Error())
	}

	return c.Redirect(http.StatusTemporaryRedirect, expandedURL)
}

func ShortenHandler(c *context.Context) error {
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("read request body error: %w", err).Error())
	}

	originalURL := string(bodyBytes)
	respStatus := http.StatusCreated

	shortURL, err := c.Server.ShortenURL(originalURL, c.SessionUserID)
	if err != nil {
		err = fmt.Errorf("shorten url error: %w", err)

		if errors.Is(err, app.ErrEmptyURL) {
			return c.String(http.StatusBadRequest, err.Error())

		} else if errors.Is(err, storage.ErrDuplicateRecord) {
			shortURL, err = c.Server.GetExistingShortURL(originalURL)
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

func APIShortenHandler(c *context.Context) error {
	var req models.ShortenRequest

	requestDecoder := json.NewDecoder(c.Request().Body)
	err := requestDecoder.Decode(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("decode request error: %w", err).Error())
	}

	respStatus := http.StatusCreated

	shortURL, err := c.Server.ShortenURL(req.URL, c.SessionUserID)
	if err != nil {
		err = fmt.Errorf("shorten url error: %w", err)

		if errors.Is(err, app.ErrEmptyURL) {
			return c.String(http.StatusBadRequest, err.Error())

		} else if errors.Is(err, storage.ErrDuplicateRecord) {
			shortURL, err = c.Server.GetExistingShortURL(req.URL)
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

func PingDatabaseHandler(c *context.Context) error {
	ctx, cancel := gocontext.WithTimeout(gocontext.Background(), time.Second)
	defer cancel()

	err := c.Server.Storage.PingContext(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("storage ping: %w", err).Error())
	}

	return c.NoContent(http.StatusOK)
}

func APIBatchShortenHandler(c *context.Context) error {
	var req []models.OriginalURLWithID

	requestDecoder := json.NewDecoder(c.Request().Body)
	err := requestDecoder.Decode(&req)
	if err != nil {
		return c.String(http.StatusBadRequest, fmt.Errorf("decode request error: %w", err).Error())
	}

	shortURLs, err := c.Server.BatchShortenURLs(req, c.SessionUserID)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("shorten url error: %w", err).Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return c.JSON(http.StatusCreated, shortURLs)
}

func GetUserURLs(c *context.Context) error {
	authCookie, _ := c.Cookie("Token")
	if authCookie == nil || c.SessionUserID == "" {
		return c.NoContent(http.StatusUnauthorized)
	}

	userURLs, err := c.Server.Storage.FindByUserID(c.SessionUserID)
	if err != nil {
		return c.String(http.StatusInternalServerError, fmt.Errorf("get user records error: %w", err).Error())
	}

	if len(userURLs) < 1 {
		return c.NoContent(http.StatusNoContent)
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	return c.JSON(http.StatusOK, userURLs)
}
