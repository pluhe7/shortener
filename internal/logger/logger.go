package logger

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"time"
)

var Log = zap.NewNop()

func Initialize(level string) error {
	atomicLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = atomicLevel

	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = logger
	return nil
}

func RequestLogger(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()

		if err := next(c); err != nil {
			c.Error(err)
		}

		duration := time.Since(start)

		Log.Info("got incoming HTTP request",
			zap.Duration("duration", duration),
			zap.Int("status", c.Response().Status),
			zap.Int64("size", c.Response().Size),
		)

		return nil
	}
}
