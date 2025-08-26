package middleware

import (
	"alert-webhooks/pkg/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/commons/timex"
	"github.com/vincent119/commons/uuidx"
)

// Logger 日誌中間件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 開始時間
		start := time.Now()

		// 生成請求 ID
		requestID := uuidx.NewUUIDv4()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)

		// 使用 logger.WithContext 將 requestID 添加到 context
		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)

		// 處理請求
		c.Next()

		// 請求結束，記錄詳細信息
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		
		// 使用 Context 版本的日誌記錄函數
		if len(c.Errors) > 0 {
			logger.ErrorContext(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("query", query),
				logger.String("client_ip", clientIP),
				logger.Int("status", status),
				logger.Duration("latency", latency),
				logger.String("timestamp", timex.TimeStamp()),
				logger.String("error", c.Errors.String()),
			)
		} else {
			logger.InfoContext(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("query", query),
				logger.String("client_ip", clientIP),
				logger.Int("status", status),
				logger.Duration("latency", latency),
				logger.String("timestamp", timex.TimeStamp()),
			)
		}
	}
}



// LoggerWithSkipPaths 支援跳過特定路徑的日誌中間件
func LoggerWithSkipPaths(skipPaths []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 檢查是否需要跳過此路徑
		path := c.Request.URL.Path
		for _, skipPath := range skipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// 原有的日誌記錄邏輯
		start := time.Now()
		requestID := uuidx.NewUUIDv4()
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		
		ctx := logger.WithRequestID(c.Request.Context(), requestID)
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
		
		// 記錄日誌
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		query := c.Request.URL.RawQuery
		
		if len(c.Errors) > 0 {
			logger.ErrorContext(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("query", query),
				logger.String("client_ip", clientIP),
				logger.Int("status", status),
				logger.Duration("latency", latency),
				logger.String("timestamp", timex.TimeStamp()),
				logger.String("error", c.Errors.String()),
			)
		} else {
			logger.InfoContext(ctx, "HTTP request",
				logger.String("method", method),
				logger.String("path", path),
				logger.String("query", query),
				logger.String("client_ip", clientIP),
				logger.Int("status", status),
				logger.Duration("latency", latency),
				logger.String("timestamp", timex.TimeStamp()),
			)
		}
	}
}