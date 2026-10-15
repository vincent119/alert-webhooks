package middleware

import (
	"alert-webhooks/config"
	"github.com/gin-gonic/gin"
)

// BasicAuth 基本認證中間件
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		webhooks := config.Current().Webhooks
		// 每個請求取得完整快照，避免認證設定跨欄位讀到不同版本。
		// 帳密皆空表示採用免認證的既有 Webhook 設定契約。
		if !webhooks.Enable || (webhooks.BaseAuthUser == "" && webhooks.BaseAuthPassword == "") {
			c.Next()
			return
		}

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
		if username != webhooks.BaseAuthUser || password != webhooks.BaseAuthPassword {
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
