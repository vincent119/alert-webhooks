package logger

import (
	"context"
	"strings"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logutil"

	"github.com/vincent119/zlogger"
)

// Field 是 zlogger.Field 的別名，讓既有呼叫端維持相容。
type Field = zlogger.Field

// Configure 依應用程式設定建立全域日誌，並回傳關閉資源的函式。
func Configure(logConfig config.LogConf, development bool) (func() error, error) {
	return zlogger.Configure(newConfigPatch(logConfig, development))
}

func newConfigPatch(logConfig config.LogConf, development bool) *zlogger.ConfigPatch {
	patch := &zlogger.ConfigPatch{
		AddCaller:     &logConfig.AddCaller,
		AddStacktrace: &logConfig.AddStacktrace,
		Development:   &development,
	}

	if level := supportedLevel(logConfig.Level); level != "" {
		patch.Level = &level
	}
	if format := supportedFormat(logConfig.Format); format != "" {
		patch.Format = &format
	}
	if logConfig.LogPath != "" {
		patch.LogPath = &logConfig.LogPath
	}
	if logConfig.LogFile != "" {
		patch.FileName = &logConfig.LogFile
	}
	if logConfig.Outputs != "" {
		outputs := supportedOutputs(logConfig.Outputs)
		if len(outputs) > 0 {
			patch.Outputs = &outputs
		}
	}

	return patch
}

func supportedLevel(level string) string {
	normalized := strings.ToLower(strings.TrimSpace(level))
	switch normalized {
	case "debug", "info", "warn", "error", "fatal":
		return normalized
	default:
		return ""
	}
}

func supportedFormat(format string) string {
	normalized := strings.ToLower(strings.TrimSpace(format))
	switch normalized {
	case "console", "json":
		return normalized
	default:
		return ""
	}
}

func supportedOutputs(rawOutputs string) []string {
	outputs := make([]string, 0, 2)
	seen := make(map[string]struct{}, 2)
	for _, output := range strings.Split(rawOutputs, ",") {
		normalized := strings.ToLower(strings.TrimSpace(output))
		if normalized != "console" && normalized != "file" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		outputs = append(outputs, normalized)
	}
	return outputs
}

// Debug 以分類欄位記錄除錯日誌。
func Debug(msg, category string, fields ...Field) {
	zlogger.Debug(msg, withCategory(category, fields)...)
}

// Info 以分類欄位記錄資訊日誌。
func Info(msg, category string, fields ...Field) {
	zlogger.Info(msg, withCategory(category, fields)...)
}

// Warn 以分類欄位記錄警告日誌。
func Warn(msg, category string, fields ...Field) {
	zlogger.Warn(msg, withCategory(category, fields)...)
}

// Error 以分類欄位記錄錯誤日誌。
func Error(msg, category string, fields ...Field) {
	zlogger.Error(msg, withCategory(category, fields)...)
}

// Fatal 以分類欄位記錄致命日誌。
func Fatal(msg, category string, fields ...Field) {
	zlogger.Fatal(msg, withCategory(category, fields)...)
}

func withCategory(category string, fields []Field) []Field {
	allFields := make([]Field, 0, len(fields)+1)
	allFields = append(allFields, zlogger.String("category", category))
	allFields = append(allFields, fields...)
	return allFields
}

// WithContext 將欄位加入 context。
func WithContext(ctx context.Context, fields ...Field) context.Context {
	return logutil.WithContext(ctx, fields...)
}

// FromContext 取得 context 中的日誌欄位。
func FromContext(ctx context.Context) []Field {
	return logutil.FromContext(ctx)
}

// DebugContext 以 context 欄位記錄除錯日誌。
func DebugContext(ctx context.Context, msg string, fields ...Field) {
	logutil.DebugContext(ctx, msg, fields...)
}

// InfoContext 以 context 欄位記錄資訊日誌。
func InfoContext(ctx context.Context, msg string, fields ...Field) {
	logutil.InfoContext(ctx, msg, fields...)
}

// WarnContext 以 context 欄位記錄警告日誌。
func WarnContext(ctx context.Context, msg string, fields ...Field) {
	logutil.WarnContext(ctx, msg, fields...)
}

// ErrorContext 以 context 欄位記錄錯誤日誌。
func ErrorContext(ctx context.Context, msg string, fields ...Field) {
	logutil.ErrorContext(ctx, msg, fields...)
}

// FatalContext 以 context 欄位記錄致命日誌。
func FatalContext(ctx context.Context, msg string, fields ...Field) {
	logutil.FatalContext(ctx, msg, fields...)
}

// WithRequestID 將請求 ID 加入 context。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return logutil.WithRequestID(ctx, requestID)
}

// WithUserID 將使用者 ID 加入 context。
func WithUserID(ctx context.Context, userID interface{}) context.Context {
	return logutil.WithUserID(ctx, userID)
}

// WithTraceID 將追蹤 ID 加入 context。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return logutil.WithTraceID(ctx, traceID)
}

// WithOperation 將操作名稱加入 context。
func WithOperation(ctx context.Context, operation string) context.Context {
	return logutil.WithOperation(ctx, operation)
}

// WithComponent 將元件名稱加入 context。
func WithComponent(ctx context.Context, component string) context.Context {
	return logutil.WithComponent(ctx, component)
}

// String 建立字串欄位。
func String(key, value string) Field {
	return zlogger.String(key, value)
}

// Int 建立整數欄位。
func Int(key string, value int) Field {
	return zlogger.Int(key, value)
}

// Int64 建立 64 位元整數欄位。
func Int64(key string, value int64) Field {
	return zlogger.Int64(key, value)
}

// Float64 建立浮點數欄位。
func Float64(key string, value float64) Field {
	return zlogger.Float64(key, value)
}

// Bool 建立布林欄位。
func Bool(key string, value bool) Field {
	return zlogger.Bool(key, value)
}

// Err 建立錯誤欄位。
func Err(err error) Field {
	return zlogger.Err(err)
}

// Any 建立任意型別欄位。
func Any(key string, value interface{}) Field {
	return zlogger.Any(key, value)
}

// Duration 建立時間長度欄位。
func Duration(key string, value time.Duration) Field {
	return zlogger.Duration(key, value)
}

// Time 建立時間欄位。
func Time(key string, value time.Time) Field {
	return zlogger.Time(key, value)
}

// Uint 建立無號整數欄位。
func Uint(key string, value uint) Field {
	return zlogger.Uint(key, value)
}

// Int8 建立 8 位元整數欄位。
func Int8(key string, value int8) Field {
	return zlogger.Int8(key, value)
}

// Uint64 建立 64 位元無號整數欄位。
func Uint64(key string, value uint64) Field {
	return zlogger.Uint64(key, value)
}

// SetLevel 更新全域日誌等級。
func SetLevel(level string) {
	zlogger.SetLevel(level)
}

// Sync 將日誌輸出同步至底層資源。
func Sync() error {
	return zlogger.Sync()
}

// GetLogger 回傳底層 zlogger。
func GetLogger() *zlogger.Logger {
	return zlogger.GetLogger()
}
