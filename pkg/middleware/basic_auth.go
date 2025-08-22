package middleware

import (
	"alert-webhooks/config"
	"github.com/gin-gonic/gin"
)

// BasicAuth 基本認證中間件
func BasicAuth() gin.HandlerFunc {
	// 檢查是否啟用 webhooks 認證
	if !config.Webhooks.Enable {
		// 如果未啟用，返回空的中間件（不進行認證）
		return func(c *gin.Context) {
			c.Next()
		}
	}
	
	// 如果啟用，使用自定義的基本認證
	return func(c *gin.Context) {
		// 檢查 Authorization header
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
			c.JSON(401, gin.H{
				"error":   "Unauthorized",
				"message": "Basic authentication required",
				"code":    401,
			})
			c.Abort()
			return
		}

		// 解析 Basic Auth
		username, password, ok := c.Request.BasicAuth()
		if !ok {
			c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
			c.JSON(401, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid authentication format",
				"code":    401,
			})
			c.Abort()
			return
		}

		// 驗證用戶名和密碼
		if username != config.Webhooks.BaseAuthUser || password != config.Webhooks.BaseAuthPassword {
			c.Header("WWW-Authenticate", "Basic realm=Authorization Required")
			c.JSON(401, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid username or password",
				"code":    401,
			})
			c.Abort()
			return
		}

		// 認證成功，繼續處理請求
		c.Next()
	}
}

// BasicAuthWithConfig 使用自定義配置的基本認證中間件
func BasicAuthWithConfig(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}
