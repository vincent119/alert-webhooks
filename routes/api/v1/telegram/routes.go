// Package telegram 提供 Telegram 相關的路由處理功能，包括訊息發送和機器人資訊查詢
package telegram

import (
	"alert-webhooks/pkg/middleware"
	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 註冊 Telegram 相關路由
func RegisterRoutes(router *gin.RouterGroup, services service.Accessor) {
	// 創建 Telegram 路由處理器
	handler := NewHandler(services)

	// Telegram 相關路由（需要基本認證）
	router.POST("/telegram/chatid_:chatid", middleware.BasicAuth(), withServiceLease(services), handler.SendMessage)
	router.GET("/telegram/info", middleware.BasicAuth(), withServiceLease(services), handler.GetBotInfo)

}

func withServiceLease(services service.Accessor) gin.HandlerFunc {
	return func(c *gin.Context) {
		lease := services.Acquire()
		if lease == nil || lease.Telegram() == nil {
			c.AbortWithStatusJSON(503, gin.H{"success": false, "message": "Telegram service is not available"})
			return
		}
		c.Set("telegram_service_lease", lease)
		defer lease.Release()
		c.Next()
	}
}
