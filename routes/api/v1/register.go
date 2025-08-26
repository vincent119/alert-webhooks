package v1

import (
	"alert-webhooks/pkg/logger"
	v1discord "alert-webhooks/routes/api/v1/discord"
	v1slack "alert-webhooks/routes/api/v1/slack"
	v1telegram "alert-webhooks/routes/api/v1/telegram"

	"github.com/gin-gonic/gin"
)

// RegisterApiV1Routes 註冊所有 API V1 路由組
func RegisterApiV1Routes(router *gin.RouterGroup) {
	// 健康檢查路由
	router.GET("/healthz", HealthCheck)
	
	// 註冊 Telegram 路由
	v1telegram.RegisterRoutes(router)
	
	// 註冊 Slack 路由
	v1slack.RegisterSlackRoutes(router)
	
	// 註冊 Discord 路由
	v1discord.SetupRoutes(router)
	
	logger.Info("API V1 routes registered successfully", "routes")
}
