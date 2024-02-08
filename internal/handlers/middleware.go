package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pluhe7/shortener/internal/app"
	"github.com/pluhe7/shortener/internal/compressor"
	"github.com/pluhe7/shortener/internal/context"
	"github.com/pluhe7/shortener/internal/logger"
	"github.com/pluhe7/shortener/internal/util"
)

func ContextMiddleware(server *app.Server) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := context.NewContext(c, server)
			return next(cc)
		}
	}
}

func RequestLoggerMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*context.Context)

		start := time.Now()

		if err := next(c); err != nil {
			c.Error(err)
		}

		duration := time.Since(start)

		logger.Log.Info("got incoming HTTP request",
			zap.Duration("duration", duration),
			zap.Int("status", c.Response().Status),
			zap.Int64("size", c.Response().Size),
		)

		return nil
	}
}

var compressibleContentTypes = map[string]bool{
	echo.MIMEApplicationJSON: true,
	echo.MIMETextHTML:        true,
	"application/x-gzip":     true,
}

func CompressorMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*context.Context)

		acceptHeader := c.Request().Header.Get(echo.HeaderAccept)
		_, isCompressibleAccept := compressibleContentTypes[acceptHeader]

		acceptEncodingHeader := c.Request().Header.Get(echo.HeaderAcceptEncoding)
		isAcceptGzip := strings.Contains(acceptEncodingHeader, "gzip")

		if isAcceptGzip && isCompressibleAccept {
			compressWriter := compressor.NewGzipCompressWriter(c.Response().Writer)
			defer compressWriter.Close()

			c.Response().Writer = compressWriter
		}

		contentTypeHeader := c.Request().Header.Get(echo.HeaderContentType)
		_, isCompressibleContentType := compressibleContentTypes[contentTypeHeader]

		contentEncodingHeader := c.Request().Header.Get(echo.HeaderContentEncoding)
		isContentGzip := strings.Contains(contentEncodingHeader, "gzip")

		if isContentGzip && isCompressibleContentType {
			compressReader, err := compressor.NewGzipCompressReader(c.Request().Body)
			if err != nil {
				return fmt.Errorf("create new compress reader: %w", err)
			}
			defer compressReader.Close()

			c.Request().Body = compressReader
		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}

func AuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(ec echo.Context) error {
		c := ec.(*context.Context)

		const cookieName = "Token"

		authCookie, err := c.Cookie(cookieName)
		if err != nil || !util.IsTokenValid(authCookie.Value, c.Server.Config.SecretKey) {
			userID, err := util.GenerateID()
			if err != nil {
				return fmt.Errorf("generate id: %w", err)
			}

			token, err := util.CreateToken(userID, c.Server.Config.SecretKey)
			if err != nil {
				return fmt.Errorf("create token: %w", err)
			}

			c.Server.SessionUserID = userID

			c.SetCookie(&http.Cookie{
				Name:  cookieName,
				Value: token,
			})
		} else {
			c.Server.SessionUserID, err = util.GetUserID(authCookie.Value, c.Server.Config.SecretKey)
			if err != nil {
				return fmt.Errorf("get user id from token: %w", err)
			}
		}

		if err = next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
