package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/pluhe7/shortener/internal/compressor"
	"github.com/pluhe7/shortener/internal/logger"
)

func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
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
	return func(c echo.Context) error {
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
