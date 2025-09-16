package providers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"alert-webhooks/pkg/notification/types"
)

// TelegramService 接口定義（避免循環依賴）
type TelegramService interface {
	SendMessage(level int, message string) error
	GetBotInfo() (interface{}, error)
}



// TelegramProvider Telegram 通知提供者
type TelegramProvider struct {
	telegramService TelegramService
	templateEngine  types.TemplateEngine
	config          *config.TelegramConf
	stats           *types.ProviderStats
}

// NewTelegramProvider 創建 Telegram 提供者
func NewTelegramProvider(telegramService TelegramService, templateEngine types.TemplateEngine) (types.NotificationProvider, error) {
	if telegramService == nil {
		return nil, fmt.Errorf("telegram service not available")
	}
	
	if templateEngine == nil {
		return nil, fmt.Errorf("template engine not available")
	}
	
	provider := &TelegramProvider{
		telegramService: telegramService,
		templateEngine:  templateEngine,
		config:          &config.Telegram,
		stats: &types.ProviderStats{
			MessagesSent:  0,
			MessagesError: 0,
		},
	}
	
	logger.Info("Telegram provider initialized", "telegram_provider")
	return provider, nil
}

// GetName 獲取提供者名稱
func (tp *TelegramProvider) GetName() string {
	return "telegram"
}

// SendMessage send message
func (tp *TelegramProvider) SendMessage(ctx context.Context, req *types.NotificationRequest) error {
	logger.Info("Sending Telegram message", "telegram_provider",
		logger.String("level", req.Level),
		logger.String("chat_id", req.ChatID))
	
	// 解析等級
	level, err := tp.parseLevelFromRequest(req)
	if err != nil {
		tp.stats.MessagesError++
		return fmt.Errorf("invalid level format: %v", err)
	}
	
	// 在 debug 模式下記錄發送前的詳細資訊
	if config.IsDevelopment() || strings.ToLower(config.App.Mode) == "debug" || strings.ToLower(config.Log.Level) == "debug" {
		// 計算訊息預覽長度
		previewLength := len(req.Message)
		if previewLength > 100 {
			previewLength = 100
		}
		messagePreview := req.Message[:previewLength]
		if len(req.Message) > 100 {
			messagePreview += "..."
		}
		
		logger.Debug("Telegram provider sending message", "telegram_provider",
			logger.String("level", req.Level),
			logger.String("chat_id", req.ChatID),
			logger.Int("parsed_level", level),
			logger.String("message_length", fmt.Sprintf("%d", len(req.Message))),
			logger.String("message_preview", messagePreview))
	}
	
	// Send message
	err = tp.telegramService.SendMessage(level, req.Message)
	if err != nil {
		tp.stats.MessagesError++
		logger.Error("Failed to send Telegram message", "telegram_provider",
			logger.Int("level", level),
			logger.Err(err))
		return err
	}
	
	// 更新統計
	tp.stats.MessagesSent++
	tp.stats.LastMessageTime = time.Now().Unix()
	
	logger.Info("Telegram message sent successfully", "telegram_provider",
		logger.Int("level", level),
		logger.Int64("messages_sent_total", tp.stats.MessagesSent))
	
	return nil
}

// parseLevelFromRequest 從請求中解析等級
func (tp *TelegramProvider) parseLevelFromRequest(req *types.NotificationRequest) (int, error) {
	// 優先使用 Level 字段
	if req.Level != "" {
		// 支援 "L0", "L1", "0", "1" 等格式
		levelStr := strings.TrimPrefix(req.Level, "L")
		level, err := strconv.Atoi(levelStr)
		if err != nil {
			return 0, fmt.Errorf("invalid level format: %s", req.Level)
		}
		
		// 驗證等級範圍
		if level < 0 || level > 6 {
			return 0, fmt.Errorf("level out of range (0-6): %d", level)
		}
		
		return level, nil
	}
	
	// 如果沒有 Level，嘗試從 ChatID 解析
	if req.ChatID != "" {
		// 假設 ChatID 格式為 "L0", "L1" 等
		if strings.HasPrefix(req.ChatID, "L") {
			levelStr := strings.TrimPrefix(req.ChatID, "L")
			level, err := strconv.Atoi(levelStr)
			if err == nil && level >= 0 && level <= 6 {
				return level, nil
			}
		}
	}
	
	return 0, fmt.Errorf("no valid level found in request")
}

// ValidateConfig 驗證配置
func (tp *TelegramProvider) ValidateConfig() error {
	if tp.config.Token == "" {
		return fmt.Errorf("telegram token is required")
	}
	
	// 檢查至少有一個聊天ID配置
	hasValidChatID := false
	chatIDs := []string{
		tp.config.ChatIDs0, tp.config.ChatIDs1, tp.config.ChatIDs2,
		tp.config.ChatIDs3, tp.config.ChatIDs4, tp.config.ChatIDs5, tp.config.ChatIDs6,
	}
	
	for _, chatID := range chatIDs {
		if chatID != "" {
			hasValidChatID = true
			break
		}
	}
	
	if !hasValidChatID {
		return fmt.Errorf("at least one telegram chat ID must be configured")
	}
	
	return nil
}

// IsEnabled 檢查是否啟用
func (tp *TelegramProvider) IsEnabled() bool {
	return tp.config.Enable
}

// GetCapabilities 獲取能力描述
func (tp *TelegramProvider) GetCapabilities() *types.ProviderCapabilities {
	// 動態獲取支援的語言
	supportedLanguages := []string{"eng", "tw", "zh", "ja", "ko"} // 預設值
	if tp.templateEngine != nil {
		supportedLanguages = tp.templateEngine.GetSupportedLanguages()
	}
	
	return &types.ProviderCapabilities{
		SupportsLevels:     true,
		SupportsChannels:   false, // Telegram 使用 ChatID，不是頻道概念
		SupportsRichText:   true,  // 支援 Markdown
		SupportsAttachments: false,
		SupportedLanguages: supportedLanguages,
		MaxMessageLength:   4096, // Telegram 限制
	}
}

// GetStatus 獲取服務狀態
func (tp *TelegramProvider) GetStatus() *types.ProviderStatus {
	// 測試連接
	connected := false
	lastError := ""
	
	if tp.telegramService != nil {
		_, err := tp.telegramService.GetBotInfo()
		if err != nil {
			lastError = err.Error()
		} else {
			connected = true
		}
	}
	
	// 構建頻道映射（這裡是聊天ID）
	channels := make(map[string]string)
	if tp.config.ChatIDs0 != "" {
		channels["L0"] = tp.config.ChatIDs0
	}
	if tp.config.ChatIDs1 != "" {
		channels["L1"] = tp.config.ChatIDs1
	}
	if tp.config.ChatIDs2 != "" {
		channels["L2"] = tp.config.ChatIDs2
	}
	if tp.config.ChatIDs3 != "" {
		channels["L3"] = tp.config.ChatIDs3
	}
	if tp.config.ChatIDs4 != "" {
		channels["L4"] = tp.config.ChatIDs4
	}
	if tp.config.ChatIDs5 != "" {
		channels["L5"] = tp.config.ChatIDs5
	}
	if tp.config.ChatIDs6 != "" {
		channels["L6"] = tp.config.ChatIDs6
	}
	
	return &types.ProviderStatus{
		Name:       "telegram",
		Enabled:    tp.config.Enable,
		Connected:  connected,
		LastError:  lastError,
		Channels:   channels,
		Statistics: tp.stats,
	}
}

// TestConnection 測試連接
func (tp *TelegramProvider) TestConnection() error {
	if tp.telegramService == nil {
		return fmt.Errorf("telegram service not available")
	}
	
	_, err := tp.telegramService.GetBotInfo()
	if err != nil {
		return fmt.Errorf("failed to connect to Telegram: %v", err)
	}
	
	return nil
}
