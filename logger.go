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

func (l *QueryBuilder) Debug(start time.Time, query string, args ...any) {
	l.logger.Debug("SQL",
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (l *QueryBuilder) Info(start time.Time, query string, args ...any) {
	l.logger.Info("SQL",
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (l *QueryBuilder) Warn(start time.Time, query string, args ...any) {
	l.logger.Warn("SQL",
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (l *QueryBuilder) Error(start time.Time, query string, args ...any) {
	l.logger.Error("SQL",
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}
