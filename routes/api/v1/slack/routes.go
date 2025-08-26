package slack

import (
	"alert-webhooks/pkg/middleware"
	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

// RegisterSlackRoutes 註冊 Slack 路由
func RegisterSlackRoutes(r *gin.RouterGroup) {
	// 獲取服務管理器
	serviceManager := service.GetServiceManager()
	
	// 檢查 Slack 服務是否可用
	if !serviceManager.IsSlackServiceReady() {
		return
	}
	
	// 創建處理器
	handler := NewHandler(serviceManager.GetSlackService())

	// Slack 路由群組
	slackGroup := r.Group("/slack")
	{
		// 需要認證的路由
		slackGroup.Use(middleware.BasicAuth())
		
		// 發送訊息到指定頻道
		slackGroup.POST("/channel/:channel", handler.SendMessageToChannel)
		
		// 發送訊息到指定等級 (格式: /chatid_L0, /chatid_L1, etc.)
		slackGroup.POST("/chatid_L:level", handler.SendMessageToLevel)
		
		// 發送富文本訊息
		slackGroup.POST("/rich/:channel", handler.SendRichMessage)
		
		// 獲取 Slack 服務狀態
		slackGroup.GET("/status", handler.GetStatus)
		
		// 獲取頻道列表
		slackGroup.GET("/channels", handler.GetChannels)
		
		// 測試連接
		slackGroup.POST("/test", handler.TestConnection)
		
		// 驗證頻道
		slackGroup.POST("/validate/:channel", handler.ValidateChannel)
	}
}
