package qb

import (
	"log/slog"
	"os"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type Logger interface {
	Debug(duration time.Time, query string, args ...any)
	Info(duration time.Time, query string, args ...any)
	Warn(duration time.Time, query string, args ...any)
	Error(duration time.Time, query string, args ...any)
}

type DefaultLogger struct {
	level  LogLevel
	logger *slog.Logger
}

func NewLogger(level LogLevel) *DefaultLogger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{
					Key:   slog.TimeKey,
					Value: slog.StringValue(time.Now().Format("2006-01-02 15:04:05")),
				}
			}
			return a
		},
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &DefaultLogger{
		level:  level,
		logger: logger,
	}
}

func (l *DefaultLogger) Debug(start time.Time, query string, args ...any) {
	l.logger.Debug("SQL", map[string]any{
		"query": query,
		"args":  args,
		"time":  time.Since(start).String(),
	})
}

func (l *DefaultLogger) Info(start time.Time, query string, args ...any) {
	l.logger.Debug("SQL", map[string]any{
		"query": query,
		"args":  args,
		"time":  time.Since(start).String(),
	})
}

func (l *DefaultLogger) Warn(start time.Time, query string, args ...any) {
	l.logger.Debug("SQL", map[string]any{
		"query": query,
		"args":  args,
		"time":  time.Since(start).String(),
	})
}

func (l *DefaultLogger) Error(start time.Time, query string, args ...any) {
	l.logger.Debug("SQL", map[string]any{
		"query": query,
		"args":  args,
		"time":  time.Since(start).String(),
	})
}
