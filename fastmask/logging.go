package main

import (
	"log/slog"
	"os"
)

func ConfigureLogging() {
	options := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	handler := slog.NewTextHandler(os.Stdout, options)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
