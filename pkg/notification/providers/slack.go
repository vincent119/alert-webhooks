package providers

import (
	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/notification/types"
	"context"
	"fmt"
	"strings"
	"time"
)

// SlackService 接口定義（避免循環依賴）
type SlackService interface {
	SendMessage(channel, message string) error
	SendMessageToLevel(level, message string) error
	TestConnection() error
}



// SlackProvider Slack 通知提供者
type SlackProvider struct {
	slackService   SlackService
	templateEngine types.TemplateEngine
	config         *config.SlackConf
	stats          *types.ProviderStats
}

// NewSlackProvider 創建 Slack 提供者
func NewSlackProvider(slackService SlackService, templateEngine types.TemplateEngine) (types.NotificationProvider, error) {
	if slackService == nil {
		return nil, fmt.Errorf("slack service not available")
	}
	
	if templateEngine == nil {
		return nil, fmt.Errorf("template engine not available")
	}
	
	provider := &SlackProvider{
		slackService:   slackService,
		templateEngine: templateEngine,
		config:         &config.Slack,
		stats: &types.ProviderStats{
			MessagesSent:  0,
			MessagesError: 0,
		},
	}
	
	logger.Info("Slack provider initialized", "slack_provider")
	return provider, nil
}

// GetName 獲取提供者名稱
func (sp *SlackProvider) GetName() string {
	return "slack"
}

// SendMessage send message
func (sp *SlackProvider) SendMessage(ctx context.Context, req *types.NotificationRequest) error {
	logger.Info("Sending Slack message", "slack_provider",
		logger.String("level", req.Level),
		logger.String("channel", req.Channel))
	
	var channel string
	var err error
	
	// 決定發送到哪個頻道
	if req.Channel != "" {
		// 直接指定頻道
		channel = req.Channel
		if !strings.HasPrefix(channel, "#") && !strings.HasPrefix(channel, "@") {
			channel = "#" + channel
		}
		err = sp.slackService.SendMessage(channel, req.Message)
	} else if req.Level != "" {
		// 根據等級發送
		err = sp.slackService.SendMessageToLevel(req.Level, req.Message)
		channel = sp.getLevelChannel(req.Level)
	} else {
		// 使用預設頻道
		channel = sp.config.Channel
		if channel == "" {
			channel = "#alerts" // 預設頻道
		}
		err = sp.slackService.SendMessage(channel, req.Message)
	}
	
	if err != nil {
		sp.stats.MessagesError++
		logger.Error("Failed to send Slack message", "slack_provider",
			logger.String("channel", channel),
			logger.Err(err))
		return err
	}
	
	// 更新統計
	sp.stats.MessagesSent++
	sp.stats.LastMessageTime = time.Now().Unix()
	
	logger.Info("Slack message sent successfully", "slack_provider",
		logger.String("channel", channel))
	
	return nil
}

// getLevelChannel 根據等級獲取頻道
func (sp *SlackProvider) getLevelChannel(level string) string {
	// 標準化等級格式
	normalizedLevel := strings.ToLower(level)
	if !strings.HasPrefix(normalizedLevel, "l") {
		normalizedLevel = "l" + normalizedLevel
	}
	
	// 從配置中查找
	if sp.config.Channels != nil {
		if channel, exists := sp.config.Channels[normalizedLevel]; exists {
			return channel
		}
		// 嘗試不同的格式
		levelFormats := []string{
			level,
			strings.ToUpper(level),
			strings.TrimPrefix(normalizedLevel, "l"),
		}
		for _, format := range levelFormats {
			if channel, exists := sp.config.Channels[format]; exists {
				return channel
			}
		}
	}
	
	// 使用預設頻道
	if sp.config.Channel != "" {
		return sp.config.Channel
	}
	
	return "#alerts"
}

// ValidateConfig 驗證配置
func (sp *SlackProvider) ValidateConfig() error {
	if sp.config.Token == "" {
		return fmt.Errorf("slack token is required")
	}
	
	if sp.config.Channel == "" && len(sp.config.Channels) == 0 {
		return fmt.Errorf("at least one slack channel must be configured")
	}
	
	return nil
}

// IsEnabled 檢查是否啟用
func (sp *SlackProvider) IsEnabled() bool {
	return sp.config.Enable
}

// GetCapabilities 獲取能力描述
func (sp *SlackProvider) GetCapabilities() *types.ProviderCapabilities {
	// 動態獲取支援的語言
	supportedLanguages := []string{"eng", "tw", "zh", "ja", "ko"} // 預設值
	if sp.templateEngine != nil {
		supportedLanguages = sp.templateEngine.GetSupportedLanguages()
	}
	
	return &types.ProviderCapabilities{
		SupportsLevels:     true,
		SupportsChannels:   true,
		SupportsRichText:   true,  // 支援 Markdown 和 Blocks
		SupportsAttachments: true, // 支援附件
		SupportedLanguages: supportedLanguages,
		MaxMessageLength:   40000, // Slack 限制較大
	}
}

// GetStatus 獲取服務狀態
func (sp *SlackProvider) GetStatus() *types.ProviderStatus {
	// 測試連接
	connected := false
	lastError := ""
	
	if sp.slackService != nil {
		err := sp.slackService.TestConnection()
		if err != nil {
			lastError = err.Error()
		} else {
			connected = true
		}
	}
	
	// 構建頻道映射
	channels := make(map[string]string)
	
	// 添加預設頻道
	if sp.config.Channel != "" {
		channels["default"] = sp.config.Channel
	}
	
	// 添加等級頻道
	if sp.config.Channels != nil {
		for level, channel := range sp.config.Channels {
			channels[level] = channel
		}
	}
	
	return &types.ProviderStatus{
		Name:       "slack",
		Enabled:    sp.config.Enable,
		Connected:  connected,
		LastError:  lastError,
		Channels:   channels,
		Statistics: sp.stats,
	}
}

// TestConnection 測試連接
func (sp *SlackProvider) TestConnection() error {
	if sp.slackService == nil {
		return fmt.Errorf("slack service not available")
	}
	
	return sp.slackService.TestConnection()
}
