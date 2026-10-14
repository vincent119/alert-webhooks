package middleware

import (
	"strings"

	"alert-webhooks/config"

	"github.com/gin-gonic/gin"
)

// CORS 只允許目前設定 allowlist 中的 origin 使用跨來源請求。
func CORS() gin.HandlerFunc {
	return corsWithOriginProvider(func() []string {
		return config.Current().Webhooks.CORSAllowedOrigins
	})
}

func corsWithOriginProvider(allowedOrigins func() []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if !isAllowedOrigin(origin, allowedOrigins()) {
			c.Next()
			return
		}
		headers := c.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Add("Vary", "Origin")

		// 設置 CORS headers
		headers.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, User-Agent, Pragma, Referer, X-Forwarded-For, X-Real-Ip, Accept-Language, utoken, x-key")
		headers.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, utoken, x-key")
		headers.Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE, PATCH")
		headers.Set("Referrer-Policy", "origin")

		// 處理 preflight 請求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed != "" && allowed != "*" && origin == allowed {
			return true
		}
	}
	return false
}
