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
	"github.com/vincent119/zlogger"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// TelegramService Telegram 服務
type TelegramService struct {
	bot     *bot.Bot
	client  telegramBotClient
	mu      sync.RWMutex
	chatIDs map[int]int64 // level -> chat_id 映射
}

type telegramBotClient interface {
	ID() int64
	GetMe(context.Context) (*models.User, error)
	SendMessage(context.Context, *bot.SendMessageParams) (*models.Message, error)
}

type telegramServiceDependencies struct {
	currentConfig func() config.Config
	newBot        func(string, ...bot.Option) (telegramBotClient, error)
	sleep         func(time.Duration)
}

// NewTelegramService 創建新的 Telegram 服務
func NewTelegramService(token string) (*TelegramService, error) {
	return newTelegramService(token, telegramServiceDependencies{
		currentConfig: config.Current,
		newBot: func(token string, options ...bot.Option) (telegramBotClient, error) {
			return bot.New(token, options...)
		},
		sleep: time.Sleep,
	})
}

func newTelegramService(token string, dependencies telegramServiceDependencies) (*TelegramService, error) {
	cfg := dependencies.currentConfig()
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

	zlogger.Info("Creating Telegram bot with extended timeout", zlogger.String("category", "telegram_service"), zlogger.String("timeout", "30s"))

	// 嘗試多次創建 bot，處理網絡連接問題
	var client telegramBotClient
	var err error
	maxRetries := 3

	for attempt := 1; attempt <= maxRetries; attempt++ {
		zlogger.Info("Attempting to create Telegram bot", zlogger.String("category", "telegram_service"), zlogger.Int("attempt", attempt),
			zlogger.Int("max_attempts", maxRetries))

		client, err = dependencies.newBot(token, opts...)
		if err == nil {
			zlogger.Info("Telegram bot created successfully", zlogger.String("category", "telegram_service"), zlogger.Int("attempt", attempt))
			break
		}

		zlogger.Warn("Failed to create Telegram bot", zlogger.String("category", "telegram_service"), zlogger.Int("attempt", attempt),
			zlogger.Err(err))

		if attempt < maxRetries {
			backoffTime := time.Duration(attempt) * 2 * time.Second
			zlogger.Info("Retrying in", zlogger.String("category", "telegram_service"), zlogger.String("backoff", backoffTime.String()))
			dependencies.sleep(backoffTime)
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
		{0, cfg.Telegram.ChatIDs0},
		{1, cfg.Telegram.ChatIDs1},
		{2, cfg.Telegram.ChatIDs2},
		{3, cfg.Telegram.ChatIDs3},
		{4, cfg.Telegram.ChatIDs4},
		{5, cfg.Telegram.ChatIDs5},
	}

	for _, cfg := range chatIDConfigs {
		if cfg.value != "" {
			chatID, err := strconv.ParseInt(cfg.value, 10, 64)
			if err != nil {
				zlogger.Warn("Invalid chat ID in config", zlogger.String("category", "telegram"), zlogger.Int("level", cfg.level),
					zlogger.String("chat_id", cfg.value),
					zlogger.Err(err))
				continue
			}

			chatIDs[cfg.level] = chatID
			zlogger.Info("Loaded chat ID from config", zlogger.String("category", "telegram"), zlogger.Int("level", cfg.level),
				zlogger.Int64("chat_id", chatID))
		}
	}

	// 如果沒有從配置檔案讀取到任何 chat_id，使用預設值
	if len(chatIDs) == 0 {
		zlogger.Warn("No valid chat IDs found in config, using default values", zlogger.String("category", "telegram"))
		chatIDs = map[int]int64{
			0: -1001234567890, // 替換為實際的 chat_id
			1: -1001234567891,
			2: -1001234567892,
			3: -1001234567893,
			4: -1001234567894,
		}
	}

	ts := &TelegramService{
		client:  client,
		chatIDs: chatIDs,
	}
	if concreteBot, ok := client.(*bot.Bot); ok {
		ts.bot = concreteBot
	}

	// 測試機器人連接
	if err := ts.testConnection(); err != nil {
		return nil, fmt.Errorf("failed to test bot connection: %v", err)
	}

	zlogger.Info("Telegram service initialized successfully", zlogger.String("category", "telegram"), zlogger.String("bot_id", fmt.Sprintf("%d", client.ID())))

	return ts, nil
}

// testConnection 測試機器人連接
func (ts *TelegramService) testConnection() error {
	ctx := context.Background()
	me, err := ts.client.GetMe(ctx)
	if err != nil {
		return err
	}

	zlogger.Info("Bot connection test successful", zlogger.String("category", "telegram"), zlogger.String("username", me.Username),
		zlogger.String("first_name", me.FirstName))

	return nil
}

// SendMessage send message to specified level chat
func (ts *TelegramService) SendMessage(ctx context.Context, level int, message string) error {
	cfg := config.Current()
	ctx, span := otel.Tracer("telegram").Start(ctx, "TelegramService.SendMessage")
	defer span.End()

	span.SetAttributes(
		attribute.String("messaging.system", "telegram"),
		attribute.Int("messaging.level", level),
	)

	ts.mu.RLock()
	defer ts.mu.RUnlock()

	// 檢查是否為降級模式
	if ts.client == nil {
		err := fmt.Errorf("telegram service is in degraded mode - bot initialization failed")
		zlogger.Error("Telegram service is in degraded mode - bot not available", zlogger.String("category", "telegram_service"), zlogger.Int("level", level),
			zlogger.String("message_preview", getMessagePreview(message)))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	chatID, exists := ts.chatIDs[level]
	if !exists {
		// 記錄所有已配置的 level
		configuredLevels := make([]int, 0, len(ts.chatIDs))
		for l := range ts.chatIDs {
			configuredLevels = append(configuredLevels, l)
		}

		zlogger.Error("No chat ID configured for requested level", zlogger.String("category", "telegram_service"), zlogger.Int("requested_level", level),
			zlogger.Any("configured_levels", configuredLevels),
			zlogger.Int("total_configured", len(ts.chatIDs)))
		err := fmt.Errorf("no chat ID configured for level %d", level)
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	zlogger.Debug("Found chat ID for level", zlogger.String("category", "telegram_service"), zlogger.Int("level", level),
		zlogger.Int64("chat_id", chatID))

	span.SetAttributes(attribute.Int64("messaging.chat_id", chatID))

	params := &bot.SendMessageParams{
		ChatID:    chatID,
		Text:      message,
		ParseMode: models.ParseModeHTML, // 使用 HTML 格式支持連結
	}

	// 在 debug 模式下記錄發送請求的詳細資訊
	if cfg.App.Mode == "development" || cfg.App.Mode == "debug" || cfg.Log.Level == "debug" {
		zlogger.Debug("Sending Telegram message request", zlogger.String("category", "telegram_service"), zlogger.Int("level", level),
			zlogger.Int64("chat_id", chatID),
			zlogger.String("parse_mode", string(models.ParseModeHTML)),
			zlogger.String("message_length", fmt.Sprintf("%d", len(message))),
			zlogger.String("message_preview", getMessagePreview(message)))
	}

	response, err := ts.client.SendMessage(ctx, params)
	if err != nil {
		zlogger.Error("Failed to send Telegram message", zlogger.String("category", "telegram_service"), zlogger.Int("level", level),
			zlogger.Int64("chat_id", chatID),
			zlogger.String("message_preview", getMessagePreview(message)),
			zlogger.Err(err))
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	// 在 debug 模式下記錄 Telegram server 的詳細回應
	if cfg.App.Mode == "development" || cfg.App.Mode == "debug" || cfg.Log.Level == "debug" {
		zlogger.Debug("Telegram server response received", zlogger.String("category", "telegram_service"), zlogger.Int("level", level),
			zlogger.Int64("chat_id", chatID),
			zlogger.Int("message_id", response.ID),
			zlogger.String("from_username", response.From.Username),
			zlogger.String("from_first_name", response.From.FirstName),
			zlogger.Bool("from_is_bot", response.From.IsBot),
			zlogger.Int64("chat_id_response", response.Chat.ID),
			zlogger.String("chat_type", string(response.Chat.Type)),
			zlogger.String("chat_title", response.Chat.Title),
			zlogger.Int("date", response.Date),
			zlogger.String("response_text_preview", getMessagePreview(response.Text)))
	}

	zlogger.Info("Telegram message sent successfully", zlogger.String("category", "telegram_service"), zlogger.Int("level", level),
		zlogger.Int64("chat_id", chatID),
		zlogger.Int("message_id", response.ID),
		zlogger.String("message_preview", getMessagePreview(message)))

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

	zlogger.Info("Chat ID updated", zlogger.String("category", "telegram"), zlogger.Int("level", level),
		zlogger.Int64("chat_id", chatID))
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
	return ts.client.GetMe(ctx)
}
