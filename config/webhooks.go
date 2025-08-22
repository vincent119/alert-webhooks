package config




type WebhooksConf struct {
	Enable bool `mapstructure:"enable" json:"enable"`
	BaseAuthUser string `mapstructure:"base_auth_user" json:"base_auth_user"`
	BaseAuthPassword string `mapstructure:"base_auth_password" json:"base_auth_password"`
}
var Webhooks WebhooksConf