package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/commons/uuidx"
)

// GinLogger 提供 Gin 框架的日誌中間件
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 生成請求 ID 並添加到請求上下文
		requestID := uuidx.NewUUIDv4()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// 創建帶請求ID的上下文
		ctx := WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// 繼續處理請求
		c.Next()

		// 計算請求時間
		latency := time.Since(start)

		// 獲取客戶端 IP、HTTP 方法、狀態碼和錯誤信息
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// 將請求信息記錄到日誌
		if len(c.Errors) > 0 {
			ErrorContext(ctx, "HTTP 請求",
				String("method", method),
				String("path", path),
				String("query", query),
				String("client_ip", clientIP),
				Int("status", statusCode),
				Duration("latency", latency),
				String("error", c.Errors.String()),
			)
		} else {
			InfoContext(ctx, "HTTP 請求",
				String("method", method),
				String("path", path),
				String("query", query),
				String("client_ip", clientIP),
				Int("status", statusCode),
				Duration("latency", latency),
			)
		}
	}
}
