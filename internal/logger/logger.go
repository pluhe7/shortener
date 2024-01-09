package logger

import (
	"go.uber.org/zap"
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
