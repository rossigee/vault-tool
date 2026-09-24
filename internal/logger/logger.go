package logger

import (
	"log/slog"
	"os"
	"sync"
)

var (
	quiet     bool
	mu        sync.Mutex
	defaulter *slog.Logger
)

func init() {
	defaulter = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func SetQuiet(q bool) {
	mu.Lock()
	defer mu.Unlock()
	quiet = q
}

func IsQuiet() bool {
	mu.Lock()
	defer mu.Unlock()
	return quiet
}

func Info(msg string, args ...any) {
	if IsQuiet() {
		return
	}
	defaulter.Info(msg, args...)
}

func Error(msg string, args ...any) {
	if IsQuiet() {
		return
	}
	defaulter.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	if IsQuiet() {
		return
	}
	defaulter.Debug(msg, args...)
}

func Logger() *slog.Logger {
	return defaulter
}
