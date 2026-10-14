package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DefaultRequestBodyLimit 是入站 request body 的預設上限，為 1 MiB。
const DefaultRequestBodyLimit int64 = 1 << 20

// RequestBodyLimit 在 JSON 解析前限制入站 request body，避免未受限讀取耗盡服務資源。
func RequestBodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 || c.Request.Body == nil || c.Request.ContentLength == 0 {
			c.Next()
			return
		}

		if c.Request.ContentLength > maxBytes {
			_ = c.Request.Body.Close()
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "請求內容超過大小上限",
			})
			return
		}

		body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBytes+1))
		closeErr := c.Request.Body.Close()
		if err != nil || closeErr != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "無法讀取請求內容",
			})
			return
		}

		if int64(len(body)) > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error": "請求內容超過大小上限",
			})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewReader(body))
		c.Next()
	}
}
