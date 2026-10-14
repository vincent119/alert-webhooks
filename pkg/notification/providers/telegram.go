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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// TelegramService 接口定義（避免循環依賴）
type TelegramService interface {
	SendMessage(ctx context.Context, level int, message string) error
	GetBotInfo() (interface{}, error)
}



// TelegramProvider Telegram 通知提供者
type TelegramProvider struct {
	telegramService TelegramService
	templateEngine  types.TemplateEngine
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
	cfg := config.Current()
	ctx, span := otel.Tracer("telegram").Start(ctx, "TelegramProvider.SendMessage")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "telegram"),
		attribute.String("messaging.level", req.Level),
		attribute.String("messaging.chat_id", req.ChatID),
	)

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
	if cfg.App.Mode == "development" || strings.ToLower(cfg.App.Mode) == "debug" || strings.ToLower(cfg.Log.Level) == "debug" {
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
	err = tp.telegramService.SendMessage(ctx, level, req.Message)
	if err != nil {
		tp.stats.MessagesError++
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
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
	cfg := config.Current()
	if cfg.Telegram.Token == "" {
		return fmt.Errorf("telegram token is required")
	}

	// 檢查至少有一個聊天ID配置
	hasValidChatID := false
	chatIDs := []string{
		cfg.Telegram.ChatIDs0, cfg.Telegram.ChatIDs1, cfg.Telegram.ChatIDs2,
		cfg.Telegram.ChatIDs3, cfg.Telegram.ChatIDs4, cfg.Telegram.ChatIDs5, cfg.Telegram.ChatIDs6,
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
	return config.Current().Telegram.Enable
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
	cfg := config.Current()
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
	if cfg.Telegram.ChatIDs0 != "" {
		channels["L0"] = cfg.Telegram.ChatIDs0
	}
	if cfg.Telegram.ChatIDs1 != "" {
		channels["L1"] = cfg.Telegram.ChatIDs1
	}
	if cfg.Telegram.ChatIDs2 != "" {
		channels["L2"] = cfg.Telegram.ChatIDs2
	}
	if cfg.Telegram.ChatIDs3 != "" {
		channels["L3"] = cfg.Telegram.ChatIDs3
	}
	if cfg.Telegram.ChatIDs4 != "" {
		channels["L4"] = cfg.Telegram.ChatIDs4
	}
	if cfg.Telegram.ChatIDs5 != "" {
		channels["L5"] = cfg.Telegram.ChatIDs5
	}
	if cfg.Telegram.ChatIDs6 != "" {
		channels["L6"] = cfg.Telegram.ChatIDs6
	}

	return &types.ProviderStatus{
		Name:       "telegram",
		Enabled:    cfg.Telegram.Enable,
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
