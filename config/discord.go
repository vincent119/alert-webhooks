package config

// DiscordConf holds Discord service configuration
type DiscordConf struct {
	Enable   bool   `json:"enable" yaml:"enable" mapstructure:"enable"`           // Enable Discord notifications
	Token    string `json:"token" yaml:"token" mapstructure:"token"`             // Discord bot token
	GuildID  string `json:"guild_id" yaml:"guild_id" mapstructure:"guild_id"`    // Discord server/guild ID
	Username string `json:"username" yaml:"username" mapstructure:"username"`    // Bot username
	
	// Channel mappings (chat_ids0-5 for alert levels 0-5)
	Channels map[string]string `json:"channels" yaml:"channels" mapstructure:"channels"`
	
	// Discord specific options
	MessageFormat string   `json:"message_format" yaml:"message_format" mapstructure:"message_format"` // Message format (markdown/text)
	MentionRoles  []string `json:"mention_roles" yaml:"mention_roles" mapstructure:"mention_roles"`    // Role IDs to mention
	
	// Template configuration
	TemplateMode     string `json:"template_mode" yaml:"template_mode" mapstructure:"template_mode"`           // Template mode (minimal/full)
	TemplateLanguage string `json:"template_language" yaml:"template_language" mapstructure:"template_language"` // Template language (eng/tw/zh/ja/ko)
}