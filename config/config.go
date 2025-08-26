package config

import (
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// 設定命令行標誌 (flags)，允許指定配置文件路徑
var (
	configFile = pflag.StringP("config", "c", "", "Path to configuration file")
	env        = pflag.StringP("env", "e", "", "Environment (dev, test, prod)")
)

// Conf 是全局配置的容器，為了保持向後兼容
var Conf struct {
	App    AppConf
	Metric MetricConf
	Log    LogConf
	Telegram TelegramConf
	Webhooks WebhooksConf
	Slack  SlackConf
	Discord DiscordConf
}

// 內部使用的配置結構體
type configStruct struct {
	App    AppConf    `mapstructure:"app" json:"app"`
	Metric MetricConf `mapstructure:"metric" json:"metric"`
	Log    LogConf    `mapstructure:"log" json:"log"`
	Telegram TelegramConf `mapstructure:"telegram" json:"telegram"`
	Webhooks WebhooksConf `mapstructure:"webhooks" json:"webhooks"`
	Slack  SlackConf  `mapstructure:"slack" json:"slack"`
	Discord DiscordConf `mapstructure:"discord" json:"discord"`
}

// 存儲內部配置
var (
	confInternal configStruct
	once         sync.Once
	configPaths  []string
)

// Init 初始化配置
func Init() {
	once.Do(func() {
		initConfig()
	})
}

// ForceReload 強制重新載入配置（用於熱加載）
func ForceReload() {
	initConfig()
}

// initConfig 實際的配置初始化邏輯
func initConfig() {
	// 解析命令行參數
	pflag.Parse()

	// 創建新的 Viper 實例
	v := viper.New()
	v.AutomaticEnv()

	// 設定配置文件路徑
	setupConfigPaths(v)

	// Load configuration
	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read configuration file: %v", err)
	}

	// Parse configuration into struct
	if err := v.Unmarshal(&confInternal); err != nil {
		log.Fatalf("Failed to parse configuration file: %v", err)
	}

	// 從環境變數覆蓋敏感配置
	overrideWithEnvVars()

	// 更新所有全局配置變數
	updateGlobalConfigs()

	// 確保加載成功
	if v.ConfigFileUsed() != "" {
		fmt.Printf("Successfully loaded config: %s\n", v.ConfigFileUsed())
	} else {
		log.Println("No valid configuration file found, using default values")
	}
	
	// 調試：打印 Telegram 配置
	fmt.Printf("Config loaded - template_language: %s, template_mode: %s, config_file: %s\n",
		confInternal.Telegram.TemplateLanguage, confInternal.Telegram.TemplateMode, v.ConfigFileUsed())
}

// setupConfigPaths 設定配置檔案路徑，支援多設定檔
func setupConfigPaths(v *viper.Viper) {
	// 如果指定了具體的配置文件
	if *configFile != "" {
		v.SetConfigFile(*configFile)
		return
	}

	// 設定配置文件名稱和類型
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// 獲取環境
	environment := getEnvironment()

	// 設定配置文件搜尋路徑
	configPaths = []string{
		".",                    // 當前目錄
		"./config",            // config 子目錄
		"./configs",           // configs 子目錄
		"./configs/" + environment, // 環境特定目錄
	}

	// 添加搜尋路徑
	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// 如果指定了環境，嘗試載入環境特定的配置文件
	if environment != "" {
		// 嘗試載入 config.{env}.yaml
		envConfigName := fmt.Sprintf("config.%s", environment)
		v.SetConfigName(envConfigName)
		
		// 如果環境特定配置不存在，回退到預設配置
		if err := v.ReadInConfig(); err != nil {
			v.SetConfigName("config")
		}
	}
}

// getEnvironment 獲取當前環境
func getEnvironment() string {
	// 優先使用命令行參數
	if *env != "" {
		return *env
	}

	// 其次使用環境變數
	if envVar := os.Getenv("APP_ENV"); envVar != "" {
		return envVar
	}

	// 最後使用 GO_ENV
	if envVar := os.Getenv("GO_ENV"); envVar != "" {
		return envVar
	}

	return "development" // 預設環境
}

// overrideWithEnvVars 從環境變數覆蓋敏感配置
func overrideWithEnvVars() {
	// Metric 配置
	if user := os.Getenv("METRIC_USER"); user != "" {
		confInternal.Metric.User = user
		fmt.Printf("Override metric user from env var: %s\n", user)
	}
	if password := os.Getenv("METRIC_PASSWORD"); password != "" {
		confInternal.Metric.Password = password
		fmt.Printf("Override metric password from env var: [REDACTED]\n")
	}

	// Webhooks 配置
	if user := os.Getenv("WEBHOOKS_USER"); user != "" {
		confInternal.Webhooks.BaseAuthUser = user
		fmt.Printf("Override webhooks user from env var: %s\n", user)
	}
	if password := os.Getenv("WEBHOOKS_PASSWORD"); password != "" {
		confInternal.Webhooks.BaseAuthPassword = password
		fmt.Printf("Override webhooks password from env var: [REDACTED]\n")
	}

	// Telegram 配置
	if token := os.Getenv("TELEGRAM_TOKEN"); token != "" {
		confInternal.Telegram.Token = token
		fmt.Printf("Override telegram token from env var: [REDACTED]\n")
	}

	// Slack 配置
	if token := os.Getenv("SLACK_TOKEN"); token != "" {
		confInternal.Slack.Token = token
		fmt.Printf("Override slack token from env var: [REDACTED]\n")
	}

	// Discord 配置
	if token := os.Getenv("DISCORD_TOKEN"); token != "" {
		confInternal.Discord.Token = token
		fmt.Printf("Override discord token from env var: [REDACTED]\n")
	}

	// 記錄服務啟用狀態（確認預設值邏輯）
	fmt.Printf("Service enable status - Webhooks: %t, Telegram: %t, Slack: %t, Discord: %t\n", 
		confInternal.Webhooks.Enable, confInternal.Telegram.Enable, confInternal.Slack.Enable, confInternal.Discord.Enable)
}

// updateGlobalConfigs 更新所有全局配置變數
func updateGlobalConfigs() {
	// 更新包級變數
	App = confInternal.App
	Metric = confInternal.Metric
	Log = confInternal.Log
	Telegram = confInternal.Telegram
	Webhooks = confInternal.Webhooks
	Slack = confInternal.Slack

	// 更新 Conf 結構體
	Conf.App = confInternal.App
	Conf.Metric = confInternal.Metric
	Conf.Log = confInternal.Log
	Conf.Telegram = confInternal.Telegram
	Conf.Webhooks = confInternal.Webhooks
	Conf.Slack = confInternal.Slack
	Conf.Discord = confInternal.Discord
}

// GetFullConfig 返回完整配置，對於需要訪問完整配置的情況
func GetFullConfig() configStruct {
	return confInternal
}

// GetEnvironment 獲取當前環境
func GetEnvironment() string {
	return getEnvironment()
}

// IsDevelopment 檢查是否為開發環境
func IsDevelopment() bool {
	return getEnvironment() == "development"
}

// IsProduction 檢查是否為生產環境
func IsProduction() bool {
	return getEnvironment() == "production"
}

// IsTest 檢查是否為測試環境
func IsTest() bool {
	return getEnvironment() == "test"
}

// LoadConfigFromFile 從指定檔案載入配置（用於測試或動態配置）
func LoadConfigFromFile(filePath string) error {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file %s: %v", filePath, err)
	}

	if err := v.Unmarshal(&confInternal); err != nil {
		return fmt.Errorf("failed to parse config file %s: %v", filePath, err)
	}

	updateGlobalConfigs()
	return nil
}

// MergeConfig 合併配置（用於動態更新配置）
func MergeConfig(partialConfig map[string]interface{}) error {
	// 將部分配置轉換為 configStruct
	tempViper := viper.New()
	for key, value := range partialConfig {
		tempViper.Set(key, value)
	}

	var tempConfig configStruct
	if err := tempViper.Unmarshal(&tempConfig); err != nil {
		return fmt.Errorf("failed to parse partial config: %v", err)
	}

	// 合併配置
	mergeConfigStruct(&confInternal, &tempConfig)
	
	// 更新全局變數
	updateGlobalConfigs()
	return nil
}

// mergeConfigStruct 合併兩個配置結構體
func mergeConfigStruct(target, source *configStruct) {
	// 這裡可以實現更複雜的合併邏輯
	// 目前使用簡單的覆蓋策略
	if source.App.Version != "" {
		target.App.Version = source.App.Version
	}
	if source.App.Mode != "" {
		target.App.Mode = source.App.Mode
	}
	if source.App.Port != "" {
		target.App.Port = source.App.Port
	}
	// 可以繼續添加其他欄位的合併邏輯
}

// GetConfigPaths 獲取配置檔案搜尋路徑
func GetConfigPaths() []string {
	return configPaths
}

// 獲取專案根目錄
func getProjectRoot() string {
	// 嘗試從環境變數獲取
	if root := os.Getenv("PROJECT_ROOT"); root != "" {
		return root
	}

	// 通過當前工作目錄判斷
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}

	return wd
}


