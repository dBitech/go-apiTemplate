package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
}

// Field defines a log field
type Field = zapcore.Field

// String creates a string field
func String(key, value string) Field {
	return zap.String(key, value)
}

// Int creates an integer field
func Int(key string, value int) Field {
	return zap.Int(key, value)
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

// Error creates an error field
func Error(err error) Field {
	return zap.Error(err)
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

// Any creates a field for any value
func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}

type loggerImpl struct {
	logger *zap.Logger
}

// New creates a new logger instance
func New(level, format string) (Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var config zap.Config
	if format == "json" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)
	if err != nil {
		return nil, err
	}

	return &loggerImpl{logger: logger}, nil
}

// Default returns a default logger instance
func Default() Logger {
	logger, err := New("info", "json")
	if err != nil {
		// If we can't create a logger, create a minimal one
		config := zap.NewProductionConfig()
		zapLogger, _ := config.Build()
		return &loggerImpl{logger: zapLogger}
	}
	return logger
}

func (l *loggerImpl) Debug(msg string, fields ...Field) {
	l.logger.Debug(msg, fields...)
}

func (l *loggerImpl) Info(msg string, fields ...Field) {
	l.logger.Info(msg, fields...)
}

func (l *loggerImpl) Warn(msg string, fields ...Field) {
	l.logger.Warn(msg, fields...)
}

func (l *loggerImpl) Error(msg string, fields ...Field) {
	l.logger.Error(msg, fields...)
}

func (l *loggerImpl) Fatal(msg string, fields ...Field) {
	l.logger.Fatal(msg, fields...)
	os.Exit(1)
}

func (l *loggerImpl) With(fields ...Field) Logger {
	return &loggerImpl{logger: l.logger.With(fields...)}
}

func (l *loggerImpl) WithContext(ctx context.Context) Logger {
	// Extract request ID or trace ID from context if available
	// For now we'll just return the same logger
	return l
}

// ContextKey is the key used to store the logger in the context
type ContextKey string

const loggerKey ContextKey = "logger"

// FromContext extracts a logger from a context
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return Default()
}

// ToContext adds a logger to a context
func ToContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
