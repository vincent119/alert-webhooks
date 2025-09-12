// Package service 提供 Telegram 服務的實現
package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TelegramService Telegram 服務
type TelegramService struct {
	bot  *bot.Bot
	mu   sync.RWMutex
	chatIDs map[int]int64 // level -> chat_id 映射
}

// NewTelegramService 創建新的 Telegram 服務
func NewTelegramService(token string) (*TelegramService, error) {
	if token == "" {
		return nil, fmt.Errorf("telegram token is required")
	}

	// 創建自定義的 HTTP 客戶端，設置較長的超時時間
	httpClient := &http.Client{
		Timeout: 30 * time.Second, // 增加超時時間到 30 秒
	}

	// 使用自定義 HTTP 客戶端創建 bot
	opts := []bot.Option{
		bot.WithHTTPClient(30*time.Second, httpClient),
	}

	logger.Info("Creating Telegram bot with extended timeout", "telegram_service",
		logger.String("timeout", "30s"))

	// 嘗試多次創建 bot，處理網絡連接問題
	var b *bot.Bot
	var err error
	maxRetries := 3
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		logger.Info("Attempting to create Telegram bot", "telegram_service",
			logger.Int("attempt", attempt),
			logger.Int("max_attempts", maxRetries))
			
		b, err = bot.New(token, opts...)
		if err == nil {
			logger.Info("Telegram bot created successfully", "telegram_service",
				logger.Int("attempt", attempt))
			break
		}
		
		logger.Warn("Failed to create Telegram bot", "telegram_service",
			logger.Int("attempt", attempt),
			logger.Err(err))
			
		if attempt < maxRetries {
			backoffTime := time.Duration(attempt) * 2 * time.Second
			logger.Info("Retrying in", "telegram_service",
				logger.String("backoff", backoffTime.String()))
			time.Sleep(backoffTime)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to create bot after %d attempts: %v", maxRetries, err)
	}

	// 從配置檔案讀取 chat_id 映射
	chatIDs := make(map[int]int64)
	
	// 讀取 ChatIDs1-6
	chatIDConfigs := []struct {
		level int
		value string
	}{
		{0, config.Telegram.ChatIDs1},
		{1, config.Telegram.ChatIDs2},
		{2, config.Telegram.ChatIDs3},
		{3, config.Telegram.ChatIDs4},
		{4, config.Telegram.ChatIDs5},
		{5, config.Telegram.ChatIDs6},
	}
	
	for _, cfg := range chatIDConfigs {
		if cfg.value != "" {
			chatID, err := strconv.ParseInt(cfg.value, 10, 64)
			if err != nil {
				logger.Warn("Invalid chat ID in config", "telegram",
					logger.Int("level", cfg.level),
					logger.String("chat_id", cfg.value),
					logger.Err(err))
				continue
			}
			
			chatIDs[cfg.level] = chatID
			logger.Info("Loaded chat ID from config", "telegram",
				logger.Int("level", cfg.level),
				logger.Int64("chat_id", chatID))
		}
	}

	// 如果沒有從配置檔案讀取到任何 chat_id，使用預設值
	if len(chatIDs) == 0 {
		logger.Warn("No valid chat IDs found in config, using default values", "telegram")
		chatIDs = map[int]int64{
			0: -1001234567890, // 替換為實際的 chat_id
			1: -1001234567891,
			2: -1001234567892,
			3: -1001234567893,
			4: -1001234567894,
		}
	}

	ts := &TelegramService{
		bot:     b,
		chatIDs: chatIDs,
	}

	// 測試機器人連接
	if err := ts.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to test bot connection: %v", err)
	}

	logger.Info("Telegram service initialized successfully", "telegram",
		logger.String("bot_id", fmt.Sprintf("%d", b.ID())))

	return ts, nil
}

// testConnection 測試機器人連接
func (ts *TelegramService) testConnection() error {
	ctx := context.Background()
	me, err := ts.bot.GetMe(ctx)
	if err != nil {
		return err
	}

	logger.Info("Bot connection test successful", "telegram",
		logger.String("username", me.Username),
		logger.String("first_name", me.FirstName))

	return nil
}



// SendMessage send message to specified level chat
func (ts *TelegramService) SendMessage(level int, message string) error {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	// 檢查是否為降級模式
	if ts.bot == nil {
		logger.Error("Telegram service is in degraded mode - bot not available", "telegram_service",
			logger.Int("level", level),
			logger.String("message_preview", getMessagePreview(message)))
		return fmt.Errorf("telegram service is in degraded mode - bot initialization failed")
	}

	chatID, exists := ts.chatIDs[level]
	if !exists {
		return fmt.Errorf("no chat ID configured for level %d", level)
	}

	ctx := context.Background()
	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: models.ParseModeHTML, // 使用 HTML 格式支持連結
	}

	_, err := ts.bot.SendMessage(ctx, params)
	if err != nil {
		logger.Error("Failed to send Telegram message", "telegram",
			logger.Int("level", level),
			logger.Int64("chat_id", chatID),
			logger.String("message", message),
			logger.Err(err))
		return err
	}

	logger.Info("Telegram message sent successfully", "telegram",
		logger.Int("level", level),
		logger.Int64("chat_id", chatID),
		logger.String("message", message))

	return nil
}

// getMessagePreview 取得消息的預覽內容，限制長度
func getMessagePreview(message string) string {
	const maxPreviewLength = 100
	if len(message) <= maxPreviewLength {
		return message
	}
	return message[:maxPreviewLength] + "..."
}

// SetChatID 設定指定等級的 chat_id
func (ts *TelegramService) SetChatID(level int, chatID int64) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.chatIDs[level] = chatID

	logger.Info("Chat ID updated", "telegram",
		logger.Int("level", level),
		logger.Int64("chat_id", chatID))
}

// GetChatID 獲取指定等級的 chat_id
func (ts *TelegramService) GetChatID(level int) (int64, bool) {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	chatID, exists := ts.chatIDs[level]
	return chatID, exists
}

// GetBot 獲取 bot 實例
func (ts *TelegramService) GetBot() *bot.Bot {
	return ts.bot
}

// GetBotInfo 獲取機器人資訊
func (ts *TelegramService) GetBotInfo() (interface{}, error) {
	ctx := context.Background()
	return ts.bot.GetMe(ctx)
}
