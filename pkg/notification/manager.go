package notification

import (
	"context"
	"fmt"
	"sync"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/notification/providers"
	"alert-webhooks/pkg/notification/types"
	"alert-webhooks/pkg/template"
)

// NotificationManager 統一通知管理器
type NotificationManager struct {
	providers      map[string]types.NotificationProvider
	templateEngine *template.TemplateEngine
	mu             sync.RWMutex
}

var (
	managerInstance *NotificationManager
	managerOnce     sync.Once
)

// GetNotificationManager 獲取通知管理器單例
func GetNotificationManager() *NotificationManager {
	managerOnce.Do(func() {
		managerInstance = &NotificationManager{
			providers: make(map[string]types.NotificationProvider),
		}
	})
	return managerInstance
}

// Initialize 初始化通知管理器
func (nm *NotificationManager) Initialize(templateEngine *template.TemplateEngine, telegramService providers.TelegramService, slackService providers.SlackService, discordService providers.DiscordService) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	logger.Info("Initializing notification manager", "notification_manager")
	
	// 設置模板引擎
	nm.templateEngine = templateEngine
	
	// 清空現有提供者
	nm.providers = make(map[string]types.NotificationProvider)
	
	// 註冊 Telegram 提供者
	if config.Telegram.Enable && telegramService != nil {
		telegramProvider, err := providers.NewTelegramProvider(telegramService, templateEngine)
		if err != nil {
			logger.Error("Failed to initialize Telegram provider", "notification_manager", logger.Err(err))
		} else {
			nm.providers["telegram"] = telegramProvider
			logger.Info("Telegram provider registered", "notification_manager")
		}
	}
	
	// 註冊 Slack 提供者
	if config.Slack.Enable && slackService != nil {
		slackProvider, err := providers.NewSlackProvider(slackService, templateEngine)
		if err != nil {
			logger.Error("Failed to initialize Slack provider", "notification_manager", logger.Err(err))
		} else {
			nm.providers["slack"] = slackProvider
			logger.Info("Slack provider registered", "notification_manager")
		}
	}
	
	// 註冊 Discord 提供者
	if config.Conf.Discord.Enable && discordService != nil {
		discordProvider := providers.NewDiscordProvider(discordService, templateEngine)
		nm.providers["discord"] = discordProvider
		logger.Info("Discord provider registered", "notification_manager")
	}
	
	logger.Info("Notification manager initialized", "notification_manager",
		logger.Int("providers_count", len(nm.providers)))
	
	return nil
}

// SendNotification 發送通知
func (nm *NotificationManager) SendNotification(ctx context.Context, providerName string, req *types.NotificationRequest) (*types.NotificationResponse, error) {
	nm.mu.RLock()
	provider, exists := nm.providers[providerName]
	nm.mu.RUnlock()
	
	if !exists {
		return &types.NotificationResponse{
			Success:  false,
			Message:  fmt.Sprintf("Provider '%s' not found", providerName),
			Provider: providerName,
		}, fmt.Errorf("provider '%s' not found", providerName)
	}
	
	if !provider.IsEnabled() {
		return &types.NotificationResponse{
			Success:  false,
			Message:  fmt.Sprintf("Provider '%s' is disabled", providerName),
			Provider: providerName,
		}, fmt.Errorf("provider '%s' is disabled", providerName)
	}
	
	// 預處理請求（渲染模板等）
	if err := nm.preprocessRequest(req, providerName); err != nil {
		return &types.NotificationResponse{
			Success:  false,
			Message:  fmt.Sprintf("Failed to preprocess request: %v", err),
			Provider: providerName,
		}, err
	}
	
	// 發送通知
		if err := provider.SendMessage(ctx, req); err != nil {
		return &types.NotificationResponse{
			Success:  false,
			Message:  fmt.Sprintf("Failed to send message: %v", err),
			Provider: providerName,
			Level:    req.Level,
			ChatID:   req.ChatID,
			Channel:  req.Channel,
		}, err
	}

	return &types.NotificationResponse{
		Success:  true,
		Message:  "Message sent successfully",
		Provider: providerName,
		Level:    req.Level,
		ChatID:   req.ChatID,
		Channel:  req.Channel,
	}, nil
}

// preprocessRequest 預處理請求
func (nm *NotificationManager) preprocessRequest(req *types.NotificationRequest, providerName string) error {
	// If AlertManager data is provided but no message content, render template
	if req.AlertData != nil && req.Message == "" {
		if nm.templateEngine != nil {
			// 獲取提供者特定的模板語言
			templateLanguage := nm.getProviderTemplateLanguage(providerName, req.TemplateLanguage)
			
			// 轉換 AlertManager 數據為模板格式
			templateData, err := nm.convertAlertManagerData(req.AlertData)
			if err != nil {
				return fmt.Errorf("failed to convert AlertManager data: %v", err)
			}
			
			// 渲染模板
			actualLanguage := nm.templateEngine.GetDefaultLanguage(templateLanguage)
			message, err := nm.templateEngine.RenderTemplate(actualLanguage, *templateData)
			if err != nil {
				logger.Warn("Failed to render template, will use raw data", "notification_manager",
					logger.String("provider", providerName),
					logger.String("language", actualLanguage),
					logger.Err(err))
				return err
			}
			
			req.Message = message
			logger.Info("Template rendered successfully", "notification_manager",
				logger.String("provider", providerName),
				logger.String("language", actualLanguage))
		}
	}
	
	return nil
}

// getProviderTemplateLanguage 獲取提供者特定的模板語言
func (nm *NotificationManager) getProviderTemplateLanguage(providerName, requestLanguage string) string {
	// 優先使用請求中指定的語言
	if requestLanguage != "" {
		return requestLanguage
	}
	
	// 根據提供者獲取配置的語言
	switch providerName {
	case "telegram":
		return config.Telegram.TemplateLanguage
	case "slack":
		return config.Slack.TemplateLanguage
	default:
		return "eng" // 預設英文
	}
}

// convertAlertManagerData 轉換 AlertManager 數據為模板格式
func (nm *NotificationManager) convertAlertManagerData(data *types.AlertManagerData) (*template.TemplateData, error) {
	if data == nil {
		return nil, fmt.Errorf("AlertManager data is nil")
	}

	// 統計警報
	firingCount := 0
	resolvedCount := 0
	var alertsForTemplate []template.AlertData

	for _, alert := range data.Alerts {
		alertData := template.AlertData{}

		// 狀態
		if status, ok := alert["status"].(string); ok {
			alertData.Status = status
			if status == "firing" {
				firingCount++
			} else if status == "resolved" {
				resolvedCount++
			}
		}

		// 標籤
		if labels, ok := alert["labels"].(map[string]interface{}); ok {
			alertData.Labels = make(map[string]string)
			for k, v := range labels {
				if str, ok := v.(string); ok {
					alertData.Labels[k] = str
				}
			}
		}

		// 註解
		if annotations, ok := alert["annotations"].(map[string]interface{}); ok {
			alertData.Annotations = make(map[string]string)
			for k, v := range annotations {
				if str, ok := v.(string); ok {
					alertData.Annotations[k] = str
				}
			}
		}

		// 時間戳
		if startsAt, ok := alert["startsAt"].(string); ok {
			alertData.StartsAt = startsAt
		}
		if endsAt, ok := alert["endsAt"].(string); ok {
			alertData.EndsAt = endsAt
		}
		if generatorURL, ok := alert["generatorURL"].(string); ok {
			alertData.GeneratorURL = generatorURL
		}

		alertsForTemplate = append(alertsForTemplate, alertData)
	}

	// Get basic message
	var alertName, env, severity, namespace string
	if data.CommonLabels != nil {
		if name, ok := data.CommonLabels["alertname"].(string); ok {
			alertName = name
		}
		if envValue, ok := data.CommonLabels["env"].(string); ok {
			env = envValue
		}
		if severityValue, ok := data.CommonLabels["severity"].(string); ok {
			severity = severityValue
		}
		if namespaceValue, ok := data.CommonLabels["namespace"].(string); ok {
			namespace = namespaceValue
		}
	}

	return &template.TemplateData{
		Status:        data.Status,
		AlertName:     alertName,
		Env:           env,
		Severity:      severity,
		Namespace:     namespace,
		TotalAlerts:   len(data.Alerts),
		FiringCount:   firingCount,
		ResolvedCount: resolvedCount,
		Alerts:        alertsForTemplate,
		ExternalURL:   data.ExternalURL,
	}, nil
}

// GetProvider 獲取指定提供者
func (nm *NotificationManager) GetProvider(name string) (types.NotificationProvider, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	provider, exists := nm.providers[name]
	return provider, exists
}

// GetAllProviders 獲取所有提供者
func (nm *NotificationManager) GetAllProviders() map[string]types.NotificationProvider {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	result := make(map[string]types.NotificationProvider)
	for name, provider := range nm.providers {
		result[name] = provider
	}
	return result
}

// GetProviderStatus 獲取提供者狀態
func (nm *NotificationManager) GetProviderStatus(name string) (*types.ProviderStatus, error) {
	nm.mu.RLock()
	provider, exists := nm.providers[name]
	nm.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("provider '%s' not found", name)
	}
	
	return provider.GetStatus(), nil
}

// Reload 重新載入通知管理器
func (nm *NotificationManager) Reload(templateEngine *template.TemplateEngine, telegramService providers.TelegramService, slackService providers.SlackService, discordService providers.DiscordService) error {
	logger.Info("Reloading notification manager", "notification_manager")
	return nm.Initialize(templateEngine, telegramService, slackService, discordService)
}
