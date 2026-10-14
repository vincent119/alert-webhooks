package discord

import (
	"alert-webhooks/pkg/middleware"
	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures Discord API routes
func SetupRoutes(router *gin.RouterGroup, services service.Accessor) {
	// Get Discord service from service manager
	//serviceManager := service.GetServiceManager()
	//if serviceManager == nil {
	//	return // Service manager not initialized, skip route registration
	//}

	//discordService := serviceManager.GetDiscordService()
	// Note: Create handler even if service is nil - handler will check service availability
	handler := NewHandler(services)

	// Discord routes (consistent with Telegram/Slack pattern)
	discordRoutes := router.Group("/discord")
	{
		discordRoutes.Use(middleware.BasicAuth())
		discordRoutes.Use(withServiceLease(services))

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

func withServiceLease(services service.Accessor) gin.HandlerFunc {
	return func(c *gin.Context) {
		lease := services.Acquire()
		if lease == nil || lease.Discord() == nil {
			c.AbortWithStatusJSON(503, gin.H{"success": false, "message": "Discord service is not available"})
			return
		}
		c.Set("discord_service_lease", lease)
		defer lease.Release()
		c.Next()
	}
}
