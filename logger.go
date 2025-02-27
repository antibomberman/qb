package qb

import (
	"log/slog"
	"time"
)

func (q *QueryBuilder) Debug(msg string, start time.Time, query string, args ...any) {
	if q.logger != nil {
		q.logger.Debug(msg,
			slog.String("query", query),
			slog.Any("arg", args),
			slog.String("mc", time.Since(start).String()),
		)
	}
}

func (q *QueryBuilder) Info(msg string, start time.Time, query string, args ...any) {
	if q.logger != nil {
		q.logger.Info(msg,
			slog.String("query", query),
			slog.Any("arg", args),
			slog.String("mc", time.Since(start).String()),
		)
	}
}

func (q *QueryBuilder) Warn(msg string, start time.Time, query string, args ...any) {
	if q.logger != nil {
		q.logger.Warn(msg,
			slog.String("query", query),
			slog.Any("arg", args),
			slog.String("mc", time.Since(start).String()),
		)
	}
}

func (q *QueryBuilder) Error(msg string, start time.Time, query string, args ...any) {
	if q.logger != nil {
		q.logger.Error(msg,
			slog.String("query", query),
			slog.Any("arg", args),
			slog.String("mc", time.Since(start).String()),
		)
	}
}
