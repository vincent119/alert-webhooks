# Discord 使用指南 🇹🇼

此文件說明如何使用 Alert Webhooks 的 Discord API 端點發送通知。

## 📋 目錄

- [API 端點](#api-端點)
- [使用範例](#使用範例)
- [Level 路由](#level-路由)
- [模板系統](#模板系統)
- [錯誤處理](#錯誤處理)
- [AlertManager 整合](#alertmanager-整合)

## 🔗 API 端點

### 基本端點格式

```
POST /api/v1/discord/channel/{channel_id}    - 發送到指定頻道
POST /api/v1/discord/chatid_L{level}         - 發送到指定等級
GET  /api/v1/discord/status                  - 獲取服務狀態
POST /api/v1/discord/test/{channel_id}       - 測試頻道連接
POST /api/v1/discord/validate/{channel_id}   - 驗證頻道權限
```

### 認證方式

所有 API 端點都需要基本認證 (Basic Auth)：

```bash
-u "username:password"
```

## 💡 使用範例

### 1. 簡單文字訊息

#### 發送到指定頻道

```bash
curl -X POST "http://localhost:9999/api/v1/discord/channel/987654321098765432" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{
    "message": "🚨 系統警報：資料庫連接異常"
  }'
```

#### 發送到 L0 (緊急警報)

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L0" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{
    "message": "🔥 緊急：生產環境服務中斷"
  }'
```

### 2. AlertManager 格式訊息

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L1" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "instance": "web-server-01",
          "severity": "warning",
          "env": "production"
        },
        "annotations": {
          "summary": "CPU 使用率過高",
          "description": "CPU 使用率已超過 80%"
        },
        "startsAt": "2024-01-15T10:30:00Z"
      }
    ],
    "status": "firing",
    "externalURL": "http://alertmanager.example.com"
  }'
```

### 3. 服務狀態檢查

```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin"
```

**回應範例:**

```json
{
  "service": "discord",
  "status": "healthy",
  "bot_info": {
    "id": "123456789012345678",
    "username": "Alert Webhooks Bot",
    "bot": true,
    "avatar": "abc123def456"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 4. 頻道驗證

```bash
curl -X POST "http://localhost:9999/api/v1/discord/validate/987654321098765432" \
  -u "admin:admin"
```

### 5. 測試訊息

```bash
curl -X POST "http://localhost:9999/api/v1/discord/test/987654321098765432" \
  -u "admin:admin"
```

## 📊 Level 路由

### 等級對應表

| Level | API 端點     | 頻道類型                 | 配置鍵      | 說明         |
| ----- | ------------ | ------------------------ | ----------- | ------------ |
| L0    | `/chatid_L0` | 📝 Information           | `chat_ids0` | 資訊群組     |
| L1    | `/chatid_L1` | 📢 General Message       | `chat_ids1` | 一般訊息群組 |
| L2    | `/chatid_L2` | 🚨 Critical Notification | `chat_ids2` | 重要通知群組 |
| L3    | `/chatid_L3` | ⚠️ Emergency Alert       | `chat_ids3` | 緊急警報群組 |
| L4    | `/chatid_L4` | 🔧 Testing               | `chat_ids4` | 測試群組     |
| L5    | `/chatid_L5` | 📦 Backup                | `chat_ids5` | 備用群組     |

### 使用建議

#### 📝 L0 - Information Group (資訊群組)

- 一般資訊通知
- 狀態更新
- 系統維護通知
- 非緊急訊息

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L0" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "📝 **資訊通知**: 系統維護將於今晚進行"}'
```

#### 📢 L1 - General Message Group (一般訊息群組)

- 標準警報
- 日常監控通知
- 一般性問題
- 常規系統事件

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L1" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "📢 **一般警報**: CPU 使用率偏高"}'
```

#### 🚨 L2 - Critical Notification Group (重要通知群組)

- 重要警報
- 關鍵系統通知
- 需要關注的問題
- 服務異常警告

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L2" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "🚨 **重要警報**: 資料庫連接異常"}'
```

#### ⚠️ L3 - Emergency Alert Group (緊急警報群組)

- 緊急事件
- 嚴重故障通知
- 服務完全中斷
- 需要立即處理的問題

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L3" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "⚠️ **緊急警報**: 服務完全中斷"}'
```

#### 🔧 L4 - Testing Group (測試群組)

- 測試環境通知
- 開發環境警報
- 測試結果
- 部署通知

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L4" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "🔧 **測試通知**: 開發環境部署完成"}'
```

#### 📦 L5 - Backup Group (備用群組)

- 備用通知
- 容災相關訊息
- 其他雜項通知
- 備份系統通知

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L5" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "📦 **備用通知**: 備份任務完成"}'
```

## 🎨 模板系統

### 模板配置

Discord 使用與 Telegram/Slack 相同的模板系統：

```yaml
discord:
  template_mode: "full" # minimal, full
  template_language: "tw" # eng, tw, zh, ja, ko
```

### 支援的格式化

Discord 支援標準 Markdown 格式：

- **粗體文字**: `**文字**`
- _斜體文字_: `*文字*`
- `行內代碼`: `` `代碼` ``
- 代碼塊: ` `代碼塊` `
- [連結](URL): `[連結文字](URL)`

### 模板變數

模板可以使用以下變數：

- `{{.alerts}}` - 警報陣列
- `{{.status}}` - 警報狀態
- `{{.externalURL}}` - 外部連結
- `{{.alertname}}` - 警報名稱
- `{{.env}}` - 環境
- `{{.severity}}` - 嚴重性
- `{{.namespace}}` - 命名空間

## 🚨 錯誤處理

### 常見錯誤回應

#### 1. 權限不足

```json
{
  "success": false,
  "message": "bot lacks necessary permissions in channel 987654321098765432. Please ensure the bot has 'Send Messages' permission"
}
```

#### 2. 頻道不存在

```json
{
  "success": false,
  "message": "channel 987654321098765432 does not exist or bot cannot access it"
}
```

#### 3. Token 無效

```json
{
  "success": false,
  "message": "invalid Discord token. Please check if the token in configuration is correct"
}
```

#### 4. 訊息內容無效

```json
{
  "success": false,
  "message": "message content is invalid or too long"
}
```

### 訊息長度限制

- Discord 訊息最大長度：**2000 字元**
- 超過限制時，系統會自動分割為多個訊息
- 分割會在換行符處進行，保持格式完整

## 🔗 AlertManager 整合

### Webhook 配置

在 AlertManager 的 `alertmanager.yml` 中配置：

```yaml
route:
  group_by: ["alertname"]
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: "discord-notifications"
  routes:
    - match:
        severity: critical
      receiver: "discord-critical"
    - match:
        severity: warning
      receiver: "discord-warning"

receivers:
  - name: "discord-critical"
    webhook_configs:
      - url: "http://alert-webhooks:9999/api/v1/discord/chatid_L0"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"

  - name: "discord-warning"
    webhook_configs:
      - url: "http://alert-webhooks:9999/api/v1/discord/chatid_L1"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"

  - name: "discord-notifications"
    webhook_configs:
      - url: "http://alert-webhooks:9999/api/v1/discord/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
```

### Prometheus Rules 範例

```yaml
groups:
  - name: discord.rules
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 80
        for: 5m
        labels:
          severity: warning
          env: production
        annotations:
          summary: "CPU 使用率過高"
          description: "實例 {{ $labels.instance }} CPU 使用率已達 {{ $value }}%"

      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
          env: production
        annotations:
          summary: "服務已停止"
          description: "服務 {{ $labels.job }} 在實例 {{ $labels.instance }} 上已停止"
```

## 📊 監控和日誌

### 檢查 Discord 服務狀態

```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin" | jq .
```

### 查看 Discord 相關日誌

```bash
grep "Discord" ./logs/server.log | tail -20
```

### 測試端到端流程

1. 檢查服務狀態
2. 驗證頻道權限
3. 發送測試訊息
4. 檢查 Discord 頻道

```bash
# 完整測試流程
curl -X GET "http://localhost:9999/api/v1/discord/status" -u "admin:admin"
curl -X POST "http://localhost:9999/api/v1/discord/validate/your-channel-id" -u "admin:admin"
curl -X POST "http://localhost:9999/api/v1/discord/test/your-channel-id" -u "admin:admin"
```

## 🔧 進階配置

### 提及角色 (Mention Roles)

可以在訊息中自動提及特定角色：

```yaml
discord:
  mention_roles:
    - "role-id-for-ops-team"
    - "role-id-for-on-call"
```

### 自訂訊息格式

可以透過模板系統自訂訊息格式，支援：

- Markdown 格式化
- 表情符號
- 自訂文字和佈局
- 多語言支援

## 📚 相關文件

- [Discord 設定指南](discord_setup.md)
- [模板系統說明](../en/template-system.md)
- [Kubernetes 環境變數](kubernetes-env-vars.md)
- [故障排除指南](../en/troubleshooting.md)

## 💡 最佳實踐

1. **使用適當的 Level** - 根據警報嚴重性選擇合適的等級
2. **設定角色提及** - 為緊急警報配置角色提及
3. **測試配置** - 定期測試 Discord 整合
4. **監控日誌** - 定期檢查服務日誌
5. **備用頻道** - 配置備用頻道以防主頻道問題
