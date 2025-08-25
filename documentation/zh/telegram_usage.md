# Telegram API 使用範例

## 概述

本專案整合了 [go-telegram/bot](https://github.com/go-telegram/bot) 框架來發送 Telegram 訊息。

## 配置

### 1. 設定 Telegram Bot Token

在配置檔案中設定您的 Telegram Bot Token：

```yaml
# configs/config.development.yaml
app:
  telegram_token: "YOUR_TELEGRAM_BOT_TOKEN_HERE"
```

### 2. 設定 Chat ID

在 `configs/config.{env}.yaml` 中設定對應的 chat_id：

```go
  chat_ids0: "-1002465088995" # information group (level 0)
  chat_ids1: "-1002465088995" # general message group (level 1)
  chat_ids2: "-1002465088995" # critical notification group (level 2)
  chat_ids3: "-1002465088995" # emergency alert group (level 3)
  chat_ids4: "-1002465088995" # testing group (level 4)
  chat_ids5: "-1002465088995" # backup group (level 5)
  chat_ids6: "" # backup group (level 5)
```

## API 端點

### 1. 發送訊息

**端點：** `POST /api/v1/chatid_L{level}`

**參數：**

- `level`: 聊天等級 (0-4)

**請求範例：**

```bash
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L0 \
  -H "Content-Type: application/json" \
  -d '{
    "message": "這是一條測試訊息"
  }'
```

**回應範例：**

```json
{
  "success": true,
  "message": "Message sent successfully",
  "level": 0
}
```

### 2. 獲取機器人資訊

**端點：** `GET /api/v1/telegram/info`

**請求範例：**

```bash
curl http://localhost:9999/api/v1/telegram/info
```

**回應範例：**

```json
{
  "success": true,
  "bot_info": {
    "id": 123456789,
    "username": "your_bot_username",
    "first_name": "Your Bot Name",
    "can_join_groups": true,
    "can_read_all_group_messages": false
  }
}
```

## 使用範例

### 1. 發送不同等級的訊息

```bash
# 發送到等級 0
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "telegram-alerts",
    "status": "firing",
    "groupLabels": {
      "alertname": "HighCPUUsage",
      "env": "production",
      "severity": "warning"
    },
    "commonAnnotations": {
      "summary": "High CPU usage detected"
    },
    "externalURL": "https://alertmanager.example.com",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "env": "production",
          "severity": "warning",
          "instance": "server-01"
        },
        "annotations": {
          "summary": "CPU usage above 80% for 5 minutes",
          "description": "Server experiencing high load"
        },
        "startsAt": "2023-01-01T10:30:00.000Z",
        "generatorURL": "https://prometheus.example.com/graph?expr=cpu_usage"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L0

# 發送到等級 1
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "telegram-alerts",
    "status": "firing",
    "groupLabels": {
      "alertname": "HighCPUUsage",
      "env": "production",
      "severity": "warning"
    },
    "commonAnnotations": {
      "summary": "High CPU usage detected"
    },
    "externalURL": "https://alertmanager.example.com",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "env": "production",
          "severity": "warning",
          "instance": "server-01"
        },
        "annotations": {
          "summary": "CPU usage above 80% for 5 minutes",
          "description": "Server experiencing high load"
        },
        "startsAt": "2023-01-01T10:30:00.000Z",
        "generatorURL": "https://prometheus.example.com/graph?expr=cpu_usage"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L1
```

### 2. 錯誤處理

**無效的 chatid 格式：**

```bash
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L5 \
  -H "Content-Type: application/json" \
  -d '{"message": "測試"}'
```

**回應：**

```json
{
  "success": false,
  "message": "Invalid chatid format. Must be L0, L1, L2, L3, or L4"
}
```

**空訊息：**

```bash
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L0 \
  -H "Content-Type: application/json" \
  -d '{"message": ""}'
```

**回應：**

```json
{
  "success": false,
  "message": "Message cannot be empty"
}
```

## 日誌記錄

系統會記錄所有 Telegram 相關的操作：

### 成功發送訊息

```
{"level":"info","ts":"2024-01-15T10:30:45.123+08:00","msg":"Telegram message sent successfully","category":"telegram","level":0,"chat_id":-1001234567890,"message":"測試訊息"}
```

### 發送失敗

```
{"level":"error","ts":"2024-01-15T10:30:45.123+08:00","msg":"Failed to send Telegram message","category":"telegram","level":0,"chat_id":-1001234567890,"message":"測試訊息","error":"Forbidden: bot was blocked by the user"}
```

## 注意事項

1. **Token 安全**：確保您的 Telegram Bot Token 不會暴露在公開的程式碼中
2. **Chat ID**：確保設定的 chat_id 是正確的，並且機器人有權限發送訊息
3. **訊息格式**：支援純文字訊息，如需更豐富的格式可以使用 Markdown 或 HTML
4. **速率限制**：Telegram API 有速率限制，請注意不要過於頻繁地發送訊息

## 擴展功能

您可以根據需求擴展以下功能：

1. **訊息格式**：支援 Markdown、HTML 格式
2. **檔案發送**：支援發送圖片、文件等
3. **按鈕**：支援內聯鍵盤按鈕
4. **群組管理**：支援群組管理功能
5. **訊息編輯**：支援編輯已發送的訊息
