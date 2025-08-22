package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"alert-webhooks/config"
)

func main() {
	// 範例 1: 基本配置載入
	fmt.Println("=== 基本配置載入 ===")
	config.Init()
	
	fmt.Printf("當前環境: %s\n", config.GetEnvironment())
	fmt.Printf("應用版本: %s\n", config.App.Version)
	fmt.Printf("運行模式: %s\n", config.App.Mode)
	fmt.Printf("監聽端口: %s\n", config.App.Port)
	
	// 範例 2: 環境檢查
	fmt.Println("\n=== 環境檢查 ===")
	if config.IsDevelopment() {
		fmt.Println("當前為開發環境")
	} else if config.IsProduction() {
		fmt.Println("當前為生產環境")
	} else if config.IsTest() {
		fmt.Println("當前為測試環境")
	}
	
	// 範例 3: 使用配置管理器
	fmt.Println("\n=== 配置管理器 ===")
	cm := config.NewConfigManager()
	
	// 載入特定配置文件
	if err := cm.LoadConfig("configs/config.development.yaml"); err != nil {
		log.Printf("載入配置失敗: %v", err)
	} else {
		fmt.Printf("成功載入配置: %s\n", cm.ConfigFileUsed())
		fmt.Printf("日誌級別: %s\n", cm.GetString("log.level"))
		fmt.Printf("日誌格式: %s\n", cm.GetString("log.format"))
	}
	
	// 範例 4: 動態配置更新
	fmt.Println("\n=== 動態配置更新 ===")
	
	// 註冊配置變化回調
	cm.OnConfigChange(func() {
		fmt.Println("配置已更新！")
		fmt.Printf("新的日誌級別: %s\n", cm.GetString("log.level"))
	})
	
	// 監聽配置文件變化（在實際應用中，這會在後台運行）
	// cm.WatchConfig()
	
	// 範例 5: 配置合併
	fmt.Println("\n=== 配置合併 ===")
	partialConfig := map[string]interface{}{
		"app": map[string]interface{}{
			"port": "8888",
			"mode": "debug",
		},
		"log": map[string]interface{}{
			"level": "trace",
		},
	}
	
	if err := config.MergeConfig(partialConfig); err != nil {
		log.Printf("合併配置失敗: %v", err)
	} else {
		fmt.Printf("合併後端口: %s\n", config.App.Port)
		fmt.Printf("合併後模式: %s\n", config.App.Mode)
		fmt.Printf("合併後日誌級別: %s\n", config.Log.Level)
	}
	
	// 範例 6: 環境變數覆蓋
	fmt.Println("\n=== 環境變數覆蓋 ===")
	
	// 設定環境變數（在實際應用中，這些會在系統中設定）
	os.Setenv("APP_PORT", "7777")
	os.Setenv("LOG_LEVEL", "warn")
	
	// 重新初始化配置以讀取環境變數
	config.Init()
	
	fmt.Printf("環境變數覆蓋後端口: %s\n", config.App.Port)
	fmt.Printf("環境變數覆蓋後日誌級別: %s\n", config.Log.Level)
	
	// 範例 7: 配置驗證
	fmt.Println("\n=== 配置驗證 ===")
	validateConfig()
	
	// 範例 8: 配置導出
	fmt.Println("\n=== 配置導出 ===")
	exportConfig(cm)
}

// validateConfig 驗證配置
func validateConfig() {
	// 檢查必要配置
	if config.App.Port == "" {
		fmt.Println("警告: 應用端口未設定")
	}
	
	if config.App.Key == "" {
		fmt.Println("警告: 應用密鑰未設定")
	}
	
	if config.Log.Level == "" {
		fmt.Println("警告: 日誌級別未設定")
	}
	
	// 檢查配置合理性
	if config.Log.MaxSize <= 0 {
		fmt.Println("警告: 日誌文件大小設定不合理")
	}
	
	if config.Log.MaxAge <= 0 {
		fmt.Println("警告: 日誌保留天數設定不合理")
	}
	
	fmt.Println("配置驗證完成")
}

// exportConfig 導出配置
func exportConfig(cm *config.ConfigManager) {
	// 獲取所有配置
	allSettings := cm.AllSettings()
	
	fmt.Println("當前所有配置:")
	for key, value := range allSettings {
		fmt.Printf("  %s: %v\n", key, value)
	}
	
	// 導出到新文件
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("config_export_%s.yaml", timestamp)
	
	if err := cm.WriteConfigAs(filename); err != nil {
		log.Printf("導出配置失敗: %v", err)
	} else {
		fmt.Printf("配置已導出到: %s\n", filename)
	}
}

// 使用說明
func printUsage() {
	fmt.Println(`
配置系統使用說明:

1. 命令行參數:
   -c, --config: 指定配置文件路徑
   -e, --env: 指定環境 (dev, test, prod)

2. 環境變數:
   APP_ENV: 設定環境
   GO_ENV: 設定環境 (備用)
   APP_PORT: 覆蓋應用端口
   LOG_LEVEL: 覆蓋日誌級別

3. 配置文件搜尋順序:
   - 命令行指定的配置文件
   - configs/{env}/config.yaml
   - configs/config.{env}.yaml
   - config/config.yaml
   - config.yaml

4. 配置檔案結構:
   configs/
   ├── config.development.yaml
   ├── config.production.yaml
   └── config.test.yaml

5. 主要功能:
   - 多環境配置支援
   - 環境變數覆蓋
   - 配置文件監聽
   - 動態配置更新
   - 配置合併
   - 配置驗證
   - 配置導出
`)
}
