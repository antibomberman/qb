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

func (q *QueryBuilder) Debug(msg string, start time.Time, query string, args ...any) {
	q.logger.Debug(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (q *QueryBuilder) Info(msg string, start time.Time, query string, args ...any) {
	q.logger.Info(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (q *QueryBuilder) Warn(msg string, start time.Time, query string, args ...any) {
	q.logger.Warn(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}

func (q *QueryBuilder) Error(msg string, start time.Time, query string, args ...any) {
	q.logger.Error(msg,
		slog.String("query", query),
		slog.Any("arg", args),
		slog.String("mc", time.Since(start).String()),
	)
}
