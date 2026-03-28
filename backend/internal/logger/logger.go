// Package logger 提供结构化日志功能
// 兼容 Go 1.18，使用标准库 log + 自定义格式化
package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level 日志级别
type Level int

const (
	LevelDebug Level = 0
	LevelInfo  Level = 1
	LevelWarn  Level = 2
	LevelError Level = 3
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	With(args ...any) Logger
	WithContext(ctx context.Context) Logger
}

// loggerImpl 日志实现
type loggerImpl struct {
	level  Level
	format string // json/text
	output io.Writer
	fields map[string]any
}

var (
	defaultLogger Logger
	once          sync.Once
)

// Init 初始化全局日志
// level: debug/info/warn/error
// format: json/text
func Init(level, format string) error {
	var err error
	once.Do(func() {
		defaultLogger, err = New(level, format, os.Stdout)
	})
	return err
}

// New 创建新的日志实例
func New(level, format string, w io.Writer) (Logger, error) {
	var logLevel Level
	switch level {
	case "debug":
		logLevel = LevelDebug
	case "info":
		logLevel = LevelInfo
	case "warn":
		logLevel = LevelWarn
	case "error":
		logLevel = LevelError
	default:
		logLevel = LevelInfo
	}

	return &loggerImpl{
		level:  logLevel,
		format: format,
		output: w,
		fields: make(map[string]any),
	}, nil
}

// Get 获取全局日志实例
func Get() Logger {
	if defaultLogger == nil {
		// 如果未初始化，使用默认配置
		Init("info", "json")
	}
	return defaultLogger
}

// SetDefault 设置全局默认日志
func SetDefault(l Logger) {
	defaultLogger = l
}

// log 内部日志输出方法
func (l *loggerImpl) log(level Level, levelStr string, msg string, args ...any) {
	if level < l.level {
		return
	}

	// 构建字段
	fields := make(map[string]any)
	for k, v := range l.fields {
		fields[k] = v
	}

	// 解析 args（key-value 对）
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}

	// 构建日志记录
	record := map[string]any{
		"time":    time.Now().Format(time.RFC3339),
		"level":   levelStr,
		"message": msg,
	}
	for k, v := range fields {
		record[k] = v
	}

	// 输出
	var output string
	if l.format == "json" {
		data, _ := json.Marshal(record)
		output = string(data)
	} else {
		// 文本格式
		output = fmt.Sprintf("[%s] %s %s", record["time"], levelStr, msg)
		if len(fields) > 0 {
			fieldStr := ""
			for k, v := range fields {
				fieldStr += fmt.Sprintf(" %s=%v", k, v)
			}
			output += fieldStr
		}
	}

	fmt.Fprintln(l.output, output)
}

// Debug 输出调试日志
func (l *loggerImpl) Debug(msg string, args ...any) {
	l.log(LevelDebug, "DEBUG", msg, args...)
}

// Info 输出信息日志
func (l *loggerImpl) Info(msg string, args ...any) {
	l.log(LevelInfo, "INFO", msg, args...)
}

// Warn 输出警告日志
func (l *loggerImpl) Warn(msg string, args ...any) {
	l.log(LevelWarn, "WARN", msg, args...)
}

// Error 输出错误日志
func (l *loggerImpl) Error(msg string, args ...any) {
	l.log(LevelError, "ERROR", msg, args...)
}

// With 添加固定字段到日志
func (l *loggerImpl) With(args ...any) Logger {
	newFields := make(map[string]any)
	for k, v := range l.fields {
		newFields[k] = v
	}

	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			newFields[key] = args[i+1]
		}
	}

	return &loggerImpl{
		level:  l.level,
		format: l.format,
		output: l.output,
		fields: newFields,
	}
}

// WithContext 从 context 提取追踪信息并添加到日志
func (l *loggerImpl) WithContext(ctx context.Context) Logger {
	// 从 context 获取追踪 ID 等信息
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return l.With("trace_id", traceID)
	}
	return l
}

// 包级便捷函数

// Debug 输出调试日志（使用全局实例）
func Debug(msg string, args ...any) {
	Get().Debug(msg, args...)
}

// Info 输出信息日志（使用全局实例）
func Info(msg string, args ...any) {
	Get().Info(msg, args...)
}

// Warn 输出警告日志（使用全局实例）
func Warn(msg string, args ...any) {
	Get().Warn(msg, args...)
}

// Error 输出错误日志（使用全局实例）
func Error(msg string, args ...any) {
	Get().Error(msg, args...)
}

// With 添加固定字段（使用全局实例）
func With(args ...any) Logger {
	return Get().With(args...)
}

// WithContext 添加上下文信息（使用全局实例）
func WithContext(ctx context.Context) Logger {
	return Get().WithContext(ctx)
}

// LogEntry 用于中间件记录请求日志
type LogEntry struct {
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	Status     int           `json:"status"`
	Duration   time.Duration `json:"duration_ms"`
	ClientIP   string        `json:"client_ip"`
	UserAgent  string        `json:"user_agent"`
	Error      string        `json:"error,omitempty"`
	EmployeeID string        `json:"employee_id,omitempty"`
}

// LogRequest 记录 HTTP 请求日志
func LogRequest(entry LogEntry) {
	logger := Get()
	args := []any{
		"method", entry.Method,
		"path", entry.Path,
		"status", entry.Status,
		"duration_ms", entry.Duration.Milliseconds(),
		"client_ip", entry.ClientIP,
		"user_agent", entry.UserAgent,
	}
	if entry.Error != "" {
		args = append(args, "error", entry.Error)
	}
	if entry.EmployeeID != "" {
		args = append(args, "employee_id", entry.EmployeeID)
	}

	// 根据状态码选择日志级别
	switch {
	case entry.Status >= 500:
		logger.Error("HTTP请求错误", args...)
	case entry.Status >= 400:
		logger.Warn("HTTP请求异常", args...)
	default:
		logger.Info("HTTP请求", args...)
	}
}

// SetupStdLog 设置标准库 log 输出
func SetupStdLog() {
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags)
}
