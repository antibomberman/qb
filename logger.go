package qb

import (
	"context"
	"log/slog"
	"time"
)

func (q *QueryBuilder) Debug(msg string, start time.Time, query string, args ...any) {
	q.log(slog.LevelDebug, msg, start, query, args...)
}

func (q *QueryBuilder) Info(msg string, start time.Time, query string, args ...any) {
	q.log(slog.LevelInfo, msg, start, query, args...)
}

func (q *QueryBuilder) Warn(msg string, start time.Time, query string, args ...any) {
	q.log(slog.LevelWarn, msg, start, query, args...)
}

func (q *QueryBuilder) Error(msg string, start time.Time, query string, args ...any) {
	q.log(slog.LevelError, msg, start, query, args...)
}

func (q *QueryBuilder) log(level slog.Level, msg string, start time.Time, query string, args ...any) {
	if q.logger != nil {
		attrs := []slog.Attr{
			slog.String("query", query),
			slog.Any("args", args),
			slog.String("duration", time.Since(start).String()),
		}
		q.logger.LogAttrs(context.Background(), level, msg, attrs...)
	}
}
