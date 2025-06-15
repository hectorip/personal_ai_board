package logger

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// Logger interface for structured logging
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// SlogLogger implements Logger using Go's structured logging
type SlogLogger struct {
	logger *slog.Logger
}

// New creates a new logger with the specified level
func New(level string) Logger {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &SlogLogger{logger: logger}
}

// NewJSON creates a new JSON logger
func NewJSON(level string) Logger {
	var logLevel slog.Level

	switch strings.ToLower(level) {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn", "warning":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &SlogLogger{logger: logger}
}

// Info logs an info message
func (l *SlogLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, convertArgs(args...)...)
}

// Error logs an error message
func (l *SlogLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, convertArgs(args...)...)
}

// Debug logs a debug message
func (l *SlogLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, convertArgs(args...)...)
}

// Warn logs a warning message
func (l *SlogLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, convertArgs(args...)...)
}

// convertArgs converts variadic args to slog.Attr format
func convertArgs(args ...interface{}) []any {
	if len(args)%2 != 0 {
		// If odd number of args, treat the last one as a value with "extra" key
		args = append(args, "extra", args[len(args)-1])
		args = args[:len(args)-2]
	}

	converted := make([]any, 0, len(args))
	for i := 0; i < len(args); i += 2 {
		key := fmt.Sprintf("%v", args[i])
		value := args[i+1]
		converted = append(converted, slog.Any(key, value))
	}

	return converted
}

// NoOpLogger is a logger that does nothing (for testing)
type NoOpLogger struct{}

// NewNoOp creates a no-op logger
func NewNoOp() Logger {
	return &NoOpLogger{}
}

func (l *NoOpLogger) Info(msg string, args ...interface{})  {}
func (l *NoOpLogger) Error(msg string, args ...interface{}) {}
func (l *NoOpLogger) Debug(msg string, args ...interface{}) {}
func (l *NoOpLogger) Warn(msg string, args ...interface{})  {}