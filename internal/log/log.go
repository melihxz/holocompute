package log

import (
	"context"
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger
type Logger struct {
	*slog.Logger
}

// New creates a new logger with the specified level
func New(level slog.Level) *Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{slog.New(handler)}
}

// FromContext retrieves a logger from context, or returns the default logger
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value("logger").(*Logger); ok {
		return logger
	}
	return New(slog.LevelInfo)
}

// WithContext adds a logger to the context
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "logger", l)
}

// With adds key-value pairs to the logger
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{l.Logger.With(args...)}
}