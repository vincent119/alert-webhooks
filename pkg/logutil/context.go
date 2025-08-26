package logutil

import (
	"context"
	"alert-webhooks/pkg/logcore"
)

// 定義 context key
type contextKey string

const loggerContextKey = contextKey("logger_fields")

// WithContext 將字段添加到上下文
func WithContext(ctx context.Context, fields ...logcore.Field) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	// 如果沒有字段，直接返回原始 context
	if len(fields) == 0 {
		return ctx
	}

	// 獲取現有字段
	var existingFields []logcore.Field
	if val := ctx.Value(loggerContextKey); val != nil {
		if existing, ok := val.([]logcore.Field); ok {
			existingFields = existing
		}
	}

	// 如果沒有現有字段，直接使用新字段
	if len(existingFields) == 0 {
		return context.WithValue(ctx, loggerContextKey, fields)
	}

	// 合併字段 - 預先分配確切大小的切片以提高效率
	newFields := make([]logcore.Field, len(existingFields)+len(fields))
	copy(newFields, existingFields)
	copy(newFields[len(existingFields):], fields)

	return context.WithValue(ctx, loggerContextKey, newFields)
}

// FromContext 從上下文中提取字段
func FromContext(ctx context.Context) []logcore.Field {
	if ctx == nil {
		return nil
	}

	if val := ctx.Value(loggerContextKey); val != nil {
		if fields, ok := val.([]logcore.Field); ok {
			return fields
		}
	}
	return nil
}

// DebugContext 使用上下文記錄調試信息
func DebugContext(ctx context.Context, msg string, fields ...logcore.Field) {
	// 檢查日誌系統是否初始化
	if logcore.Log == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// 合併上下文中的字段和提供的字段
	ctxFields := FromContext(ctx)
	if len(ctxFields) > 0 {
		// 創建足夠大的切片一次性分配內存
		allFields := make([]logcore.Field, len(ctxFields)+len(fields))
		copy(allFields, ctxFields)
		copy(allFields[len(ctxFields):], fields)
		logcore.Debug(msg, "default", allFields...)
	} else {
		logcore.Debug(msg, "default", fields...)
	}
}

// InfoContext 使用上下文記錄信息
func InfoContext(ctx context.Context, msg string, fields ...logcore.Field) {
	// 檢查日誌系統是否初始化
	if logcore.Log == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// 合併上下文中的字段和提供的字段
	ctxFields := FromContext(ctx)
	if len(ctxFields) > 0 {
		// 創建足夠大的切片一次性分配內存
		allFields := make([]logcore.Field, len(ctxFields)+len(fields))
		copy(allFields, ctxFields)
		copy(allFields[len(ctxFields):], fields)
		logcore.Info(msg, "default", allFields...)
	} else {
		logcore.Info(msg, "default", fields...)
	}
}

// WarnContext 使用上下文記錄警告信息
func WarnContext(ctx context.Context, msg string, fields ...logcore.Field) {
	// 檢查日誌系統是否初始化
	if logcore.Log == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// 合併上下文中的字段和提供的字段
	ctxFields := FromContext(ctx)
	if len(ctxFields) > 0 {
		// 創建足夠大的切片一次性分配內存
		allFields := make([]logcore.Field, len(ctxFields)+len(fields))
		copy(allFields, ctxFields)
		copy(allFields[len(ctxFields):], fields)
		logcore.Warn(msg, "default", allFields...)
	} else {
		logcore.Warn(msg, "default", fields...)
	}
}

// ErrorContext 使用上下文記錄錯誤信息
func ErrorContext(ctx context.Context, msg string, fields ...logcore.Field) {
	// 檢查日誌系統是否初始化
	if logcore.Log == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// 合併上下文中的字段和提供的字段
	ctxFields := FromContext(ctx)
	if len(ctxFields) > 0 {
		// 創建足夠大的切片一次性分配內存
		allFields := make([]logcore.Field, len(ctxFields)+len(fields))
		copy(allFields, ctxFields)
		copy(allFields[len(ctxFields):], fields)
		logcore.Error(msg, "default", allFields...)
	} else {
		logcore.Error(msg, "default", fields...)
	}
}

// FatalContext 使用上下文記錄致命錯誤
func FatalContext(ctx context.Context, msg string, fields ...logcore.Field) {
	// 檢查日誌系統是否初始化
	if logcore.Log == nil {
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	// 合併上下文中的字段和提供的字段
	ctxFields := FromContext(ctx)
	if len(ctxFields) > 0 {
		// 創建足夠大的切片一次性分配內存
		allFields := make([]logcore.Field, len(ctxFields)+len(fields))
		copy(allFields, ctxFields)
		copy(allFields[len(ctxFields):], fields)
		logcore.Fatal(msg, "default", allFields...)
	} else {
		logcore.Fatal(msg, "default", fields...)
	}
}

// WithRequestID 將請求 ID 添加到上下文
func WithRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	return WithContext(ctx, logcore.String("request_id", requestID))
}

// WithUserID 將用戶 ID 添加到上下文
func WithUserID(ctx context.Context, userID interface{}) context.Context {
	if userID == nil {
		return ctx
	}
	return WithContext(ctx, logcore.Any("user_id", userID))
}

// WithTraceID 將追蹤 ID 添加到上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	if traceID == "" {
		return ctx
	}
	return WithContext(ctx, logcore.String("trace_id", traceID))
}

// WithOperation 將操作名稱添加到上下文
func WithOperation(ctx context.Context, operation string) context.Context {
	if operation == "" {
		return ctx
	}
	return WithContext(ctx, logcore.String("operation", operation))
}

// WithComponent 將組件名稱添加到上下文
func WithComponent(ctx context.Context, component string) context.Context {
	if component == "" {
		return ctx
	}
	return WithContext(ctx, logcore.String("component", component))
}

// 添加輔助函數，減少重複代碼
func mergeContextFields(ctx context.Context, fields []logcore.Field) []logcore.Field {
	ctxFields := FromContext(ctx)
	if len(ctxFields) == 0 {
		return fields
	}

	allFields := make([]logcore.Field, len(ctxFields)+len(fields))
	copy(allFields, ctxFields)
	copy(allFields[len(ctxFields):], fields)
	return allFields
}
