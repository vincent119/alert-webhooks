package discord

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/notification/types"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/template"

	"github.com/gin-gonic/gin"
)

// Handler handles Discord API requests
type Handler struct {
	discordService *service.DiscordService
}

// NewHandler creates a new Discord handler
func NewHandler(discordService *service.DiscordService) *Handler {
	return &Handler{
		discordService: discordService,
	}
}

// SendMessageRequest send message request structure
type SendMessageRequest struct {
	Message          string                 `json:"message,omitempty"`           // Simple text message
	AlertManagerData map[string]interface{} `json:"alertmanager_data,omitempty"` // AlertManager webhook data (wrapped format)
	
	// Direct AlertManager JSON format (consistent with Slack/Telegram)
	Receiver          string                   `json:"receiver,omitempty"`
	Status            string                   `json:"status,omitempty"`
	Alerts            []map[string]interface{} `json:"alerts,omitempty"`
	GroupLabels       map[string]interface{}   `json:"groupLabels,omitempty"`
	CommonLabels      map[string]interface{}   `json:"commonLabels,omitempty"`
	CommonAnnotations map[string]interface{}   `json:"commonAnnotations,omitempty"`
	ExternalURL       string                   `json:"externalURL,omitempty"`
	Version           string                   `json:"version,omitempty"`
	GroupKey          string                   `json:"groupKey,omitempty"`
	TruncatedAlerts   int                      `json:"truncatedAlerts,omitempty"`
}

// SendMessageResponse send message response structure
type SendMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Level   string `json:"level"`
}

// StatusResponse status response structure
type StatusResponse struct {
	Service   string      `json:"service"`
	Status    string      `json:"status"`
	BotInfo   interface{} `json:"bot_info,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

// SendMessageToChannel handles sending messages to a specific Discord channel
// @Summary Send message to Discord channel
// @Description Send a message to a specific Discord channel
// @Tags discord
// @Accept json
// @Produce json
// @Param channel path string true "Discord Channel ID"
// @Param request body SendMessageRequest true "Message request"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /discord/channel/{channel} [post]
func (h *Handler) SendMessageToChannel(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Channel ID is required",
		})
		return
	}

	// Check if Discord service is available
	if h.discordService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Discord service is not available",
		})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request format: %s", err.Error()),
		})
		return
	}

	// Handle direct message
	if req.Message != "" {
		err := h.discordService.SendMessageToChannel(channel, req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %s", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Message sent successfully",
		})
		return
	}

	// Handle AlertManager data (wrapped format)
	if len(req.AlertManagerData) > 0 {
		alertDataBytes, err := json.Marshal(req.AlertManagerData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to process alertmanager_data: %s", err.Error()),
			})
			return
		}
		
		message, err := h.generateAlertManagerMessage(alertDataBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate message: %s", err.Error()),
			})
			return
		}

		err = h.discordService.SendMessageToChannel(channel, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %s", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Discord message sent successfully",
		})
		return
	}

	// Handle direct AlertManager format (consistent with Slack/Telegram)
	if len(req.Alerts) > 0 || req.Status != "" {
		// Convert direct format to AlertManagerData structure
		alertManagerData := types.AlertManagerData{
			Receiver:          req.Receiver,
			Status:            req.Status,
			Alerts:            req.Alerts,
			GroupLabels:       req.GroupLabels,
			CommonLabels:      req.CommonLabels,
			CommonAnnotations: req.CommonAnnotations,
			ExternalURL:       req.ExternalURL,
			Version:           req.Version,
			GroupKey:          req.GroupKey,
			TruncatedAlerts:   req.TruncatedAlerts,
		}
		
		alertDataBytes, err := json.Marshal(alertManagerData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to process direct AlertManager data: %s", err.Error()),
			})
			return
		}
		
		message, err := h.generateAlertManagerMessage(alertDataBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate message: %s", err.Error()),
			})
			return
		}

		err = h.discordService.SendMessageToChannel(channel, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %s", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Discord message sent successfully",
		})
		return
	}

	c.JSON(http.StatusBadRequest, SendMessageResponse{
		Success: false,
		Message: "Either message, alertmanager_data, or direct AlertManager format must be provided",
	})
}

// SendMessageToLevel handles sending messages to level-based channels
// @Summary Send message to Discord level channel
// @Description Send a message to a Discord channel based on alert level
// @Tags discord
// @Accept json
// @Produce json
// @Param level path string true "Alert Level (0-5)"
// @Param request body SendMessageRequest true "Message request"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /discord/chatid_L{level} [post]
func (h *Handler) SendMessageToLevel(c *gin.Context) {
	level := c.Param("level")
	if level == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Alert level is required",
		})
		return
	}

	// Check if Discord service is available
	if h.discordService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Discord service is not available",
		})
		return
	}

	// Convert numeric level to L{number} format (e.g., "0" -> "L0")
	levelKey := "L" + level

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid request format: %s", err.Error()),
		})
		return
	}

	// Handle direct message
	if req.Message != "" {
		err := h.discordService.SendMessage(levelKey, req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %s", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Successfully sent message to Discord",
			Level:   level,
		})
		return
	}

	// Handle AlertManager data (wrapped format)
	if len(req.AlertManagerData) > 0 {
		alertDataBytes, err := json.Marshal(req.AlertManagerData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to process alertmanager_data: %s", err.Error()),
			})
			return
		}
		
		message, err := h.generateAlertManagerMessage(alertDataBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate message: %s", err.Error()),
			})
			return
		}

		err = h.discordService.SendMessage(levelKey, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %s", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Successfully sent message to Discord",
			Level:   level,
		})
		return
	}

	// Handle direct AlertManager format (consistent with Slack/Telegram)
	if len(req.Alerts) > 0 || req.Status != "" {
		// Convert direct format to AlertManagerData structure
		alertManagerData := types.AlertManagerData{
			Receiver:          req.Receiver,
			Status:            req.Status,
			Alerts:            req.Alerts,
			GroupLabels:       req.GroupLabels,
			CommonLabels:      req.CommonLabels,
			CommonAnnotations: req.CommonAnnotations,
			ExternalURL:       req.ExternalURL,
			Version:           req.Version,
			GroupKey:          req.GroupKey,
			TruncatedAlerts:   req.TruncatedAlerts,
		}
		
		alertDataBytes, err := json.Marshal(alertManagerData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to process direct AlertManager data: %s", err.Error()),
			})
			return
		}
		
		message, err := h.generateAlertManagerMessage(alertDataBytes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to generate message: %s", err.Error()),
			})
			return
		}

		err = h.discordService.SendMessage(levelKey, message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, SendMessageResponse{
				Success: false,
				Message: fmt.Sprintf("Failed to send message: %s", err.Error()),
			})
			return
		}

		c.JSON(http.StatusOK, SendMessageResponse{
			Success: true,
			Message: "Successfully sent message to Discord",
			Level:   level,
		})
		return
	}

	c.JSON(http.StatusBadRequest, SendMessageResponse{
		Success: false,
		Message: "Either message, alertmanager_data, or direct AlertManager format must be provided",
	})
}

// GetStatus returns Discord service status
// @Summary Get Discord service status
// @Description Get the status of Discord service and bot information
// @Tags discord
// @Produce json
// @Success 200 {object} StatusResponse
// @Failure 500 {object} StatusResponse
// @Router /discord/status [get]
func (h *Handler) GetStatus(c *gin.Context) {
	response := StatusResponse{
		Service:   "discord",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Check if Discord service is available
	if h.discordService == nil {
		response.Status = "unavailable"
		response.Error = "Discord service is not initialized"
		c.JSON(http.StatusServiceUnavailable, response)
		return
	}

	// Test connection
	err := h.discordService.TestConnection()
	if err != nil {
		response.Status = "error"
		response.Error = err.Error()
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Get bot info
	botInfo, err := h.discordService.GetBotInfo()
	if err != nil {
		response.Status = "partial"
		response.Error = fmt.Sprintf("Connection OK but failed to get bot info: %s", err.Error())
	} else {
		response.Status = "healthy"
		response.BotInfo = botInfo
	}

	c.JSON(http.StatusOK, response)
}

// TestChannel tests sending a message to a specific channel
// @Summary Test Discord channel
// @Description Send a test message to verify channel accessibility
// @Tags discord
// @Param channel path string true "Discord Channel ID"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /discord/test/{channel} [post]
func (h *Handler) TestChannel(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Channel ID is required",
		})
		return
	}

	// Check if Discord service is available
	if h.discordService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Discord service is not available",
		})
		return
	}

	testMessage := fmt.Sprintf("🤖 Discord Test Message\n\nTimestamp: %s\nChannel: %s\n\nThis is a test message from Alert Webhooks Discord bot.",
		time.Now().UTC().Format("2006-01-02 15:04:05 UTC"), channel)

	err := h.discordService.SendMessageToChannel(channel, testMessage)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SendMessageResponse{
			Success: false,
			Message: fmt.Sprintf("❌ Test message failed: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: fmt.Sprintf("✅ Test message sent successfully to channel %s", channel),
	})
}

// ValidateChannel validates channel accessibility and bot permissions
// @Summary Validate Discord channel
// @Description Validate if the bot can access and send messages to the channel
// @Tags discord
// @Param channel path string true "Discord Channel ID"
// @Success 200 {object} SendMessageResponse
// @Failure 400 {object} SendMessageResponse
// @Failure 500 {object} SendMessageResponse
// @Router /discord/validate/{channel} [post]
func (h *Handler) ValidateChannel(c *gin.Context) {
	channel := c.Param("channel")
	if channel == "" {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: "Channel ID is required",
		})
		return
	}

	// Check if Discord service is available
	if h.discordService == nil {
		c.JSON(http.StatusServiceUnavailable, SendMessageResponse{
			Success: false,
			Message: "Discord service is not available",
		})
		return
	}

	err := h.discordService.ValidateChannel(channel)
	if err != nil {
		c.JSON(http.StatusBadRequest, SendMessageResponse{
			Success: false,
			Message: fmt.Sprintf("❌ Channel validation failed: %s", err.Error()),
		})
		return
	}

	c.JSON(http.StatusOK, SendMessageResponse{
		Success: true,
		Message: fmt.Sprintf("✅ Channel %s validation successful! Bot has access and can send messages", channel),
	})
}

// generateAlertManagerMessage generates a formatted message from AlertManager data
func (h *Handler) generateAlertManagerMessage(alertData json.RawMessage) (string, error) {
	// Parse AlertManager JSON
	var req types.AlertManagerData
	if err := json.Unmarshal(alertData, &req); err != nil {
		logger.Error("Failed to parse AlertManager data", "DiscordHandler", logger.String("error", err.Error()))
		return h.generateBuiltInMessage(alertData)
	}

		// Try to use template engine if available
	serviceManager := service.GetServiceManager()
	if serviceManager != nil {
		templateEngine := serviceManager.GetTemplateEngine()
		if templateEngine != nil {
			// Get template configuration
			templateLanguage := config.Conf.Discord.TemplateLanguage
			if templateLanguage == "" {
				templateLanguage = "tw" // Default to Traditional Chinese
			}

			// Convert AlertManagerData to TemplateData
			templateData, err := h.convertToTemplateData(req)
			if err != nil {
				logger.Error("Failed to convert AlertManager data to template data", "DiscordHandler", logger.String("error", err.Error()))
				return h.generateBuiltInMessage(alertData)
			}

			// Use template engine to render message for Discord platform
			message, err := templateEngine.RenderTemplateForPlatform(templateLanguage, "discord", *templateData)
			if err != nil {
				logger.Error("Failed to render template, falling back to built-in", "DiscordHandler", logger.String("error", err.Error()))
				return h.generateBuiltInMessage(alertData)
			}
			return message, nil
		}
	}

	// Fallback to built-in message generation
	return h.generateBuiltInMessage(alertData)
}

// generateBuiltInMessage generates a built-in formatted message when template rendering fails
func (h *Handler) generateBuiltInMessage(alertData json.RawMessage) (string, error) {
	var req types.AlertManagerData
	if err := json.Unmarshal(alertData, &req); err != nil {
		return "", fmt.Errorf("failed to parse AlertManager data: %w", err)
	}

	var message strings.Builder

	// Header
	message.WriteString("🚨 **Alert Notification**\n\n")

	// Get basic information
	var alertName, env, severity, namespace string
	if len(req.Alerts) > 0 {
		alert := req.Alerts[0]
		if labels, ok := alert["labels"].(map[string]interface{}); ok {
			alertName = h.getStringValue(labels["alertname"])
			env = h.getStringValue(labels["env"])
			severity = h.getStringValue(labels["severity"])
			namespace = h.getStringValue(labels["namespace"])
		}
	}

	// Use escapeText for dynamic content to ensure Discord compatibility
	message.WriteString(fmt.Sprintf("**Alert Name:** %s\n", h.escapeText(alertName)))
	message.WriteString(fmt.Sprintf("**Environment:** %s\n", h.escapeText(env)))
	message.WriteString(fmt.Sprintf("**Severity:** %s\n", h.escapeText(severity)))
	message.WriteString(fmt.Sprintf("**Namespace:** %s\n", h.escapeText(namespace)))
	message.WriteString(fmt.Sprintf("**Total Alerts:** %d\n", len(req.Alerts)))

	// Count firing and resolved alerts
	firingCount := 0
	resolvedCount := 0
	for _, alert := range req.Alerts {
		if status, ok := alert["status"].(string); ok {
			if status == "firing" {
				firingCount++
			} else if status == "resolved" {
				resolvedCount++
			}
		}
	}

	message.WriteString(fmt.Sprintf("**Firing:** %d\n", firingCount))
	message.WriteString(fmt.Sprintf("**Resolved:** %d\n", resolvedCount))

	// Add firing alerts summary
	if firingCount > 0 {
		message.WriteString("\n🔥 **Firing Alerts:**\n")
		for i, alert := range req.Alerts {
			if status, ok := alert["status"].(string); ok && status == "firing" && i < 5 { // Limit to first 5
				if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
					if summary := h.getStringValue(annotations["summary"]); summary != "" {
						message.WriteString(fmt.Sprintf("• %s\n", h.escapeText(summary)))
					}
				}
			}
		}
	}

	// Add resolved alerts summary
	if resolvedCount > 0 {
		message.WriteString("\n✅ **Resolved Alerts:**\n")
		for i, alert := range req.Alerts {
			if status, ok := alert["status"].(string); ok && status == "resolved" && i < 5 { // Limit to first 5
				if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
					if summary := h.getStringValue(annotations["summary"]); summary != "" {
						message.WriteString(fmt.Sprintf("• %s\n", h.escapeText(summary)))
					}
				}
			}
		}
	}

	// Add external URL
	if req.ExternalURL != "" {
		message.WriteString(fmt.Sprintf("\n🔗 [View Details](%s)", req.ExternalURL))
	}

	return message.String(), nil
}

// formatTime formats time with Discord-compatible format
func (h *Handler) formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// escapeText escapes special characters for Discord
func (h *Handler) escapeText(text string) string {
	if text == "" {
		return text
	}
	// Discord uses standard Markdown, so we need to escape basic Markdown characters
	text = strings.ReplaceAll(text, "*", "\\*")
	text = strings.ReplaceAll(text, "_", "\\_")
	text = strings.ReplaceAll(text, "`", "\\`")
	text = strings.ReplaceAll(text, "~", "\\~")
	text = strings.ReplaceAll(text, "|", "\\|")
	return text
}

// getStringValue safely extracts string value from interface{}
func (h *Handler) getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}

// convertToTemplateData converts AlertManagerData to TemplateData for template engine
func (h *Handler) convertToTemplateData(req types.AlertManagerData) (*template.TemplateData, error) {
	// Get FormatOptions based on Discord template mode configuration
	templateMode := config.Conf.Discord.TemplateMode
	if templateMode == "" {
		templateMode = "full" // Default to full mode
	}
	
	// Debug: log the template mode being used
	logger.Debug("Template mode configuration", "DiscordHandler", 
		logger.String("templateMode", templateMode),
		logger.String("templateLanguage", config.Conf.Discord.TemplateLanguage))
	
	// Get FormatOptions from template engine based on mode
	serviceManager := service.GetServiceManager()
	if serviceManager == nil {
		return nil, fmt.Errorf("service manager not available")
	}
	
	templateEngine := serviceManager.GetTemplateEngine()
	if templateEngine == nil {
		return nil, fmt.Errorf("template engine not available")
	}
	
	// Get FormatOptions from template engine based on mode (using config files)
	var formatOptions template.FormatOptions
	if templateMode == "minimal" {
		formatOptions = templateEngine.GetMinimalDefaultConfig().FormatOptions
		logger.Debug("Using minimal config FormatOptions", "DiscordHandler", 
			logger.String("templateMode", templateMode),
			logger.Bool("ShowEmoji", formatOptions.ShowEmoji.Enabled),
			logger.Bool("ShowTimestamps", formatOptions.ShowTimestamps.Enabled),
			logger.Bool("ShowGeneratorURL", formatOptions.ShowGeneratorURL.Enabled),
			logger.Bool("ShowExternalURL", formatOptions.ShowExternalURL.Enabled))
	} else {
		formatOptions = templateEngine.GetFullDefaultConfig().FormatOptions
		logger.Debug("Using full config FormatOptions", "DiscordHandler", 
			logger.String("templateMode", templateMode),
			logger.Bool("ShowEmoji", formatOptions.ShowEmoji.Enabled),
			logger.Bool("ShowTimestamps", formatOptions.ShowTimestamps.Enabled),
			logger.Bool("ShowGeneratorURL", formatOptions.ShowGeneratorURL.Enabled),
			logger.Bool("ShowExternalURL", formatOptions.ShowExternalURL.Enabled))
	}
	// Extract basic information from first alert
	var alertName, env, severity, namespace string
	if len(req.Alerts) > 0 {
		alert := req.Alerts[0]
		if labels, ok := alert["labels"].(map[string]interface{}); ok {
			alertName = h.getStringValue(labels["alertname"])
			env = h.getStringValue(labels["env"])
			severity = h.getStringValue(labels["severity"])
			namespace = h.getStringValue(labels["namespace"])
		}
	}

	// Count firing and resolved alerts
	firingCount := 0
	resolvedCount := 0
	var templateAlerts []template.AlertData
	
	for _, alert := range req.Alerts {
		status := h.getStringValue(alert["status"])
		if status == "firing" {
			firingCount++
		} else if status == "resolved" {
			resolvedCount++
		}

		// Convert to template.AlertData
		alertData := template.AlertData{
			Status:       status,
			Labels:       make(map[string]string),
			Annotations:  make(map[string]string),
			StartsAt:     h.getStringValue(alert["startsAt"]),
			EndsAt:       h.getStringValue(alert["endsAt"]),
			GeneratorURL: h.getStringValue(alert["generatorURL"]),
		}

		// Convert labels
		if labels, ok := alert["labels"].(map[string]interface{}); ok {
			for k, v := range labels {
				alertData.Labels[k] = h.getStringValue(v)
			}
		}

		// Convert annotations
		if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
			for k, v := range annotations {
				alertData.Annotations[k] = h.getStringValue(v)
			}
		}

		templateAlerts = append(templateAlerts, alertData)
	}

	// Create TemplateData
	templateData := &template.TemplateData{
		Status:        req.Status,
		AlertName:     alertName,
		Env:           env,
		Severity:      severity,
		Namespace:     namespace,
		TotalAlerts:   len(req.Alerts),
		FiringCount:   firingCount,
		ResolvedCount: resolvedCount,
		Alerts:        templateAlerts,
		ExternalURL:   req.ExternalURL,
		Platform:      "discord",
		FormatOptions: formatOptions, // 使用從模板引擎獲取的 FormatOptions
	}

	return templateData, nil
}

// getFormatOptionsForDiscord 根據 Discord 配置返回對應的 FormatOptions
func (h *Handler) getFormatOptionsForDiscord() template.FormatOptions {
	templateMode := config.Conf.Discord.TemplateMode
	if templateMode == "" {
		templateMode = "full" // Default to full mode
	}
	
	logger.Debug("getFormatOptionsForDiscord called", "DiscordHandler", 
		logger.String("templateMode", templateMode))
	
	if templateMode == "minimal" {
		// Minimal mode: disable most options
		return template.FormatOptions{
			ShowLinks: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示超連結"},
			ShowTimestamps: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示時間戳"},
			ShowExternalURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示外部連結"},
			ShowGeneratorURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示生成器連結"},
			ShowEmoji: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "是否顯示表情符號"},
			CompactMode: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "緊湊模式（簡化顯示）"},
		}
	} else {
		// Full mode: enable all options
		return template.FormatOptions{
			ShowLinks: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示超連結"},
			ShowTimestamps: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示時間戳"},
			ShowExternalURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示外部連結"},
			ShowGeneratorURL: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示生成器連結"},
			ShowEmoji: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: true, Description: "是否顯示表情符號"},
			CompactMode: struct {
				Enabled     bool   `yaml:"enabled"`
				Description string `yaml:"description"`
			}{Enabled: false, Description: "緊湊模式（簡化顯示）"},
		}
	}
}
