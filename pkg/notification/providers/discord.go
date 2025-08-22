package providers

import (
	"context"
	"fmt"
	"strings"

	"alert-webhooks/pkg/notification/types"
)

// DiscordService interface to avoid circular dependencies
type DiscordService interface {
	SendMessage(level string, message string) error
	SendMessageToChannel(channelID, message string) error
	SendMessageToLevel(level string, message string) error
	TestConnection() error
	ValidateChannel(channelID string) error
	GetBotInfo() (interface{}, error)
}

// DiscordProvider implements NotificationProvider for Discord
type DiscordProvider struct {
	service        DiscordService
	templateEngine types.TemplateEngine
}

// NewDiscordProvider creates a new Discord notification provider
func NewDiscordProvider(service DiscordService, templateEngine types.TemplateEngine) types.NotificationProvider {
	return &DiscordProvider{
		service:        service,
		templateEngine: templateEngine,
	}
}

// GetName returns the provider name
func (dp *DiscordProvider) GetName() string {
	return "discord"
}

// GetCapabilities returns the provider capabilities
func (dp *DiscordProvider) GetCapabilities() *types.ProviderCapabilities {
	return &types.ProviderCapabilities{
		SupportsLevels:       true,
		SupportsChannels:     true,
		SupportsRichText:     true,
		SupportsAttachments:  false,
		MaxMessageLength:     2000,
		SupportedLanguages:   dp.templateEngine.GetSupportedLanguages(),
	}
}

// GetStatus returns the provider status
func (dp *DiscordProvider) GetStatus() *types.ProviderStatus {
	err := dp.service.TestConnection()
	if err != nil {
		return &types.ProviderStatus{
			Name:      "discord",
			Enabled:   true,
			Connected: false,
			LastError: err.Error(),
			Statistics: &types.ProviderStats{},
		}
	}

	return &types.ProviderStatus{
		Name:      "discord",
		Enabled:   true,
		Connected: true,
		LastError: "",
		Statistics: &types.ProviderStats{},
	}
}

// SendMessage sends a notification message
func (dp *DiscordProvider) SendMessage(ctx context.Context, req *types.NotificationRequest) error {
	// Use direct message if provided
	if req.Message != "" {
		if req.Channel != "" {
			return dp.service.SendMessageToChannel(req.Channel, req.Message)
		} else if req.Level != "" {
			return dp.service.SendMessage(req.Level, req.Message)
		} else {
			return fmt.Errorf("either channel or level must be specified")
		}
	} else if req.AlertData != nil {
		// Render template message
		var message string
		var err error
		if req.Channel != "" {
			message, err = dp.renderAlertMessage(*req)
			if err != nil {
				return err
			}
			return dp.service.SendMessageToChannel(req.Channel, message)
		} else if req.Level != "" {
			message, err = dp.renderAlertMessage(*req)
			if err != nil {
				return err
			}
			return dp.service.SendMessage(req.Level, message)
		} else {
			return fmt.Errorf("either channel or level must be specified")
		}
	} else {
		return fmt.Errorf("either message or alert data must be provided")
	}
}

// IsEnabled returns if the provider is enabled
func (dp *DiscordProvider) IsEnabled() bool {
	return true // Discord provider is enabled if it's initialized
}

// ValidateConfig validates the provider configuration
func (dp *DiscordProvider) ValidateConfig() error {
	return dp.service.TestConnection()
}

// TestConnection tests the provider connection
func (dp *DiscordProvider) TestConnection() error {
	return dp.service.TestConnection()
}

// renderAlertMessage renders alert data using the template engine
func (dp *DiscordProvider) renderAlertMessage(req types.NotificationRequest) (string, error) {
	// For now, return a basic formatted message
	// TODO: Integrate with template engine when available
	return dp.formatBasicMessage(req.AlertData), nil
}

// formatBasicMessage creates a basic Discord-formatted message from alert data
func (dp *DiscordProvider) formatBasicMessage(alertData *types.AlertManagerData) string {
	var message strings.Builder
	
	// Header
	message.WriteString("🚨 **Alert Notification**\n\n")
	
	if len(alertData.Alerts) > 0 {
		alert := alertData.Alerts[0]
		
		if labels, ok := alert["labels"].(map[string]interface{}); ok {
			if alertName := getStringValue(labels["alertname"]); alertName != "" {
				message.WriteString(fmt.Sprintf("**Alert Name:** %s\n", alertName))
			}
			if env := getStringValue(labels["env"]); env != "" {
				message.WriteString(fmt.Sprintf("**Environment:** %s\n", env))
			}
			if severity := getStringValue(labels["severity"]); severity != "" {
				message.WriteString(fmt.Sprintf("**Severity:** %s\n", severity))
			}
			if namespace := getStringValue(labels["namespace"]); namespace != "" {
				message.WriteString(fmt.Sprintf("**Namespace:** %s\n", namespace))
			}
		}
		
		message.WriteString(fmt.Sprintf("**Status:** %s\n", alertData.Status))
		message.WriteString(fmt.Sprintf("**Total Alerts:** %d\n", len(alertData.Alerts)))
		
		// Count firing and resolved alerts
		firingCount := 0
		resolvedCount := 0
		for _, alert := range alertData.Alerts {
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
		
		// Add summary of firing alerts
		if firingCount > 0 {
			message.WriteString("\n🔥 **Firing Alerts:**\n")
			for i, alert := range alertData.Alerts {
				if status, ok := alert["status"].(string); ok && status == "firing" && i < 5 { // Limit to first 5
					if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
						if summary := getStringValue(annotations["summary"]); summary != "" {
							message.WriteString(fmt.Sprintf("• %s\n", summary))
						}
					}
				}
			}
		}
		
		// Add summary of resolved alerts
		if resolvedCount > 0 {
			message.WriteString("\n✅ **Resolved Alerts:**\n")
			for i, alert := range alertData.Alerts {
				if status, ok := alert["status"].(string); ok && status == "resolved" && i < 5 { // Limit to first 5
					if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
						if summary := getStringValue(annotations["summary"]); summary != "" {
							message.WriteString(fmt.Sprintf("• %s\n", summary))
						}
					}
				}
			}
		}
		
		// Add external URL if available
		if alertData.ExternalURL != "" {
			message.WriteString(fmt.Sprintf("\n🔗 [View Details](%s)", alertData.ExternalURL))
		}
	}
	
	return message.String()
}

// getStringValue safely extracts string value from interface{}
func getStringValue(value interface{}) string {
	if value == nil {
		return ""
	}
	if str, ok := value.(string); ok {
		return str
	}
	return fmt.Sprintf("%v", value)
}
