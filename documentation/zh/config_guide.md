# 多設定檔配置系統使用指南

## 概述

本配置系統支援多環境配置管理，提供靈活的配置載入、環境變數覆蓋、動態更新等功能。

## 目錄結構

```
alert-webhooks/
├── config/                      # 配置管理模組
│   ├── config.go               # 核心配置載入邏輯
│   ├── manager.go              # 配置管理器
│   ├── app.go                  # 應用配置結構
│   ├── metric.go               # 指標配置結構
│   ├── logger.go               # 日誌配置結構
│   ├── discord.go              # Discord 配置結構
│   ├── slack.go                # Slack 配置結構
│   ├── telgram.go              # Telegram 配置結構
│   └── webhooks.go             # Webhook 配置結構
├── configs/                     # 配置文件目錄
│   ├── config.development.yaml # 開發環境配置
│   ├── alert_config.yaml       # 完整警報配置範例
│   └── alert_config.minimal.yaml # 最小警報配置範例
└── examples/                    # 使用範例
    ├── config_usage.go         # 配置使用範例
    └── config.expamle.yaml     # 配置範例文件
```

## 快速開始

### 1. 基本使用

```go
package main

import "alert-webhooks/config"

func main() {
    // 初始化配置
    config.Init()

    // 使用配置
    fmt.Printf("端口: %s\n", config.App.Port)
    fmt.Printf("日誌級別: %s\n", config.Log.Level)
}
```

### 2. 配置文件設置

```bash
# 複製範例配置文件
cp examples/config.expamle.yaml configs/config.development.yaml

# 編輯配置文件
vim configs/config.development.yaml
```

### 3. 指定環境

```bash
# 使用命令行參數
./alert-webhooks -e development
./alert-webhooks -e production

# 使用環境變數
export APP_ENV=development
./alert-webhooks
```

### 4. 指定配置文件

```bash
./alert-webhooks -c /path/to/custom/config.yaml
```

## 配置檔案格式

### 開發環境配置 (configs/config.development.yaml)

```yaml
app:
  version: "1.2.3-dev"
  mode: "development"
  port: "9999"
  key: "dev-key"
  token: "dev-token"
  salt: "dev-salt"
  trusted_proxies: "127.0.0.1,10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"

log:
  level: "debug"
  format: "console"
  outputs: "console,file"
  log_path: "./logs"
  log_file: "server-dev.log"
  max_size: 100
  max_age: 7
  max_backups: 5
  compress: false
  add_caller: true
  add_stacktrace: true

metric:
  user: "dev-admin"
  password: "dev-password"

# Webhook 服務配置
webhooks:
  enable: true
  base_auth_user: "admin"
  base_auth_password: "admin"

# Telegram 配置
telegram:
  enable: true
  token: "your-telegram-bot-token"
  chat_ids0: "-1001234567890"
  template_mode: "full"
  template_language: "tw"

# Slack 配置
slack:
  enable: true
  token: "xoxb-your-slack-token"
  channel: "#alerts"
  template_mode: "full"
  template_language: "eng"

# Discord 配置
discord:
  enable: true
  token: "your-discord-bot-token"
  channel_id: "123456789012345678"
  template_mode: "full"
  template_language: "eng"
```

### 生產環境配置 (configs/config.production.yaml)

```yaml
app:
  version: "1.2.3"
  mode: "production"
  port: "8080"
  key: "prod-key"
  token: "prod-token"
  salt: "prod-salt"
  trusted_proxies: "10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"

log:
  level: "info"
  format: "json"
  outputs: "file"
  log_path: "/var/log/alert-webhooks"
  log_file: "server.log"
  max_size: 500
  max_age: 30
  max_backups: 10
  compress: true
  add_caller: true
  add_stacktrace: false

metric:
  user: "prod-admin"
  password: "prod-password"
```

## 進階功能

### 1. 配置管理器

```go
// 創建配置管理器
cm := config.NewConfigManager()

// 載入特定配置文件
err := cm.LoadConfig("configs/config.development.yaml")

// 監聽配置文件變化
cm.OnConfigChange(func() {
    fmt.Println("配置已更新！")
})
cm.WatchConfig()

// 獲取配置值
port := cm.GetString("app.port")
level := cm.GetString("log.level")
```

### 2. 動態配置更新

```go
// 合併部分配置
partialConfig := map[string]interface{}{
    "app": map[string]interface{}{
        "port": "8888",
        "mode": "debug",
    },
    "log": map[string]interface{}{
        "level": "trace",
    },
}

err := config.MergeConfig(partialConfig)
```

### 3. 環境變數覆蓋

```bash
# 設定環境變數
export APP_PORT=7777
export LOG_LEVEL=warn
export APP_MODE=debug

# 運行應用
./alert-webhooks
```

### 4. 配置驗證

```go
// 檢查環境
if config.IsDevelopment() {
    fmt.Println("開發環境")
}

if config.IsProduction() {
    fmt.Println("生產環境")
}

// 獲取當前環境
env := config.GetEnvironment()
```

## 配置載入順序

1. **命令行指定的配置文件** (`-c` 參數)
2. **環境特定配置** (`configs/config.{env}.yaml`)
3. **預設配置** (`config.yaml`)
4. **環境變數覆蓋**

## 環境變數對應

配置項可以通過環境變數覆蓋，使用點號分隔的鍵名轉換為下劃線：

```yaml
# 配置文件
app:
  port: "8080"
  mode: "production"

log:
  level: "info"
```

對應的環境變數：

```bash
export APP_PORT=9999
export APP_MODE=development
export LOG_LEVEL=debug
```

## 最佳實踐

### 1. 配置檔案組織

- 將不同環境的配置分離到不同檔案
- 使用有意義的檔案名稱
- 保持配置結構一致

### 2. 敏感資訊處理

- 不要在配置檔案中存放密碼等敏感資訊
- 使用環境變數或外部密鑰管理系統
- 考慮使用配置加密

### 3. 配置驗證

- 在應用啟動時驗證必要配置
- 檢查配置值的合理性
- 提供清晰的錯誤訊息

### 4. 配置監聽

- 在開發環境中使用配置監聽
- 生產環境謹慎使用動態配置更新
- 確保配置更新的原子性

## 故障排除

### 常見問題

1. **配置文件找不到**

   - 檢查檔案路徑是否正確
   - 確認檔案權限
   - 檢查檔案格式是否為 YAML

2. **配置值未生效**

   - 檢查環境變數是否正確設定
   - 確認配置載入順序
   - 檢查配置結構是否匹配

3. **配置監聽不工作**
   - 確認檔案系統支援 inotify
   - 檢查檔案權限
   - 確認監聽路徑正確

### 調試技巧

```go
// 啟用詳細日誌
log.SetLevel(log.DebugLevel)

// 檢查配置載入路徑
paths := config.GetConfigPaths()
fmt.Printf("配置搜尋路徑: %v\n", paths)

// 檢查使用的配置文件
fmt.Printf("使用的配置文件: %s\n", cm.ConfigFileUsed())
```

## 範例程式碼

完整的範例程式碼請參考 `examples/config_usage.go` 檔案。
