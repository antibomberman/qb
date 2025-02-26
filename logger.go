package qb

import (
	"log/slog"
	"os"
	"time"
)

func NewDefaultLogger() *slog.Logger {
	return slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
	)
}

func (l *QueryBuilder) Debug(msg string, start time.Time, query string, args ...any) {
	l.logger.Debug(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (l *QueryBuilder) Info(msg string, start time.Time, query string, args ...any) {
	l.logger.Info(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (l *QueryBuilder) Warn(msg string, start time.Time, query string, args ...any) {
	l.logger.Warn(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (l *QueryBuilder) Error(msg string, start time.Time, query string, args ...any) {
	l.logger.Error(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}
