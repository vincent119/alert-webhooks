package config

// LogConf 定義日誌配置
type LogConf struct {
	Level         string `mapstructure:"level" json:"level"`
	Format        string `mapstructure:"format" json:"format"`
	LogPath       string `mapstructure:"log_path" json:"log_path"`
	LogFile       string `mapstructure:"log_file" json:"log_file"`
	Outputs       string `mapstructure:"outputs" json:"outputs"`
	MaxSize       int    `mapstructure:"max_size" json:"max_size"`
	MaxAge        int    `mapstructure:"max_age" json:"max_age"`
	MaxBackups    int    `mapstructure:"max_backups" json:"max_backups"`
	Compress      bool   `mapstructure:"compress" json:"compress"`
	AddCaller     bool   `mapstructure:"add_caller" json:"add_caller"`
	AddStacktrace bool   `mapstructure:"add_stacktrace" json:"add_stacktrace"`
}

// Log 是全局日誌配置
var Log LogConf
