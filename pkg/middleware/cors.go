package middleware

import (
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin") // 安全获取 Origin
		if origin == "" {
			origin = "*"
		}
		// 僅允許白名單 origin
		if origin == "http://localhost:3000" || origin == "http://127.0.0.1:3000" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}

		headers := c.Writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Access-Control-Allow-Credentials", "true")
		headers.Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Accept, Origin, Cache-Control, X-Requested-With, User-Agent, Pragma, Referer, X-Forwarded-For, X-Real-Ip, Accept-Language, utoken, x-key")
		headers.Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, Expires, Last-Modified, utoken, x-key")
		headers.Set("Access-Control-Allow-Methods", "OPTIONS, GET, POST, PUT, DELETE, PATCH")
		headers.Set("Referrer-Policy", "origin")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
