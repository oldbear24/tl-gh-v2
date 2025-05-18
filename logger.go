package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

func LogDebugDuration(msg string, attrs ...any) func() {
	return LogDuration(slog.LevelDebug, msg, attrs...)
}
func LogInfoDuration(msg string, attrs ...any) func() {
	return LogDuration(slog.LevelInfo, msg, attrs...)
}
func LogWarnDuration(msg string, attrs ...any) func() {
	return LogDuration(slog.LevelWarn, msg, attrs...)
}
func LogErrorDuration(msg string, attrs ...any) func() {
	return LogDuration(slog.LevelError, msg, attrs...)
}
func LogDuration(level slog.Level, msg string, attrs ...any) func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		all := append(attrs, slog.Duration("duration", elapsed))
		log.Log(context.Background(), level, msg, all...)
	}

}

type AppQueryTracer struct {
	pgx.QueryTracer
}

func (qt *AppQueryTracer) TraceQueryStart(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	log.Debug("Query started", "query", data.SQL, "args", data.Args)
	return ctx
}

func (qt *AppQueryTracer) TraceQueryEnd(ctx context.Context, conn *pgx.Conn, data pgx.TraceQueryEndData) {
}
