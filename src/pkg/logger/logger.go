package logger

import (
	"log/slog"
	"os"
)

var logLevels = map[string]slog.Level{
	"debug": slog.LevelDebug,
	"info":  slog.LevelInfo,
}

func Init(logLevel string) {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: logLevel == "debug",
		Level:     logLevels[logLevel],
	})

	slog.SetDefault(slog.New(handler))
}

func Info(msg string, args ...any) {
	slog.Info(msg, args...)
}

func Error(msg string, args ...any) {
	slog.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
}
