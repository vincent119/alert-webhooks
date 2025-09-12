// Package telegram 提供 Telegram 相關的路由處理功能，包括訊息發送和機器人資訊查詢
package telegram

import (
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/middleware"
	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes 註冊 Telegram 相關路由
func RegisterRoutes(router *gin.RouterGroup) {
	// 獲取服務管理器
	serviceManager := service.GetServiceManager()
	
	// 檢查 Telegram 服務是否可用
	if serviceManager.IsTelegramServiceReady() {
		logger.Info("Telegram service is ready, creating handler", "telegram_routes")
		
		// 創建 Telegram 路由處理器
		handler := NewHandler(serviceManager.GetTelegramService())
		
		// Telegram 相關路由（需要基本認證）
		router.POST("/telegram/chatid_:chatid", middleware.BasicAuth(), handler.SendMessage)
		router.GET("/telegram/info", middleware.BasicAuth(), handler.GetBotInfo)
		
		logger.Info("Telegram routes registered successfully", "telegram_routes")
	} else {
		logger.Warn("Telegram service not available, skipping Telegram routes", "telegram_routes")
	}
}
