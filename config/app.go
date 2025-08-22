package config


type AppConf struct {
	Version        string `mapstructure:"version" json:"version"`
	AppName        string `mapstructure:"app-name" json:"app-name"`
	Mode           string `mapstructure:"mode" json:"mode"`
	Port           string `mapstructure:"port" json:"port"`
	Key           string `mapstructure:"key" json:"key"`
	Token         string `mapstructure:"token" json:"token"`
	Salt           string `mapstructure:"salt" json:"salt"`
	TrustedProxies string `mapstructure:"trusted-proxies" json:"trusted-proxies"`
}

var App AppConf
