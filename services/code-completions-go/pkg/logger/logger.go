package logger

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局 logger 实例
var Logger *zap.Logger

func init() {
	// 使用 NewProductionConfig 并可选调整日志级别
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel) // 设置日志级别为 Info

	// 配置时间戳格式为本地时区
	config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
	}

	// 注意：禁止将日志输出到文件
	// config.OutputPaths = []string{"./app.log"}

	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}

	// 替换 zap 的全局 logger
	zap.ReplaceGlobals(Logger)
}

func SetLevel(level string) {
	levelValue, err := zapcore.ParseLevel(level)
	if err != nil {
		Logger.Warn("Invalid log level, using default level (info)")
		return
	}
	Logger.Core().Enabled(levelValue)
}

func SetMode(mode string) {
	var l *zap.Logger
	var err error
	if mode == "debug" {
		// 开发模式也使用本地时区
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
		}
		l, err = config.Build()
	} else {
		// 生产模式也使用本地时区
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
		}
		l, err = config.Build()
	}
	if err != nil {
		panic(err)
	}
	zap.ReplaceGlobals(l)
}

// Sync 刷新所有日志到输出
func Sync() {
	Logger.Sync()
}

// 便捷函数，直接调用全局 logger 的方法

// Info 记录信息级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Error 记录错误级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Debug 记录调试级别日志
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Warn 记录警告级别日志
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Fatal 记录致命错误级别日志
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// Panic 记录恐慌级别日志
func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}

// With 创建带有额外字段的 logger
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}
