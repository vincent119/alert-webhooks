package service

import (
	"context"
	"fmt"
	"strconv"
	"sync"

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

	b, err := bot.New(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot: %v", err)
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
	chatID, exists := ts.chatIDs[level]
	ts.mu.RUnlock()

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
