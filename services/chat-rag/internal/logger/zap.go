package logger

import (
	"context"
	"fmt"

	"github.com/zgsm-ai/chat-rag/internal/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// L is the global logger instance
var L *zap.Logger

func init() {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	L, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize zap logger: %v", err))
	}
}

func WithRequestID(ctx context.Context) *zap.Logger {
	if requestID := ctx.Value(types.HeaderRequestId); requestID != nil {
		if id, ok := requestID.(string); ok && id != "" {
			return L.With(zap.String("x-request-id", id))
		}
	}
	return L
}

// Sync flushes any buffered log entries and should be called before application exit
func Sync() {
	if err := L.Sync(); err != nil {
		L.Error("Failed to sync logger",
			zap.Error(err),
		)
	}
}

// Info logs a message at InfoLevel
func Info(msg string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// Debug logs a message at DebugLevel
func Debug(msg string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// Error logs a message at ErrorLevel
func Error(msg string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// Warn logs a message at WarnLevel
func Warn(msg string, fields ...zap.Field) {
	L.WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

// InfoC logs a message at InfoLevel with x-request-id from context
func InfoC(ctx context.Context, msg string, fields ...zap.Field) {
	WithRequestID(ctx).WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

// DebugC logs a message at DebugLevel with x-request-id from context
func DebugC(ctx context.Context, msg string, fields ...zap.Field) {
	WithRequestID(ctx).WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

// ErrorC logs a message at ErrorLevel with x-request-id from context
func ErrorC(ctx context.Context, msg string, fields ...zap.Field) {
	WithRequestID(ctx).WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

// WarnC logs a message at WarnLevel with x-request-id from context
func WarnC(ctx context.Context, msg string, fields ...zap.Field) {
	WithRequestID(ctx).WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}
