package config

// MetricConf 定義指標配置
// @Summary Metric configuration
// @Description Metric configuration
// @Tags Metric
// @ID metric-conf
type MetricConf struct {
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

// Metric 是全局指標配置
var Metric MetricConf
