package qb

import (
	"fmt"
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
	Debug(query string, args ...any)
	Info(query string, args ...any)
	Warn(query string, args ...any)
	Error(query string, args ...any)
	QueryTiming(duration time.Duration, query string, args ...any)
}

type DefaultLogger struct {
	level LogLevel
}

func NewLogger(level LogLevel) *DefaultLogger {
	return &DefaultLogger{level: level}
}

func (l *DefaultLogger) log(level LogLevel, query string, args ...any) {
	if level >= l.level {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf("[%s] %s; args: %v\n", timestamp, query, args)
		fmt.Fprint(os.Stdout, message)
	}
}

func (l *DefaultLogger) QueryTiming(duration time.Duration, query string, args ...any) {
	if l.level <= LogLevelDebug {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf("[%s] Query took %s: %s; args: %v\n",
			timestamp,
			duration.String(),
			query,
			args,
		)
		fmt.Fprint(os.Stdout, message)
	}
}

func (l *DefaultLogger) Debug(query string, args ...any) {
	l.log(LogLevelDebug, query, args...)
}

func (l *DefaultLogger) Info(query string, args ...any) {
	l.log(LogLevelInfo, query, args...)
}

func (l *DefaultLogger) Warn(query string, args ...any) {
	l.log(LogLevelWarn, query, args...)
}

func (l *DefaultLogger) Error(query string, args ...any) {
	l.log(LogLevelError, query, args...)
}
