package discord

import (
	"github.com/gin-gonic/gin"
	"alert-webhooks/pkg/service"
)

// SetupRoutes configures Discord API routes
func SetupRoutes(router *gin.RouterGroup) {
	// Get Discord service from service manager
	serviceManager := service.GetServiceManager()
	if serviceManager == nil {
		return // Service manager not initialized, skip route registration
	}

	discordService := serviceManager.GetDiscordService()
	// Note: Create handler even if service is nil - handler will check service availability
	handler := NewHandler(discordService)

	// Discord routes (consistent with Telegram/Slack pattern)
	discordRoutes := router.Group("/discord")
	{
		// Send message to specific channel
		discordRoutes.POST("/channel/:channel", handler.SendMessageToChannel)
		
		// Send message to level-based channel (L0, L1, L2, etc.)
		discordRoutes.POST("/chatid_L:level", handler.SendMessageToLevel)
		
		// Service status and validation
		discordRoutes.GET("/status", handler.GetStatus)
		discordRoutes.POST("/test/:channel", handler.TestChannel)
		discordRoutes.POST("/validate/:channel", handler.ValidateChannel)
	}
}
