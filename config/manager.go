package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	viper      *viper.Viper
	configFile string
	watcher    *fsnotify.Watcher
	callbacks  []func()
}

// NewConfigManager 創建新的配置管理器
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		viper:     viper.New(),
		callbacks: make([]func(), 0),
	}
}

// LoadConfig 載入配置
func (cm *ConfigManager) LoadConfig(configPath string) error {
	cm.configFile = configPath
	
	// 設定配置文件
	cm.viper.SetConfigFile(configPath)
	cm.viper.SetConfigType("yaml")
	
	// 自動讀取環境變數
	cm.viper.AutomaticEnv()
	cm.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// 讀取配置文件
	if err := cm.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}
	
	// 解析到結構體
	if err := cm.viper.Unmarshal(&confInternal); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}
	
	// 更新全局配置
	updateGlobalConfigs()
	
	return nil
}

// WatchConfig 監聽配置文件變化
func (cm *ConfigManager) WatchConfig() error {
	if cm.configFile == "" {
		return fmt.Errorf("no config file set")
	}
	
	// 創建文件監聽器
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %v", err)
	}
	
	cm.watcher = watcher
	
	// 監聽配置文件目錄
	configDir := filepath.Dir(cm.configFile)
	if err := watcher.Add(configDir); err != nil {
		return fmt.Errorf("failed to watch config directory: %v", err)
	}
	
	// 啟動監聽協程
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				
				// 檢查是否是我們的配置文件
				if event.Name == cm.configFile && (event.Op&fsnotify.Write == fsnotify.Write) {
					fmt.Printf("Config file changed: %s\n", event.Name)
					
					// 重新載入配置
					if err := cm.LoadConfig(cm.configFile); err != nil {
						fmt.Printf("Failed to reload config: %v\n", err)
						continue
					}
					
					// 執行回調函數
					for _, callback := range cm.callbacks {
						callback()
					}
				}
				
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Printf("Config watcher error: %v\n", err)
			}
		}
	}()
	
	return nil
}

// OnConfigChange 註冊配置變化回調函數
func (cm *ConfigManager) OnConfigChange(callback func()) {
	cm.callbacks = append(cm.callbacks, callback)
}

// Close 關閉配置管理器
func (cm *ConfigManager) Close() error {
	if cm.watcher != nil {
		return cm.watcher.Close()
	}
	return nil
}

// GetString 獲取字符串配置
func (cm *ConfigManager) GetString(key string) string {
	return cm.viper.GetString(key)
}

// GetInt 獲取整數配置
func (cm *ConfigManager) GetInt(key string) int {
	return cm.viper.GetInt(key)
}

// GetBool 獲取布爾配置
func (cm *ConfigManager) GetBool(key string) bool {
	return cm.viper.GetBool(key)
}

// GetDuration 獲取時間配置
func (cm *ConfigManager) GetDuration(key string) time.Duration {
	return cm.viper.GetDuration(key)
}

// Set 設定配置值
func (cm *ConfigManager) Set(key string, value interface{}) {
	cm.viper.Set(key, value)
}

// WriteConfig 寫入配置文件
func (cm *ConfigManager) WriteConfig() error {
	return cm.viper.WriteConfig()
}

// WriteConfigAs 寫入指定格式的配置文件
func (cm *ConfigManager) WriteConfigAs(filename string) error {
	return cm.viper.WriteConfigAs(filename)
}

// SafeWriteConfig 安全寫入配置文件（如果文件不存在）
func (cm *ConfigManager) SafeWriteConfig() error {
	return cm.viper.SafeWriteConfig()
}

// SafeWriteConfigAs 安全寫入指定格式的配置文件
func (cm *ConfigManager) SafeWriteConfigAs(filename string) error {
	return cm.viper.SafeWriteConfigAs(filename)
}

// ConfigFileUsed 獲取使用的配置文件路徑
func (cm *ConfigManager) ConfigFileUsed() string {
	return cm.viper.ConfigFileUsed()
}

// AllSettings 獲取所有配置
func (cm *ConfigManager) AllSettings() map[string]interface{} {
	return cm.viper.AllSettings()
}

// IsSet 檢查配置是否已設定
func (cm *ConfigManager) IsSet(key string) bool {
	return cm.viper.IsSet(key)
}

// GetViper 獲取底層的 Viper 實例
func (cm *ConfigManager) GetViper() *viper.Viper {
	return cm.viper
}
