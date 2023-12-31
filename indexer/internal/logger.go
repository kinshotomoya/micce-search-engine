package internal

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

func InitLogger() {
	option := &slog.HandlerOptions{
		AddSource: false,
		Level:     slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, option)
	Logger = slog.New(handler)
	slog.Default()
}
