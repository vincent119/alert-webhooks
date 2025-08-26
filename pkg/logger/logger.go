
package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"alert-webhooks/pkg/logcore"
	"alert-webhooks/pkg/logutil"
)

// Field 是 logcore.Field 的別名，使用時更簡潔
type Field = logcore.Field

var initialized = false

// 初始化函數代理
func InitLogger(level string, development bool) {
	logcore.InitLogger(level, development)
	if initialized {
		return
	}
	initialized = true
}

// 日誌函數代理
func Debug(msg, category string, fields ...Field) {
	logcore.Debug(msg, category, fields...)
}

func Info(msg, category string, fields ...Field) {
	logcore.Info(msg, category, fields...)
}

func Warn(msg, category string, fields ...Field) {
	logcore.Warn(msg, category, fields...)
}

func Error(msg, category string, fields ...Field) {
	logcore.Error(msg, category, fields...)
}

func Fatal(msg, category string, fields ...Field) {
	logcore.Fatal(msg, category, fields...)
}

// Context 相關函數代理
func WithContext(ctx context.Context, fields ...Field) context.Context {
	return logutil.WithContext(ctx, fields...)
}

func FromContext(ctx context.Context) []Field {
	return logutil.FromContext(ctx)
}

func DebugContext(ctx context.Context, msg string, fields ...Field) {
	logutil.DebugContext(ctx, msg, fields...)
}

func InfoContext(ctx context.Context, msg string, fields ...Field) {
	logutil.InfoContext(ctx, msg, fields...)
}

func WarnContext(ctx context.Context, msg string, fields ...Field) {
	logutil.WarnContext(ctx, msg, fields...)
}

func ErrorContext(ctx context.Context, msg string, fields ...Field) {
	logutil.ErrorContext(ctx, msg, fields...)
}

func FatalContext(ctx context.Context, msg string, fields ...Field) {
	logutil.FatalContext(ctx, msg, fields...)
}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return logutil.WithRequestID(ctx, requestID)
}

func WithUserID(ctx context.Context, userID interface{}) context.Context {
	return logutil.WithUserID(ctx, userID)
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return logutil.WithTraceID(ctx, traceID)
}

func WithOperation(ctx context.Context, operation string) context.Context {
	return logutil.WithOperation(ctx, operation)
}

func WithComponent(ctx context.Context, component string) context.Context {
	return logutil.WithComponent(ctx, component)
}

// Field 創建函數代理
func String(key, value string) Field {
	return logcore.String(key, value)
}

func Int(key string, value int) Field {
	return logcore.Int(key, value)
}

func Int64(key string, value int64) Field {
	return logcore.Int64(key, value)
}

func Float64(key string, value float64) Field {
	return logcore.Float64(key, value)
}

func Bool(key string, value bool) Field {
	return logcore.Bool(key, value)
}

func Err(err error) Field {
	return logcore.Err(err)
}

func Any(key string, value interface{}) Field {
	return logcore.Any(key, value)
}

func Duration(key string, value time.Duration) Field {
	return logcore.Duration(key, value)
}

func Time(key string, value time.Time) Field {
	return logcore.Time(key, value)
}

func SetLevel(level string) {
	logcore.SetLevel(level)
}

// GetLogger 返回原始 zap logger
func GetLogger() *zap.Logger {
	return logcore.Log
}

func Uint(key string, value uint) Field {
	return zap.Uint(key, value)
}

// Int8 添加 int8 類型的日誌字段
func Int8(key string, value int8) zap.Field {
	return zap.Int8(key, value)
}

// Uint64 creates a field with the given key and uint64 value
// Uint64 creates a field with the given key and uint64 value
func Uint64(key string, value uint64) Field {
	return zap.Uint64(key, value)
}
