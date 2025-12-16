package v1

import (
	"alert-webhooks/pkg/logger"
	v1discord "alert-webhooks/routes/api/v1/discord"
	v1slack "alert-webhooks/routes/api/v1/slack"
	v1telegram "alert-webhooks/routes/api/v1/telegram"
  "alert-webhooks/pkg/service"
	"github.com/gin-gonic/gin"
)

// RegisterApiV1Routes 註冊所有 API V1 路由組
func RegisterApiV1Routes(router *gin.RouterGroup) {
	// 健康檢查路由
	router.GET("/healthz", HealthCheck)
	
	sm := service.GetServiceManager()
	if sm.IsTelegramServiceReady() {
		// 註冊 Telegram 路由
		v1telegram.RegisterRoutes(router, sm.GetTelegramService())
 	} else {
		logger.Info("Telegram service not ready, skipping Telegram routes", "routes")
	}


	if sm.IsSlackServiceReady() {
	// 註冊 Slack 路由
		v1slack.RegisterSlackRoutes(router, sm.GetSlackService())
	} else {
		logger.Info("Slack service not ready, skipping Slack routes", "routes")
	}
	
	// 註冊 Discord 路由
	if sm.IsDiscordServiceReady() {
		v1discord.SetupRoutes(router, sm.GetDiscordService())
	} else {
		logger.Info("Discord service not ready, skipping Discord routes", "routes")
	}
	
	logger.Info("API V1 routes registered successfully", "routes")
}
