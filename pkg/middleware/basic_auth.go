package middleware

import (
	"crypto/sha256"
	"crypto/subtle"

	"alert-webhooks/config"
	"github.com/gin-gonic/gin"
)

// BasicAuth 基本認證中間件
func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		webhooks := config.Current().Webhooks
		// 每個請求取得完整快照，避免認證設定跨欄位讀到不同版本。
		// 任一帳密空值表示採用免認證的既有 Webhook 設定契約。
		if !webhooks.Enable || webhooks.BaseAuthUser == "" || webhooks.BaseAuthPassword == "" {
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

		// 帳密皆完成固定時間比較後才判斷，避免任一欄位失敗時提早返回。
		if !credentialsMatch(username, password, webhooks.BaseAuthUser, webhooks.BaseAuthPassword) {
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

func credentialsMatch(username, password, expectedUsername, expectedPassword string) bool {
	usernameHash := sha256.Sum256([]byte(username))
	passwordHash := sha256.Sum256([]byte(password))
	expectedUsernameHash := sha256.Sum256([]byte(expectedUsername))
	expectedPasswordHash := sha256.Sum256([]byte(expectedPassword))

	usernameMatches := subtle.ConstantTimeCompare(usernameHash[:], expectedUsernameHash[:])
	passwordMatches := subtle.ConstantTimeCompare(passwordHash[:], expectedPasswordHash[:])
	return usernameMatches&passwordMatches == 1
}

// BasicAuthWithConfig 使用自定義配置的基本認證中間件
func BasicAuthWithConfig(username, password string) gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts{
		username: password,
	})
}
