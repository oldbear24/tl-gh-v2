package main

import (
	"context"
	"log/slog"
	"time"
)

type AppLogger struct {
	*slog.Logger
}

func (logger *AppLogger) DebugDuration(msg string, attrs ...any) func() {
	return logger.LogDuration(slog.LevelDebug, msg, attrs...)
}
func (logger *AppLogger) InfoDuration(msg string, attrs ...any) func() {
	return logger.LogDuration(slog.LevelInfo, msg, attrs...)
}
func (logger *AppLogger) WarnDuration(msg string, attrs ...any) func() {
	return logger.LogDuration(slog.LevelWarn, msg, attrs...)
}
func (logger *AppLogger) ErrorDuration(msg string, attrs ...any) func() {
	return logger.LogDuration(slog.LevelError, msg, attrs...)
}
func (logger *AppLogger) LogDuration(level slog.Level, msg string, attrs ...any) func() {
	if level == slog.LevelDebug {
		return func() {}
	}
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		all := append(attrs, slog.Duration("duration", elapsed))
		logger.Log(context.Background(), level, msg, all...)
	}
}
