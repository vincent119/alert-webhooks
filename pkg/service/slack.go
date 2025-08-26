package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
)

// SlackService Slack 服務
type SlackService struct {
	token    string
	client   *http.Client
	mu       sync.RWMutex
	channels map[string]string // level -> channel 映射
}

// SlackMessage Slack message structure
type SlackMessage struct {
	Channel     string       `json:"channel"`
	Text        string       `json:"text,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconURL     string       `json:"icon_url,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	ThreadTS    string       `json:"thread_ts,omitempty"`
	LinkNames   bool         `json:"link_names,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links,omitempty"`
	UnfurlMedia bool         `json:"unfurl_media,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	Blocks      []Block      `json:"blocks,omitempty"`
}

// Attachment Slack 附件結構
type Attachment struct {
	Color      string  `json:"color,omitempty"`
	Title      string  `json:"title,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Timestamp  int64   `json:"ts,omitempty"`
}

// Field Slack 字段結構
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Block Slack 塊結構（簡化版）
type Block struct {
	Type string      `json:"type"`
	Text interface{} `json:"text,omitempty"`
}

// SlackResponse Slack API 響應
type SlackResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// NewSlackService 創建新的 Slack 服務
func NewSlackService(token string) (*SlackService, error) {
	if token == "" {
		return nil, fmt.Errorf("slack token is required")
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// 從配置檔案讀取頻道映射
	channels := make(map[string]string)
	
	// 如果有配置多頻道，使用多頻道配置
	if len(config.Slack.Channels) > 0 {
		channels = config.Slack.Channels
	} else if config.Slack.Channel != "" {
		// 否則使用預設頻道
		channels["default"] = config.Slack.Channel
	}

	ss := &SlackService{
		token:    token,
		client:   client,
		channels: channels,
	}

	// 測試連接
	if err := ss.TestConnection(); err != nil {
		return nil, fmt.Errorf("failed to test slack connection: %v", err)
	}

	logger.Info("Slack service initialized successfully", "slack",
		logger.String("channels_count", fmt.Sprintf("%d", len(channels))))

	return ss, nil
}

// testConnection 測試 Slack 連接
func (ss *SlackService) TestConnection() error {
	url := "https://slack.com/api/auth.test"
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	
	req.Header.Set("Authorization", "Bearer "+ss.token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := ss.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	
	var slackResp SlackResponse
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return err
	}
	
	if !slackResp.OK {
		return fmt.Errorf("slack auth test failed: %s", slackResp.Error)
	}
	
	logger.Info("Slack connection test successful", "slack")
	return nil
}

// SendMessage send message to specified channel
func (ss *SlackService) SendMessage(channel, message string) error {
	return ss.SendMessageWithOptions(channel, message, nil)
}

// SendMessageWithOptions send message to specified channel with options
func (ss *SlackService) SendMessageWithOptions(channel, message string, options *SlackMessage) error {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	// Build message
	msg := &SlackMessage{
		Channel: channel,
		Text:    message,
	}

	// 應用配置中的預設值
	if config.Slack.Username != "" {
		msg.Username = config.Slack.Username
	}
	if config.Slack.IconURL != "" {
		msg.IconURL = config.Slack.IconURL
	}
	if config.Slack.IconEmoji != "" {
		msg.IconEmoji = config.Slack.IconEmoji
	}
	msg.LinkNames = config.Slack.LinkNames
	msg.UnfurlLinks = config.Slack.UnfurlLinks
	msg.UnfurlMedia = config.Slack.UnfurlMedia

	// 如果提供了選項，覆蓋預設值
	if options != nil {
		if options.Username != "" {
			msg.Username = options.Username
		}
		if options.IconURL != "" {
			msg.IconURL = options.IconURL
		}
		if options.IconEmoji != "" {
			msg.IconEmoji = options.IconEmoji
		}
		if options.ThreadTS != "" {
			msg.ThreadTS = options.ThreadTS
		}
		if len(options.Attachments) > 0 {
			msg.Attachments = options.Attachments
		}
		if len(options.Blocks) > 0 {
			msg.Blocks = options.Blocks
		}
	}

	return ss.sendSlackMessage(msg)
}

// SendMessageToLevel send message to specified level channel
func (ss *SlackService) SendMessageToLevel(level string, message string) error {
	// 動態讀取最新的配置以支持熱重載
	var channel string
	var exists bool
	
	if len(config.Slack.Channels) > 0 {
		// 將 "L0" 格式轉換為 "chat_ids0" 格式
		var configKey string
		if strings.HasPrefix(strings.ToUpper(level), "L") {
			// 從 "L0" 提取數字部分
			levelNumber := level[1:]
			configKey = "chat_ids" + levelNumber
		} else {
			configKey = level
		}
		
		// 嘗試使用轉換後的鍵
		channel, exists = config.Slack.Channels[configKey]
		
		// 如果沒找到，嘗試原始格式作為備用
		if !exists {
			channel, exists = config.Slack.Channels[level]
		}
		
		// 如果還是沒找到，嘗試小寫版本
		if !exists {
			lowerLevel := strings.ToLower(level)
			channel, exists = config.Slack.Channels[lowerLevel]
		}
	}

	if !exists {
		// 如果沒有找到特定等級的頻道，使用預設頻道
		if config.Slack.Channel != "" {
			channel = config.Slack.Channel
		} else {
			return fmt.Errorf("no channel configured for level %s and no default channel", level)
		}
	}

	return ss.SendMessage(channel, message)
}

// SendRichMessage send rich text message
func (ss *SlackService) SendRichMessage(channel, title, message, color string, fields []Field) error {
	attachment := Attachment{
		Color:  color,
		Title:  title,
		Text:   message,
		Fields: fields,
		Footer: "Alert Webhooks",
		Timestamp: time.Now().Unix(),
	}

	options := &SlackMessage{
		Attachments: []Attachment{attachment},
	}

	return ss.SendMessageWithOptions(channel, "", options)
}

// sendSlackMessage actually send Slack message
func (ss *SlackService) sendSlackMessage(msg *SlackMessage) error {
	url := "https://slack.com/api/chat.postMessage"
	
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ss.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := ss.client.Do(req)
	if err != nil {
		logger.Error("Failed to send Slack message", "slack",
			logger.String("channel", msg.Channel),
			logger.String("text", msg.Text),
			logger.Err(err))
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var slackResp SlackResponse
	if err := json.Unmarshal(body, &slackResp); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if !slackResp.OK {
		logger.Error("Slack API returned error", "slack",
			logger.String("channel", msg.Channel),
			logger.String("error", slackResp.Error))
		
		// Provide more friendly error message
		var friendlyError string
		switch slackResp.Error {
		case "not_in_channel":
			friendlyError = fmt.Sprintf("Bot is not in channel %s. Please invite the bot to this channel in Slack: /invite @your_bot_name", msg.Channel)
		case "channel_not_found":
			friendlyError = fmt.Sprintf("Channel %s does not exist. Please check if the channel name is correct", msg.Channel)
		case "invalid_auth":
			friendlyError = "Invalid Slack token. Please check if the token in configuration is correct"
		case "missing_scope":
			friendlyError = "Bot lacks necessary permissions. Please add 'chat:write' permission in Slack App settings"
		default:
			friendlyError = fmt.Sprintf("Slack API error: %s", slackResp.Error)
		}
		
		return fmt.Errorf("%s", friendlyError)
	}

	logger.Info("Slack message sent successfully", "slack",
		logger.String("channel", msg.Channel),
		logger.String("text", msg.Text))

	return nil
}

// SetChannelForLevel 設定指定等級的頻道
func (ss *SlackService) SetChannelForLevel(level, channel string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.channels[level] = channel

	logger.Info("Slack channel updated", "slack",
		logger.String("level", level),
		logger.String("channel", channel))
}

// GetChannelForLevel 獲取指定等級的頻道
func (ss *SlackService) GetChannelForLevel(level string) (string, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	channel, exists := ss.channels[level]
	return channel, exists
}

// GetChannels 獲取所有頻道配置
func (ss *SlackService) GetChannels() map[string]string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	
	channels := make(map[string]string)
	for k, v := range ss.channels {
		channels[k] = v
	}
	return channels
}
