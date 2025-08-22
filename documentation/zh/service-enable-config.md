# 服務啟用配置指南

本指南說明如何正確配置各項服務的啟用狀態。

## 服務啟用邏輯

Alert-Webhooks 系統包含四個主要服務：

- **Webhooks**: HTTP API 端點服務
- **Telegram**: Telegram Bot 通知服務
- **Slack**: Slack Bot 通知服務
- **Discord**: Discord Bot 通知服務

### 預設行為

**重要**: 所有服務的 `enable` 欄位預設值為 `false`（停用）。

這樣的設計遵循「安全優先」原則：

- 只有明確配置啟用的服務才會運行
- 避免意外啟用不需要的服務
- 防止配置錯誤導致的安全問題

### 配置語法

```yaml
# 方式 1: 明確啟用服務
telegram:
  enable: true
  token: "your-token"

# 方式 2: 明確停用服務
telegram:
  enable: false
  token: "your-token"

# 方式 3: 不設定 enable 欄位（預設為 false）
telegram:
  token: "your-token"
  # enable 欄位未設定，服務將被停用

# 方式 4: 設定 enable 為空值（預設為 false）
telegram:
  enable:
  token: "your-token"
```

## 各服務配置範例

### Telegram 服務

```yaml
# 啟用 Telegram Bot
telegram:
  enable: true
  token: "your-telegram-bot-token"
  chat_ids0: "-1001234567890"
  template_mode: "full"
  template_language: "tw"

# 停用 Telegram Bot
telegram:
  enable: false
  token: "your-telegram-bot-token"
```

### Slack 服務

```yaml
# 啟用 Slack Bot
slack:
  enable: true
  token: "xoxb-your-slack-token"
  channel: "#alerts"
  template_mode: "full"
  template_language: "eng"

# 停用 Slack Bot
slack:
  enable: false
  token: "xoxb-your-slack-token"
```

### Discord 服務

```yaml
# 啟用 Discord Bot
discord:
  enable: true
  token: "your-discord-bot-token"
  guild_id: "your-server-id"
  chat_ids0: "info-channel-id"
  template_mode: "full"
  template_language: "eng"

# 停用 Discord Bot
discord:
  enable: false
  token: "your-discord-bot-token"
```

## 常見使用情境

### 情境 1: 只使用 Telegram

```yaml
webhooks:
  enable: false # 停用 Webhooks API

telegram:
  enable: true # 啟用 Telegram
  token: "your-telegram-token"
  # ... 其他配置

slack:
  enable: false # 停用 Slack

discord:
  enable: false # 停用 Discord
```

### 情境 2: 只使用 Slack

```yaml
webhooks:
  enable: false # 停用 Webhooks API

telegram:
  enable: false # 停用 Telegram

slack:
  enable: true # 啟用 Slack
  token: "your-slack-token"
  # ... 其他配置

discord:
  enable: false # 停用 Discord
```

### 情境 3: 只使用 Discord

```yaml
webhooks:
  enable: false # 停用 Webhooks API

telegram:
  enable: false # 停用 Telegram

slack:
  enable: false # 停用 Slack

discord:
  enable: true # 啟用 Discord
  token: "your-discord-token"
  # ... 其他配置
```

### 情境 4: 使用所有服務

```yaml
telegram:
  enable: true # 啟用 Telegram
  token: "your-telegram-token"
  # ... 其他配置

slack:
  enable: true # 啟用 Slack
  token: "your-slack-token"
  # ... 其他配置

discord:
  enable: true # 啟用 Discord
  token: "your-discord-token"
  # ... 其他配置
```

## 驗證服務狀態

### 查看啟動日誌

啟動應用時，系統會記錄各服務的啟用狀態：

```
Service enable status - Webhooks: true, Telegram: true, Slack: false, Discord: true
```

### API 端點檢查

您可以通過以下 API 端點檢查服務狀態：

```bash
# 檢查 Telegram 服務狀態
curl -u admin:admin http://localhost:9999/api/v1/telegram/status

# 檢查 Slack 服務狀態
curl -u admin:admin http://localhost:9999/api/v1/slack/status

# 檢查 Discord 服務狀態
curl -u admin:admin http://localhost:9999/api/v1/discord/status
```

## 最佳實踐

1. **明確配置**: 總是明確設定 `enable: true` 或 `enable: false`，避免依賴預設值
2. **環境區分**: 不同環境使用不同的配置檔案，精確控制每個環境的服務啟用狀態
3. **安全考量**: 生產環境只啟用必要的服務，減少攻擊面
4. **測試驗證**: 部署後檢查日誌和 API 端點，確認服務按預期啟用

## 故障排除

### 服務沒有啟動

檢查配置檔案中的 `enable` 欄位：

```yaml
# 確保設定為 true
telegram:
  enable: true
```

### 服務意外啟動

檢查配置檔案，確認 `enable` 欄位設定：

```yaml
# 確保設定為 false 或移除 enable 欄位
telegram:
  enable: false
```

### 無法確定服務狀態

查看應用啟動日誌，找到服務狀態記錄：

```
Service enable status - Webhooks: true, Telegram: false, Slack: true, Discord: false
```

## 相關文檔

- [Kubernetes 環境變數配置](./kubernetes-env-vars.md)
- [Telegram 設定指南](./telegram_setup.md)
- [Slack 設定指南](./slack_setup.md)
- [Discord 設定指南](./discord_setup.md)
