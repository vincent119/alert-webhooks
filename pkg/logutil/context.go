package logutil

import (
	"context"

	"github.com/vincent119/zlogger"
)

// WithContext 將欄位加入 context。
func WithContext(ctx context.Context, fields ...zlogger.Field) context.Context {
	return zlogger.WithContext(ctx, fields...)
}

// FromContext 取得 context 中的日誌欄位。
func FromContext(ctx context.Context) []zlogger.Field {
	return zlogger.FromContext(ctx)
}

// DebugContext 以 context 欄位記錄除錯日誌。
func DebugContext(ctx context.Context, msg string, fields ...zlogger.Field) {
	zlogger.DebugContext(ctx, msg, withDefaultCategory(fields)...)
}

// InfoContext 以 context 欄位記錄資訊日誌。
func InfoContext(ctx context.Context, msg string, fields ...zlogger.Field) {
	zlogger.InfoContext(ctx, msg, withDefaultCategory(fields)...)
}

// WarnContext 以 context 欄位記錄警告日誌。
func WarnContext(ctx context.Context, msg string, fields ...zlogger.Field) {
	zlogger.WarnContext(ctx, msg, withDefaultCategory(fields)...)
}

// ErrorContext 以 context 欄位記錄錯誤日誌。
func ErrorContext(ctx context.Context, msg string, fields ...zlogger.Field) {
	zlogger.ErrorContext(ctx, msg, withDefaultCategory(fields)...)
}

// FatalContext 以 context 欄位記錄致命日誌。
func FatalContext(ctx context.Context, msg string, fields ...zlogger.Field) {
	zlogger.FatalContext(ctx, msg, withDefaultCategory(fields)...)
}

func withDefaultCategory(fields []zlogger.Field) []zlogger.Field {
	allFields := make([]zlogger.Field, 0, len(fields)+1)
	allFields = append(allFields, fields...)
	allFields = append(allFields, zlogger.String("category", "default"))
	return allFields
}

// WithRequestID 將請求 ID 加入 context。
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return zlogger.WithRequestID(ctx, requestID)
}

// WithUserID 將使用者 ID 加入 context。
func WithUserID(ctx context.Context, userID interface{}) context.Context {
	return zlogger.WithUserID(ctx, userID)
}

// WithTraceID 將追蹤 ID 加入 context。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return zlogger.WithTraceID(ctx, traceID)
}

// WithOperation 將操作名稱加入 context。
func WithOperation(ctx context.Context, operation string) context.Context {
	return zlogger.WithOperation(ctx, operation)
}

// WithComponent 將元件名稱加入 context。
func WithComponent(ctx context.Context, component string) context.Context {
	return zlogger.WithComponent(ctx, component)
}
