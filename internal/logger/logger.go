package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	logLevel  slog.Level
	mu        sync.Mutex
	defaulter *slog.Logger
)

func init() {
	// Default: quiet (no output)
	logLevel = slog.LevelError + 1 // Higher than any standard level
	defaulter = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

// SetVerbose enables info-level logging
func SetVerbose() {
	mu.Lock()
	defer mu.Unlock()
	logLevel = slog.LevelInfo
	defaulter = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

// SetDebug enables debug-level logging
func SetDebug() {
	mu.Lock()
	defer mu.Unlock()
	logLevel = slog.LevelDebug
	defaulter = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	}))
}

func Info(msg string, args ...any) {
	defaulter.Info(msg, args...)
}

func Error(msg string, args ...any) {
	defaulter.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	defaulter.Debug(msg, args...)
}

func Logger() *slog.Logger {
	return defaulter
}
