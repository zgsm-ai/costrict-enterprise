package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"completion-agent/pkg/env"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

/**
 * sizeLimitedWriter 日志文件大小限制写入器
 * @description
 * - 实现文件大小限制和自动轮转功能
 * - 当文件达到最大大小时，会重命名原文件并创建新文件
 * - 线程安全的实现
 * - 实现 zapcore.WriteSyncer 接口
 */
type sizeLimitedWriter struct {
	filePath string
	maxSize  int64
	file     *os.File
	mu       sync.Mutex
}

// 实现 zapcore.WriteSyncer 接口
func (w *sizeLimitedWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	return w.file.Sync()
}

var (
	sizeLimitedWriterInstance *sizeLimitedWriter
)

/**
 * 全局 logger 实例
 * @description
 * - 全局日志记录器实例
 * - 使用zap高性能日志库
 * - 在init函数中初始化
 * - 提供应用程序的统一日志记录入口
 * @example
 * Logger.Info("应用启动")
 * Logger.Error("发生错误", zap.Error(err))
 */
var Logger *zap.Logger

/**
 * 初始化日志系统
 * @description
 * - 创建生产级别的日志配置
 * - 设置日志级别为Info
 * - 配置时间戳格式为本地时区
 * - 构建日志记录器实例
 * - 替换zap的全局logger
 * @throws
 * - 如果日志构建失败，会导致程序panic
 * @example
 * // 包初始化时自动调用
 * // 不需要手动调用
 */
/**
 * InitLogger 初始化日志系统
 * @param {string} logPath - 日志文件路径，如果为空或"console"则使用默认路径
 * @param {string} level - 日志级别，支持"debug", "info", "warn", "error"
 * @param {bool} toConsole - 是否同时输出到控制台
 * @param {int64} maxSize - 日志文件最大大小（字节），默认50MB
 * @description
 * - 初始化zap日志配置
 * - 支持日志文件大小限制和自动轮转
 * - 自动创建日志目录
 * - 支持同时输出到文件和控制台
 * - 错误级别日志单独保存为JSON格式
 * @throws
 * - 如果日志构建失败，会导致程序panic
 * @example
 * InitLogger("", "info", true, 5*1024*1024)
 * // 使用默认路径，info级别，同时输出到控制台，最大5MB
 */
func InitLogger(logPath string, mode string, maxSize int64) {
	// 设置默认值
	if logPath == "console" || logPath == "" {
		logPath = filepath.Join(env.GetCostrictDir(), "logs", "completion-agent.log")
	}
	if maxSize <= 0 {
		maxSize = 5 * 1024 * 1024 // 默认5MB
	}

	// 确保日志目录存在
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(err)
	}

	// 创建大小限制的文件写入器
	var err error
	sizeLimitedWriterInstance, err = newSizeLimitedWriter(logPath, maxSize)
	if err != nil {
		panic(err)
	}
	if err := removeRedundantBackups(logPath, 1); err != nil {
		fmt.Fprintf(os.Stderr, "remove redundant backups: %s", err.Error())
	}

	// 根据模式创建不同的配置
	var core zapcore.Core
	if mode == "debug" {
		// debug模式：控制台使用开发格式，文件使用JSON格式
		consoleEncoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
			},
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		})

		fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
			},
			EncodeCaller: zapcore.ShortCallerEncoder,
		})

		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)
		fileCore := zapcore.NewCore(fileEncoder, sizeLimitedWriterInstance, zapcore.InfoLevel)
		core = zapcore.NewTee(consoleCore, fileCore)
	} else {
		// 生产模式：控制台和文件都使用JSON格式，但控制台有更好的可读性
		consoleEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Local().Format("06-01-02 15:04:05"))
			},
			EncodeCaller:   zapcore.ShortCallerEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
		})

		fileEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:       "ts",
			LevelKey:      "level",
			NameKey:       "logger",
			CallerKey:     "caller",
			FunctionKey:   zapcore.OmitKey,
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.CapitalLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Local().Format("2006-01-02 15:04:05.000"))
			},
			EncodeCaller: zapcore.ShortCallerEncoder,
		})

		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.InfoLevel)
		fileCore := zapcore.NewCore(fileEncoder, sizeLimitedWriterInstance, zapcore.InfoLevel)
		core = zapcore.NewTee(consoleCore, fileCore)
	}

	// 创建logger
	Logger = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(Logger)
}

/**
 * 创建新的大小限制写入器
 * @param {string} filePath - 日志文件路径
 * @param {int64} maxSize - 最大文件大小
 * @returns {sizeLimitedWriter} 返回写入器实例
 * @returns {error} 返回错误信息
 */
func newSizeLimitedWriter(filePath string, maxSize int64) (*sizeLimitedWriter, error) {
	w := &sizeLimitedWriter{
		filePath: filePath,
		maxSize:  maxSize,
	}

	if err := w.rotateIfNeeded(); err != nil {
		return nil, err
	}

	return w, nil
}

/**
 * 写入数据，检查文件大小并轮转
 * @param {[]byte} p - 要写入的数据
 * @returns {int} 写入的字节数
 * @returns {error} 错误信息
 */
func (w *sizeLimitedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 写入前检查是否需要轮转
	if err := w.rotateIfNeeded(); err != nil {
		return 0, err
	}

	return w.file.Write(p)
}

/**
 * 关闭文件
 * @returns {error} 错误信息
 */
func (w *sizeLimitedWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}

/**
 * 检查文件大小并轮转
 * @returns {error} 错误信息
 */
func (w *sizeLimitedWriter) rotateIfNeeded() error {
	// 检查文件是否存在并获取大小
	if w.file != nil {
		fileInfo, err := w.file.Stat()
		if err != nil {
			return err
		}
		if fileInfo.Size() < w.maxSize {
			// 文件大小在限制内，不需要轮转
			return nil
		}
		// 关闭当前文件
		if err := w.file.Close(); err != nil {
			return err
		}
		// 重命名当前文件，添加时间戳后缀
		timestamp := time.Now().Format("20060102-150405")
		backupPath := w.filePath + "." + timestamp
		if err := os.Rename(w.filePath, backupPath); err != nil {
			return err
		}
		if err := removeRedundantBackups(w.filePath, 1); err != nil {
			fmt.Fprintf(os.Stderr, "remove redundant backups: %s", err.Error())
		}
	}

	// 创建/打开日志文件
	file, err := os.OpenFile(w.filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = file
	return nil
}

func removeRedundantBackups(filePath string, backupCount int) error {
	if backupCount < 0 {
		return nil
	}
	dir := filepath.Dir(filePath)
	fprefix := filepath.Base(filePath)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	type item struct {
		path string
		tm   time.Time
	}
	var backups []item
	const tsLen = len("20060102-150405")

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, fprefix) {
			continue
		}
		// 后缀必须是 <timestamp>
		if len(name) < tsLen {
			continue
		}
		tsStr := name[len(name)-tsLen:]
		tm, err := time.Parse("20060102-150405", tsStr)
		if err != nil {
			continue // 格式不符，跳过
		}
		backups = append(backups, item{
			path: filepath.Join(dir, name),
			tm:   tm,
		})
	}

	// 按时间升序
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].tm.Before(backups[j].tm)
	})

	// 删除多余的
	toDel := len(backups) - backupCount
	for i := 0; i < toDel; i++ {
		if err := os.Remove(backups[i].path); err != nil {
			return err
		}
	}
	return nil
}

/**
 * SetLevel 设置日志级别
 * @param {string} level - 日志级别字符串，如"debug", "info", "warn", "error"
 * @description
 * - 解析输入的日志级别字符串
 * - 如果解析失败，记录警告日志并使用默认级别
 * - 更新全局logger的核心日志级别
 * - 支持的标准级别：debug, info, warn, error, dpanic, panic, fatal
 * @example
 * SetLevel("debug")
 * // 设置日志级别为debug，将显示更详细的日志
 *
 * SetLevel("invalid")
 * // 输出警告: Invalid log level, using default level (info)
 */
func SetLevel(level string) {
	levelValue, err := zapcore.ParseLevel(level)
	if err != nil {
		Logger.Warn("Invalid log level, using default level (info)")
		return
	}
	Logger.Core().Enabled(levelValue)
}

/**
 * 刷新所有日志到输出
 * @description
 * - 调用全局logger的Sync方法
 * - 确保所有缓冲的日志都被写入输出
 * - 通常在应用程序退出前调用
 * - 用于保证日志数据不丢失
 * @example
 * defer Sync()
 * // 在main函数结束时调用，确保所有日志都被写入
 */
func Sync() {
	Logger.Sync()
}

// 便捷函数，直接调用全局 logger 的方法

/**
 * 记录信息级别日志
 * @param {string} msg - 日志消息内容
 * @param {...zap.Field} fields - 可变参数，额外的日志字段
 * @description
 * - 记录Info级别的日志消息
 * - 支持添加结构化的字段信息
 * - 委托给全局Logger的Info方法
 * - 用于记录常规的应用程序运行信息
 * @example
 * Info("用户登录", zap.String("username", "john"), zap.Int("userId", 123))
 * // 输出: {"level":"info","msg":"用户登录","username":"john","userId":123}
 */
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

/**
 * 记录错误级别日志
 * @param {string} msg - 错误消息内容
 * @param {...zap.Field} fields - 可变参数，额外的日志字段
 * @description
 * - 记录Error级别的日志消息
 * - 支持添加结构化的字段信息
 * - 委托给全局Logger的Error方法
 * - 用于记录应用程序中的错误情况
 * @example
 * Error("数据库连接失败", zap.Error(err), zap.String("host", "localhost"))
 * // 输出: {"level":"error","msg":"数据库连接失败","error":"connection refused","host":"localhost"}
 */
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

/**
 * 记录调试级别日志
 * @param {string} msg - 调试消息内容
 * @param {...zap.Field} fields - 可变参数，额外的日志字段
 * @description
 * - 记录Debug级别的日志消息
 * - 支持添加结构化的字段信息
 * - 委托给全局Logger的Debug方法
 * - 用于记录详细的调试信息，仅在调试模式下输出
 * @example
 * Debug("处理请求", zap.String("method", "GET"), zap.String("path", "/api/users"))
 * // 在debug模式下输出: {"level":"debug","msg":"处理请求","method":"GET","path":"/api/users"}
 */
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

/**
 * 记录警告级别日志
 * @param {string} msg - 警告消息内容
 * @param {...zap.Field} fields - 可变参数，额外的日志字段
 * @description
 * - 记录Warn级别的日志消息
 * - 支持添加结构化的字段信息
 * - 委托给全局Logger的Warn方法
 * - 用于记录可能需要注意但不会导致程序错误的情况
 * @example
 * Warn("缓存即将过期", zap.Time("expireTime", time.Now().Add(time.Hour)))
 * // 输出: {"level":"warn","msg":"缓存即将过期","expireTime":"2023-01-01T12:00:00Z"}
 */
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

/**
 * 记录致命错误级别日志
 * @param {string} msg - 致命错误消息内容
 * @param {...zap.Field} fields - 可变参数，额外的日志字段
 * @description
 * - 记录Fatal级别的日志消息
 * - 支持添加结构化的字段信息
 * - 委托给全局Logger的Fatal方法
 * - 记录后会导致程序调用os.Exit(1)退出
 * - 用于记录无法恢复的致命错误
 * @example
 * Fatal("配置文件读取失败", zap.Error(err), zap.String("configPath", "/etc/app/config.json"))
 * // 输出错误信息后程序退出
 */
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

/**
 * 记录恐慌级别日志
 * @param {string} msg - 恐慌消息内容
 * @param {...zap.Field} fields - 可变参数，额外的日志字段
 * @description
 * - 记录Panic级别的日志消息
 * - 支持添加结构化的字段信息
 * - 委托给全局Logger的Panic方法
 * - 记录后会导致程序panic
 * - 用于记录严重错误并触发panic机制
 * @example
 * Panic("系统状态异常", zap.String("state", "critical"), zap.Int("errorCode", 500))
 * // 输出错误信息后触发panic
 */
func Panic(msg string, fields ...zap.Field) {
	Logger.Panic(msg, fields...)
}

/**
 * 创建带有额外字段的 logger
 * @param {...zap.Field} fields - 可变参数，要添加到logger的字段
 * @returns {*zap.Logger} 返回带有额外字段的新logger实例
 * @description
 * - 基于全局logger创建新的logger实例
 * - 新logger包含所有指定的额外字段
 * - 每次调用都创建新的logger实例
 * - 用于在特定上下文中添加固定的日志字段
 * @example
 * userLogger := With(zap.String("userId", "123"), zap.String("sessionId", "abc"))
 * userLogger.Info("用户操作")
 * // 输出: {"level":"info","msg":"用户操作","userId":"123","sessionId":"abc"}
 */
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}
