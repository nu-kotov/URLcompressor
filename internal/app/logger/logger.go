package logger

import (
	"go.uber.org/zap"
)

// Log - глобальный экземпляр логгера.
var Log *zap.Logger = zap.NewNop()

// NewLogger - конструктор логгера.
func NewLogger(level string) error {

	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}

	cfg := zap.NewProductionConfig()

	cfg.Level = lvl

	zl, err := cfg.Build()
	if err != nil {
		return err
	}

	Log = zl
	return nil
}
