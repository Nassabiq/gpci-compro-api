package utils

import (
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	opt := &slog.HandlerOptions{Level: slog.LevelInfo}
	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, opt)
	if env == "development" {
		handler = slog.NewTextHandler(os.Stdout, opt)
	}
	return slog.New(handler)
}
