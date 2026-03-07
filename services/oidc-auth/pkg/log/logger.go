package log

import (
	"os"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Config logger configuration
type Config struct {
	Level    string `json:"level" mapstructure:"level"`
	Filename string `json:"filename" mapstructure:"filename"`
	MaxSize  int    `json:"maxSize" mapstructure:"maxSize"`
	MaxAge   int    `json:"maxAge" mapstructure:"maxAge"`
	Compress bool   `json:"compress" mapstructure:"compress"`
}

// InitLogger initializes the global logger
func InitLogger(cfg *Config) {
	once.Do(func() {
		// Set log level
		var level zapcore.Level
		switch cfg.Level {
		case "debug":
			level = zapcore.DebugLevel
		case "info":
			level = zapcore.InfoLevel
		case "warn":
			level = zapcore.WarnLevel
		case "error":
			level = zapcore.ErrorLevel
		default:
			level = zapcore.InfoLevel
		}

		// Set log encoder
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     timeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// Configure console output
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)

		// Add caller information
		opts := []zap.Option{
			zap.AddCaller(),
			zap.AddCallerSkip(1),
		}

		// Create logger
		globalLogger = zap.New(consoleCore, opts...)
	})
}

// Custom time encoder
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.DateTime))
}

// GetLogger returns the global logger
func GetLogger() *zap.Logger {
	if globalLogger == nil {
		// If not initialized, use default configuration
		InitLogger(&Config{
			Level: "info",
		})
	}
	return globalLogger
}

// Debug level log
func Debug(ctx any, format string, args ...any) {
	GetLogger().Sugar().Debugf(format, args...)
}

// Info level log
func Info(ctx any, format string, args ...any) {
	GetLogger().Sugar().Infof(format, args...)
}

// Warn level log
func Warn(ctx any, format string, args ...any) {
	GetLogger().Sugar().Warnf(format, args...)
}

// Error level log
func Error(ctx any, format string, args ...any) {
	GetLogger().Sugar().Errorf(format, args...)
}

// Fatal level log
func Fatal(ctx any, format string, args ...any) {
	GetLogger().Sugar().Fatalf(format, args...)
}

// With creates a logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	return GetLogger().With(fields...)
}
