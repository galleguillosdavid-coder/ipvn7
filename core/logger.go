package core

import (
	"log/slog"
	"os"
	"sync"
)

var (
	loggerOnce sync.Once
	loggerMu   sync.RWMutex
	appLogger  *slog.Logger
)

// InitLogger initializes the global structured logger for IPv7
func InitLogger(asJSON bool, debugLevel bool) {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	var level slog.Level = slog.LevelInfo
	if debugLevel {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if asJSON {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	appLogger = slog.New(handler)
	slog.SetDefault(appLogger)
}

// Log returns the active structured logger instance
func Log() *slog.Logger {
	loggerOnce.Do(func() {
		if appLogger == nil {
			InitLogger(false, false)
		}
	})
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return appLogger
}
