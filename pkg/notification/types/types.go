package types

import (
	"context"
)

// TemplateEngine 接口定義（避免循環依賴）
type TemplateEngine interface {
	GetSupportedLanguages() []string
}

// AlertManagerData 簡化 AlertManager webhook 數據結構
type AlertManagerData struct {
	Receiver          string                   `json:"receiver"`
	Status            string                   `json:"status"`
	Alerts            []map[string]interface{} `json:"alerts"`
	GroupLabels       map[string]interface{}   `json:"groupLabels"`
	CommonLabels      map[string]interface{}   `json:"commonLabels"`
	CommonAnnotations map[string]interface{}   `json:"commonAnnotations"`
	ExternalURL       string                   `json:"externalURL"`
	Version           string                   `json:"version"`
	GroupKey          string                   `json:"groupKey"`
	TruncatedAlerts   int                      `json:"truncatedAlerts"`
}

// NotificationRequest 統一的通知請求結構
type NotificationRequest struct {
	ProviderName string // 例如 "telegram", "slack"
	Level        string // 例如 "L0", "critical"
	Channel      string // 例如 "#alerts", "-123456789"
	ChatID       string // Telegram 專用
	Message      string // Simple text message
	AlertData    *AlertManagerData // AlertManager 數據
	TemplateLanguage string // 模板語言
	
	// 針對特定提供者的額外選項
	Options map[string]interface{}
}

// NotificationResponse 統一的通知響應結構
type NotificationResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Details  string `json:"details,omitempty"`
	Provider string `json:"provider,omitempty"`
	Level    string `json:"level,omitempty"`
	ChatID   string `json:"chat_id,omitempty"`
	Channel  string `json:"channel,omitempty"`
}

// ProviderStatus 提供者狀態
type ProviderStatus struct {
	Name       string            `json:"name"`
	Enabled    bool              `json:"enabled"`
	Connected  bool              `json:"connected"`
	LastError  string            `json:"last_error,omitempty"`
	Channels   map[string]string `json:"channels,omitempty"`
	Statistics *ProviderStats    `json:"statistics"`
}

// ProviderStats 提供者統計數據
type ProviderStats struct {
	MessagesSent    int64 `json:"messages_sent"`
	MessagesError   int64 `json:"messages_error"`
	LastMessageTime int64 `json:"last_message_time,omitempty"`
}

// ProviderCapabilities 提供者能力
type ProviderCapabilities struct {
	SupportsLevels      bool     `json:"supports_levels"`
	SupportsChannels    bool     `json:"supports_channels"`
	SupportsRichText    bool     `json:"supports_rich_text"`
	SupportsAttachments bool     `json:"supports_attachments"`
	SupportedLanguages  []string `json:"supported_languages"`
	MaxMessageLength    int      `json:"max_message_length"`
}

// NotificationProvider 統一通知提供者接口
type NotificationProvider interface {
	GetName() string
	SendMessage(ctx context.Context, req *NotificationRequest) error
	GetStatus() *ProviderStatus
	GetCapabilities() *ProviderCapabilities
	IsEnabled() bool
	ValidateConfig() error
	TestConnection() error
}

// Ptr 輔助函數，用於返回指針
func Ptr[T any](v T) *T {
	return &v
}
