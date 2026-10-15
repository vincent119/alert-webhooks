package middleware

import (
	"fmt"
	//"going-admin-backend/internal/api"
	"github.com/vincent119/zlogger"
	//"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

var middlewareRecoveryString = "middleware-recovery"

// Recovery 是一個 Gin 中間件，用於捕獲和處理 panic
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 記錄堆疊信息
				stack := debug.Stack()

				httpRequest, _ := c.Get("HttpRequest")

				// 記錄詳細錯誤日誌
				zlogger.Error("[Recovery] Critical system error", zlogger.String("category", middlewareRecoveryString), zlogger.Any("error", err),
					zlogger.String("stack", string(stack)),
					zlogger.Any("request", httpRequest),
					zlogger.String("path", c.Request.URL.Path),
				)

				// 返回 500 錯誤給客戶端
				// c.AbortWithStatusJSON(http.StatusInternalServerError, api.Response{
				// 	Code:    http.StatusInternalServerError,
				// 	Message: "伺服器內部錯誤，請稍後再試",
				// 	Data:    nil,
				// })
			}
		}()

		// 處理請求
		c.Next()
	}
}

// RecoveryWithWriter 是一個可自定義日誌輸出的 Recovery 中間件
func RecoveryWithWriter(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 取得堆疊信息
				stackInfo := debug.Stack()

				// 格式化錯誤訊息
				errMsg := fmt.Sprintf("[Recovery] unexpected panic: %v", err)

				// 記錄日誌
				zlogger.Error(errMsg, zlogger.String("category", middlewareRecoveryString), zlogger.String("path", c.Request.URL.Path),
					zlogger.String("method", c.Request.Method),
					zlogger.String("client_ip", c.ClientIP()),
				)

				// 如果需要輸出堆疊信息
				if stack {
					zlogger.Error("stack trace", zlogger.String("category", middlewareRecoveryString), zlogger.String("stack", string(stackInfo)))
				}

				// 返回 500 錯誤給客戶端
				// c.AbortWithStatusJSON(http.StatusInternalServerError, api.Response{
				// 	Code:    http.StatusInternalServerError,
				// 	Message: "伺服器內部錯誤，請稍後再試",
				// 	Data:    nil,
				// })
			}
		}()

		c.Next()
	}
}
