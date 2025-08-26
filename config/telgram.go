package config




type TelegramConf struct {
	Enable bool `mapstructure:"enable" json:"enable"`
	Token string `mapstructure:"token" json:"token"`
	ChatIDs0 string `mapstructure:"chat_ids0" json:"chat_ids0"`
	ChatIDs1 string `mapstructure:"chat_ids1" json:"chat_ids1"`
	ChatIDs2 string `mapstructure:"chat_ids2" json:"chat_ids2"`
	ChatIDs3 string `mapstructure:"chat_ids3" json:"chat_ids3"`
	ChatIDs4 string `mapstructure:"chat_ids4" json:"chat_ids4"`
	ChatIDs5 string `mapstructure:"chat_ids5" json:"chat_ids5"`
	ChatIDs6 string `mapstructure:"chat_ids6" json:"chat_ids6"`
	TemplateMode string `mapstructure:"template_mode" json:"template_mode"`
	TemplateLanguage string `mapstructure:"template_language" json:"template_language"`
}


var Telegram TelegramConf