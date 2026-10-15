package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"

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
	App      AppConf
	Metric   MetricConf
	Trace    TraceConf
	Log      LogConf
	Telegram TelegramConf
	Webhooks WebhooksConf
	Slack    SlackConf
	Discord  DiscordConf
}

// Config 表示完整的 typed 應用程式設定快照。
type Config struct {
	App      AppConf      `mapstructure:"app" json:"app"`
	Metric   MetricConf   `mapstructure:"metric" json:"metric"`
	Trace    TraceConf    `mapstructure:"trace" json:"trace"`
	Log      LogConf      `mapstructure:"log" json:"log"`
	Telegram TelegramConf `mapstructure:"telegram" json:"telegram"`
	Webhooks WebhooksConf `mapstructure:"webhooks" json:"webhooks"`
	Slack    SlackConf    `mapstructure:"slack" json:"slack"`
	Discord  DiscordConf  `mapstructure:"discord" json:"discord"`
}

// TraceConf 定義追蹤匯出與驗證設定。
type TraceConf struct {
	Enable        bool    `mapstructure:"enable"`
	Url           string  `mapstructure:"url"`
	Port          string  `mapstructure:"port"`
	UrlPath       string  `mapstructure:"urlPath"`
	Insecure      bool    `mapstructure:"insecure"`
	TlsSkipVerify bool    `mapstructure:"tlsSkipVerify"`
	AuthUser      string  `mapstructure:"authUser"`
	AuthPasswd    string  `mapstructure:"authPasswd"`
	SampleRate    float64 `mapstructure:"sampleRate"`
}

// Trace 是全局追蹤配置
var Trace TraceConf

// 存儲內部配置
var (
	once        sync.Once
	initErr     error
	configPaths []string
	configStore atomic.Pointer[Config]
)

// Init 初始化配置
func Init() error {
	once.Do(func() {
		initErr = initConfig(true)
	})
	return initErr
}

// ForceReload 強制重新載入配置（用於熱加載）
func ForceReload() error {
	return initConfig(false)
}

// initConfig 實際的配置初始化邏輯
func initConfig(updateLegacyGlobals bool) error {
	// 解析命令行參數
	pflag.Parse()

	// 創建新的 Viper 實例
	v := viper.New()
	v.AutomaticEnv()

	// 設定配置文件路徑
	setupConfigPaths(v)

	// Load configuration
	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("讀取設定檔失敗: %w", err)
	}

	// Parse configuration into struct
	var candidate Config
	if err := v.Unmarshal(&candidate); err != nil {
		return fmt.Errorf("解析設定檔失敗: %w", err)
	}

	// 從環境變數覆蓋敏感配置
	overrideWithEnvVars(&candidate)

	publish(candidate, updateLegacyGlobals)

	// 確保加載成功
	if v.ConfigFileUsed() != "" {
		fmt.Printf("Successfully loaded config: %s\n", v.ConfigFileUsed())
	} else {
		fmt.Println("No valid configuration file found, using default values")
	}

	// 調試：打印 Telegram 配置
	fmt.Printf("Config loaded - template_language: %s, template_mode: %s, config_file: %s\n",
		candidate.Telegram.TemplateLanguage, candidate.Telegram.TemplateMode, v.ConfigFileUsed())

	return nil
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
		".",                        // 當前目錄
		"./config",                 // config 子目錄
		"./configs",                // configs 子目錄
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
func overrideWithEnvVars(candidate *Config) {
	// Metric 配置
	if user := os.Getenv("METRIC_USER"); user != "" {
		candidate.Metric.User = user
		fmt.Printf("Override metric user from env var: %s\n", user)
	}
	if password := os.Getenv("METRIC_PASSWORD"); password != "" {
		candidate.Metric.Password = password
		fmt.Printf("Override metric password from env var: [REDACTED]\n")
	}

	// Webhooks 配置
	if user := os.Getenv("WEBHOOKS_USER"); user != "" {
		candidate.Webhooks.BaseAuthUser = user
		fmt.Printf("Override webhooks user from env var: %s\n", user)
	}
	if password := os.Getenv("WEBHOOKS_PASSWORD"); password != "" {
		candidate.Webhooks.BaseAuthPassword = password
		fmt.Printf("Override webhooks password from env var: [REDACTED]\n")
	}
	if origins, ok := os.LookupEnv("WEBHOOKS_CORS_ALLOWED_ORIGINS"); ok {
		candidate.Webhooks.CORSAllowedOrigins = parseCORSAllowedOrigins(origins)
	}

	// Trace 配置
	if authUser := os.Getenv("TRACE_AUTH_USER"); authUser != "" {
		candidate.Trace.AuthUser = authUser
		fmt.Printf("Override trace auth user from env var: %s\n", authUser)
	}
	if authPasswd := os.Getenv("TRACE_AUTH_PASSWD"); authPasswd != "" {
		candidate.Trace.AuthPasswd = authPasswd
		fmt.Printf("Override trace auth password from env var: [REDACTED]\n")
	}

	// Telegram 配置
	if token := os.Getenv("TELEGRAM_TOKEN"); token != "" {
		candidate.Telegram.Token = token
		fmt.Printf("Override telegram token from env var: [REDACTED]\n")
	}

	// Slack 配置
	if token := os.Getenv("SLACK_TOKEN"); token != "" {
		candidate.Slack.Token = token
		fmt.Printf("Override slack token from env var: [REDACTED]\n")
	}

	// Discord 配置
	if token := os.Getenv("DISCORD_TOKEN"); token != "" {
		candidate.Discord.Token = token
		fmt.Printf("Override discord token from env var: [REDACTED]\n")
	}

	// 記錄服務啟用狀態（確認預設值邏輯）
	fmt.Printf("Service enable status - Webhooks: %t, Telegram: %t, Slack: %t, Discord: %t\n",
		candidate.Webhooks.Enable, candidate.Telegram.Enable, candidate.Slack.Enable, candidate.Discord.Enable)
}

// publish 只在候選設定完整載入與覆蓋完成後，才原子發布設定快照。
func publish(candidate Config, updateLegacyGlobals bool) {
	immutable := cloneConfig(candidate)
	configStore.Store(&immutable)

	if !updateLegacyGlobals {
		return
	}

	// 舊欄位僅於啟動階段設定，避免熱重載與讀取端並行寫入。
	// 更新包級變數
	App = immutable.App
	Metric = immutable.Metric
	Trace = immutable.Trace
	Log = immutable.Log
	Telegram = immutable.Telegram
	Webhooks = immutable.Webhooks
	Slack = immutable.Slack

	// 更新 Conf 結構體
	Conf.App = immutable.App
	Conf.Metric = immutable.Metric
	Conf.Trace = immutable.Trace
	Conf.Log = immutable.Log
	Conf.Telegram = immutable.Telegram
	Conf.Webhooks = immutable.Webhooks
	Conf.Slack = immutable.Slack
	Conf.Discord = immutable.Discord
}

// Current 回傳防禦性複製，確保呼叫端不會修改已發布的快照。
func Current() Config {
	current := configStore.Load()
	if current == nil {
		// 初始化前保留既有測試與範例的讀取行為；正式啟動完成後不再讀取舊全域欄位。
		return cloneConfig(Config{
			App:      App,
			Metric:   Metric,
			Trace:    Trace,
			Log:      Log,
			Telegram: Telegram,
			Webhooks: Webhooks,
			Slack:    Slack,
			Discord:  Conf.Discord,
		})
	}
	return cloneConfig(*current)
}

func cloneConfig(source Config) Config {
	cloned := source
	cloned.Webhooks.CORSAllowedOrigins = append([]string(nil), source.Webhooks.CORSAllowedOrigins...)
	cloned.Slack.Channels = cloneStringMap(source.Slack.Channels)
	cloned.Discord.Channels = cloneStringMap(source.Discord.Channels)
	cloned.Discord.MentionRoles = append([]string(nil), source.Discord.MentionRoles...)
	return cloned
}

func parseCORSAllowedOrigins(value string) []string {
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		origin := strings.TrimSpace(part)
		if origin == "" || origin == "*" {
			continue
		}
		origins = append(origins, origin)
	}
	return origins
}

func cloneStringMap(source map[string]string) map[string]string {
	if source == nil {
		return nil
	}
	cloned := make(map[string]string, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

// GetFullConfig 保留既有完整設定讀取介面。
func GetFullConfig() Config {
	return Current()
}

// GetEnvironment 獲取當前環境
func GetEnvironment() string {
	return getEnvironment()
}

// IsDevelopment 檢查是否為開發環境
func IsDevelopment() bool {
	return Current().App.Mode == "development"
}

// IsProduction 檢查是否為生產環境
func IsProduction() bool {
	return Current().App.Mode == "production"
}

// IsTest 檢查是否為測試環境
func IsTest() bool {
	return Current().App.Mode == "test"
}

// LoadConfigFromFile 從指定檔案載入配置（用於測試或動態配置）
func LoadConfigFromFile(filePath string) error {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("讀取設定檔 %s 失敗: %w", filePath, err)
	}

	var candidate Config
	if err := v.Unmarshal(&candidate); err != nil {
		return fmt.Errorf("解析設定檔 %s 失敗: %w", filePath, err)
	}

	publish(candidate, false)
	return nil
}

// MergeConfig 合併配置（用於動態更新配置）
func MergeConfig(partialConfig map[string]interface{}) error {
	// 將部分配置轉換為 configStruct
	tempViper := viper.New()
	for key, value := range partialConfig {
		tempViper.Set(key, value)
	}

	var tempConfig Config
	if err := tempViper.Unmarshal(&tempConfig); err != nil {
		return fmt.Errorf("解析部分設定失敗: %w", err)
	}

	// 合併配置
	current := Current()
	mergeConfigStruct(&current, &tempConfig)

	// 更新全局變數
	publish(current, false)
	return nil
}

// mergeConfigStruct 合併兩個配置結構體
func mergeConfigStruct(target, source *Config) {
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
