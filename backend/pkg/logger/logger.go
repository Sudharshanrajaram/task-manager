package logger

import (
	"log/slog"
	"os"
	"strings"
)

// InitLogger initializes and sets the default global slog logger based on the environment.
// In production, it formats logs as structured JSON. In development, it formats as clean text.
func InitLogger(env string) *slog.Logger {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}

	if strings.ToLower(env) == "production" {
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	l := slog.New(handler)
	slog.SetDefault(l)
	return l
}
