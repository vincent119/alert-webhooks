package config

// WebhooksConf 定義 Webhook HTTP 服務設定。
type WebhooksConf struct {
	Enable             bool     `mapstructure:"enable" json:"enable"`
	BaseAuthUser       string   `mapstructure:"base_auth_user" json:"base_auth_user"`
	BaseAuthPassword   string   `mapstructure:"base_auth_password" json:"base_auth_password"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins" json:"cors_allowed_origins"`
}

// Webhooks 保存目前 Webhook 設定的相容存取入口。
var Webhooks WebhooksConf
