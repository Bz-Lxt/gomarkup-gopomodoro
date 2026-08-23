package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var (
	mu      sync.RWMutex
	current *slog.Logger
)

func Init(level string, json bool) *slog.Logger {
	var lv slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lv = slog.LevelDebug
	case "warn", "warning":
		lv = slog.LevelWarn
	case "error":
		lv = slog.LevelError
	default:
		lv = slog.LevelInfo
	}
	opts := &slog.HandlerOptions{Level: lv}
	var h slog.Handler
	if json {
		h = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		h = slog.NewTextHandler(os.Stdout, opts)
	}
	l := slog.New(h)
	mu.Lock()
	current = l
	mu.Unlock()
	slog.SetDefault(l)
	return l
}

func L() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		return slog.Default()
	}
	return current
}
