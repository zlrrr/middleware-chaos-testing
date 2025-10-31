package middleware

import (
	"fmt"
	"log"
	"time"
)

// Logger 中间件日志记录器
type Logger struct {
	prefix string
	debug  bool
}

// NewLogger 创建新的日志记录器
func NewLogger(prefix string, debug bool) *Logger {
	return &Logger{
		prefix: prefix,
		debug:  debug,
	}
}

// Info 记录信息日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", format, args...)
}

// Error 记录错误日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", format, args...)
}

// Debug 记录调试日志
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.debug {
		l.log("DEBUG", format, args...)
	}
}

// Warn 记录警告日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log("WARN", format, args...)
}

// log 内部日志记录方法
func (l *Logger) log(level, format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	log.Printf("[%s] [%s] [%s] %s", timestamp, level, l.prefix, message)
}

// OperationLog 操作日志结构
type OperationLog struct {
	Timestamp   time.Time              `json:"timestamp"`
	Operation   string                 `json:"operation"`
	Key         string                 `json:"key"`
	Success     bool                   `json:"success"`
	Duration    time.Duration          `json:"duration"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// LogOperation 记录操作日志
func (l *Logger) LogOperation(opLog *OperationLog) {
	if opLog.Success {
		l.Info("Operation succeeded: op=%s key=%s duration=%v metadata=%v",
			opLog.Operation, opLog.Key, opLog.Duration, opLog.Metadata)
	} else {
		l.Error("Operation failed: op=%s key=%s duration=%v error=%s metadata=%v",
			opLog.Operation, opLog.Key, opLog.Duration, opLog.Error, opLog.Metadata)
	}
}
