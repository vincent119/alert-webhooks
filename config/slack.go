package config

// SlackConf 定義 Slack 通知服務設定。
type SlackConf struct {
	Enable           bool              `mapstructure:"enable" json:"enable"`
	Token            string            `mapstructure:"token" json:"token"`
	Channel          string            `mapstructure:"channel" json:"channel"`
	Username         string            `mapstructure:"username" json:"username"`                   // Bot 顯示名稱
	IconURL          string            `mapstructure:"icon_url" json:"icon_url"`                   // Bot 頭像 URL
	IconEmoji        string            `mapstructure:"icon_emoji" json:"icon_emoji"`               // Bot 表情符號
	Channels         map[string]string `mapstructure:"channels" json:"channels"`                   // 多頻道支持 (level -> channel)
	ThreadTS         bool              `mapstructure:"thread_ts" json:"thread_ts"`                 // 是否使用線程回覆
	LinkNames        bool              `mapstructure:"link_names" json:"link_names"`               // 是否連結 @mentions
	UnfurlLinks      bool              `mapstructure:"unfurl_links" json:"unfurl_links"`           // 是否展開連結預覽
	UnfurlMedia      bool              `mapstructure:"unfurl_media" json:"unfurl_media"`           // 是否展開媒體預覽
	TemplateMode     string            `mapstructure:"template_mode" json:"template_mode"`         // 模板模式 (minimal, full)
	TemplateLanguage string            `mapstructure:"template_language" json:"template_language"` // 模板語言 (eng, tw, zh, ja, ko)
}

// Slack 保存目前 Slack 設定的相容存取入口。
var Slack SlackConf
