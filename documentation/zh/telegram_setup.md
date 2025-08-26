# Telegram Bot 設定指南

## 概述

本專案支援透過 Telegram Bot 發送警報訊息。本文檔將指導您如何設定 Telegram Bot。

## 步驟 1：創建 Telegram Bot

### 1. 與 BotFather 對話

1. 在 Telegram 中搜尋 `@BotFather`
2. 點擊 "Start" 開始對話
3. 發送 `/newbot` 命令

### 2. 設定 Bot 資訊

BotFather 會要求您提供以下資訊：

- **Bot 名稱**：例如 "Alert Webhooks Bot"
- **Bot 用戶名**：例如 "alert_webhooks_bot"（必須以 `_bot` 結尾）

### 3. 獲取 Bot Token

創建成功後，BotFather 會提供一個 Bot Token，格式如下：

```
123456789:ABCdefGHIjklMNOpqrsTUVwxyz
```

**重要：請妥善保管此 Token，不要分享給他人！**

## 步驟 2：配置 Bot Token

### 1. 更新配置檔案

編輯 `configs/config.development.yaml`：

```yaml
app:
  telegram_token: "123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
```

### 2. 使用環境變數（推薦）

為了安全性，建議使用環境變數：

```bash
export TELEGRAM_BOT_TOKEN="123456789:ABCdefGHIjklMNOpqrsTUVwxyz"
```

然後在配置檔案中：

```yaml
app:
  telegram_token: "${TELEGRAM_BOT_TOKEN}"
```

## 步驟 3：設定 Chat ID

### 1. 獲取 Chat ID

#### 方法 1：使用 Bot API

```bash
curl "https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates"
```

#### 方法 2：使用 @userinfobot

1. 在 Telegram 中搜尋 `@userinfobot`
2. 點擊 "Start"
3. 將機器人加入您想要發送訊息的群組
4. 在群組中發送 `/start`
5. 查看返回的資訊，找到 `chat_id`

### 2. 更新 Chat ID 配置

編輯 `configs/config.{env}.yaml`：

```go
  chat_ids0: "-1002465088995" # information group (level 0)
  chat_ids1: "-1002465088995" # general message group (level 1)
  chat_ids2: "-1002465088995" # critical notification group (level 2)
  chat_ids3: "-1002465088995" # emergency alert group (level 3)
  chat_ids4: "-1002465088995" # testing group (level 4)
  chat_ids5: "-1002465088995" # backup group (level 5)
  chat_ids6: "" # backup group (level 5)
```

## 步驟 4：測試 Bot

### 1. 啟動應用程式

```bash
go run cmd/main.go -e development
```

### 2. 測試發送訊息

```bash
# 發送測試訊息到等級 0
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L0 \
  -H "Content-Type: application/json" \
  -d '{"message": "這是一條測試訊息"}'
```

### 3. 檢查機器人資訊

```bash
curl http://localhost:9999/api/v1/telegram/info
```

## 常見問題

### 1. "Not Found" 錯誤

**原因：** Bot Token 無效或格式錯誤

**解決方案：**

- 檢查 Token 是否正確複製
- 確認 Bot 是否已被刪除
- 重新從 BotFather 獲取 Token

### 2. "Forbidden" 錯誤

**原因：** Bot 沒有權限發送訊息到指定群組

**解決方案：**

- 確保 Bot 已加入群組
- 檢查群組權限設定
- 確認 Chat ID 是否正確

### 3. "Chat not found" 錯誤

**原因：** Chat ID 不正確

**解決方案：**

- 重新獲取正確的 Chat ID
- 確保 Bot 在群組中
- 檢查群組是否為公開群組

## 安全注意事項

1. **Token 安全**：

   - 不要將 Token 提交到版本控制系統
   - 使用環境變數或密鑰管理系統
   - 定期更換 Token

2. **群組權限**：

   - 限制 Bot 的權限
   - 只允許必要的操作
   - 定期審查群組成員

3. **訊息內容**：
   - 避免發送敏感資訊
   - 實施訊息內容過濾
   - 記錄所有發送的訊息

## 進階配置

### 1. 自定義訊息格式

您可以修改 `pkg/service/telegram.go` 中的 `SendMessage` 方法來支援更豐富的訊息格式：

```go
params := &bot.SendMessageParams{
    ChatID:    chatID,
    Text:      message,
    ParseMode: "Markdown", // 支援 Markdown 格式
}
```

### 2. 添加按鈕

```go
keyboard := &models.InlineKeyboardMarkup{
    InlineKeyboard: [][]models.InlineKeyboardButton{
        {
            {Text: "確認", CallbackData: "confirm"},
            {Text: "取消", CallbackData: "cancel"},
        },
    },
}

params := &bot.SendMessageParams{
    ChatID:      chatID,
    Text:        message,
    ReplyMarkup: keyboard,
}
```

### 3. 檔案發送

```go
fileParams := &bot.SendDocumentParams{
    ChatID:  chatID,
    Document: &models.InputFileUpload{
        Filename: "report.pdf",
        Data:     fileReader,
    },
    Caption: "附件報告",
}
```

## 監控和日誌

系統會記錄所有 Telegram 相關的操作：

- 訊息發送成功/失敗
- Bot 連接狀態
- 錯誤詳情

您可以在日誌中查看這些資訊來監控 Bot 的運行狀態。
