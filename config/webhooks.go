package config

type WebhooksConf struct {
	Enable             bool     `mapstructure:"enable" json:"enable"`
	BaseAuthUser       string   `mapstructure:"base_auth_user" json:"base_auth_user"`
	BaseAuthPassword   string   `mapstructure:"base_auth_password" json:"base_auth_password"`
	CORSAllowedOrigins []string `mapstructure:"cors_allowed_origins" json:"cors_allowed_origins"`
}

var Webhooks WebhooksConf
