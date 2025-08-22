# Slack 使用指南

本指南說明如何在 Alert-Webhooks 系統中使用 Slack 進行警報通知。

## API 端點概覽

Alert-Webhooks 提供以下 Slack API 端點：

| 端點                               | 方法 | 描述                      |
| ---------------------------------- | ---- | ------------------------- |
| `/api/v1/slack/channel/{channel}`  | POST | 發送訊息到指定頻道        |
| `/api/v1/slack/chatid_L{level}`    | POST | 發送訊息到指定等級頻道    |
| `/api/v1/slack/rich/{channel}`     | POST | 發送富文本訊息到指定頻道  |
| `/api/v1/slack/status`             | GET  | 獲取 Slack 服務狀態       |
| `/api/v1/slack/channels`           | GET  | 獲取已配置的頻道列表      |
| `/api/v1/slack/test`               | POST | 測試 Slack 連接           |
| `/api/v1/slack/validate/{channel}` | GET  | 驗證 Bot 是否在指定頻道中 |

## 認證

所有 API 端點都需要 HTTP Basic 認證：

- **用戶名**: `config.webhooks.base_auth_user`
- **密碼**: `config.webhooks.base_auth_password`

## 使用範例

### 1. 發送簡單訊息

#### 發送到指定頻道

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{"message": "系統維護通知：將於今晚 10 點進行維護"}' \
  "http://localhost:9999/api/v1/slack/channel/general"
```

#### 發送到等級頻道

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{"message": "緊急警報：資料庫連接失敗"}' \
  "http://localhost:9999/api/v1/slack/chatid_L0"
```

### 2. 發送 AlertManager 警報

#### 使用 AlertManager 數據格式

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{
    "receiver": "web.hook",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "instance": "server-01",
          "severity": "warning"
        },
        "annotations": {
          "summary": "CPU 使用率過高",
          "description": "server-01 的 CPU 使用率已超過 80%"
        },
        "startsAt": "2024-01-15T10:30:00Z",
        "generatorURL": "http://prometheus:9090/graph?g0.expr=cpu_usage"
      }
    ],
    "groupLabels": {"alertname": "HighCPUUsage"},
    "commonLabels": {"alertname": "HighCPUUsage", "severity": "warning"},
    "externalURL": "http://alertmanager:9093"
  }' \
  "http://localhost:9999/api/v1/slack/chatid_L1"
```

### 3. 發送富文本訊息

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{
    "blocks": [
      {
        "type": "header",
        "text": {
          "type": "plain_text",
          "text": "🚨 系統警報"
        }
      },
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*狀態*: 觸發中\n*嚴重程度*: 高\n*服務*: Web API"
        }
      },
      {
        "type": "actions",
        "elements": [
          {
            "type": "button",
            "text": {
              "type": "plain_text",
              "text": "查看詳情"
            },
            "url": "http://monitoring.example.com/alerts"
          }
        ]
      }
    ]
  }' \
  "http://localhost:9999/api/v1/slack/rich/alerts"
```

### 4. 服務狀態檢查

#### 獲取 Slack 服務狀態

```bash
curl -u admin:admin "http://localhost:9999/api/v1/slack/status"
```

回應範例：

```json
{
  "success": true,
  "service": "slack",
  "status": "active",
  "config": {
    "enabled": true,
    "default_channel": "#alerts",
    "bot_username": "Alert Bot"
  },
  "channels": {
    "L0": "#critical-alerts",
    "L1": "#warning-alerts",
    "L2": "#info-alerts"
  }
}
```

#### 獲取頻道配置

```bash
curl -u admin:admin "http://localhost:9999/api/v1/slack/channels"
```

#### 測試連接

```bash
curl -X POST \
  -u admin:admin \
  "http://localhost:9999/api/v1/slack/test"
```

#### 驗證頻道

```bash
curl -u admin:admin "http://localhost:9999/api/v1/slack/validate/alerts"
```

## 等級路由系統

### 等級配置

系統支援 6 個等級（L0-L5），每個等級對應不同的頻道：

| 等級      | 路由 | 建議用途                                    | 配置鍵                 |
| --------- | ---- | ------------------------------------------- | ---------------------- |
| chat_ids0 | L0   | Information Group（資訊群組）               | 一般資訊和狀態更新通知 |
| chat_ids1 | L1   | General Message Group（一般訊息群組）       | 標準警報和日常監控通知 |
| chat_ids2 | L2   | Critical Notification Group（重要通知群組） | 重要警報和關鍵系統通知 |
| chat_ids3 | L3   | Emergency Alert Group（緊急警報群組）       | 緊急事件和嚴重故障通知 |
| chat_ids4 | L4   | Testing Group（測試群組）                   | 測試和開發環境通知     |
| chat_ids5 | L5   | Backup Group（備用群組）                    | 備用和容災通知群組     |

### 配置範例

```yaml
slack:
  channels:
    chat_ids0: "#critical-alerts" # 緊急警報
    chat_ids1: "#warning-alerts" # 警告
    chat_ids2: "#info-alerts" # 資訊
    chat_ids3: "#debug-alerts" # 調試
    chat_ids4: "#test-alerts" # 測試
    chat_ids5: "#other-alerts" # 其他
```

## 訊息格式

### 模板系統

系統使用模板來格式化 AlertManager 警報訊息。模板位於：

- `templates/alerts/alert_template_tw.tmpl`（繁體中文）🇹🇼
- `templates/alerts/alert_template_eng.tmpl`（英文）🇺🇸
- `templates/alerts/alert_template_zh.tmpl`（簡體中文）🇨🇳
- `templates/alerts/alert_template_ja.tmpl`（日文）🇯🇵
- `templates/alerts/alert_template_ko.tmpl`（韓文）🇰🇷

### 模板配置

在配置檔案中設定：

```yaml
slack:
  template_mode: "full" # minimal 或 full
  template_language: "tw" # eng, tw, zh, ja, ko
```

### 格式化範例

#### Full 模式訊息範例：

```
🚨 警報通知

狀態: firing
警報名稱: HighCPUUsage
環境: production
嚴重程度: warning
命名空間: default
總警報數: 1
觸發中: 1

🚨 觸發中的警報:

警報 1:
• 摘要: CPU 使用率過高
• Pod: web-server-01
• 開始時間: 2024-01-15 10:30:00
• 查看詳情: http://prometheus:9090/graph?g0.expr=cpu_usage

查看所有警報詳情: http://alertmanager:9093
```

#### Minimal 模式訊息範例：

```
🚨 HighCPUUsage - warning
CPU 使用率過高 (web-server-01)
```

## 錯誤處理

### 常見錯誤和解決方案

#### 1. 認證失敗

```json
{
  "error": "Unauthorized",
  "message": "Invalid username or password",
  "code": 401
}
```

**解決方案**: 檢查 Basic Auth 用戶名和密碼

#### 2. 服務未啟用

```json
{
  "success": false,
  "message": "Slack service is not enabled"
}
```

**解決方案**: 在配置檔案中設定 `slack.enable: true`

#### 3. 頻道不存在

```json
{
  "success": false,
  "message": "Channel not found: #nonexistent"
}
```

**解決方案**: 確認頻道名稱正確，Bot 已被邀請到該頻道

#### 4. Token 無效

```json
{
  "success": false,
  "message": "Invalid token"
}
```

**解決方案**: 檢查 Slack Bot Token 是否正確設定

## 整合 AlertManager

### AlertManager 配置

在 AlertManager 的 `alertmanager.yml` 中配置 webhook：

```yaml
global:
  # 全域設定

route:
  group_by: ["alertname"]
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: "web.hook"
  routes:
    - match:
        severity: critical
      receiver: "slack-critical"
    - match:
        severity: warning
      receiver: "slack-warning"

receivers:
  - name: "web.hook"
    slack_configs:
      - send_resolved: true
        api_url: "http://alert-webhooks:9999/api/v1/slack/chatid_L2"
        username: "Alert Bot"
        channel: "#info-alerts"

  - name: "slack-critical"
    slack_configs:
      - send_resolved: true
        api_url: "http://alert-webhooks:9999/api/v1/slack/chatid_L0"
        username: "Alert Bot"
        channel: "#critical-alerts"

  - name: "slack-warning"
    slack_configs:
      - send_resolved: true
        api_url: "http://alert-webhooks:9999/api/v1/slack/chatid_L1"
        username: "Alert Bot"
        channel: "#warning-alerts"
```

### Prometheus 規則範例

```yaml
groups:
  - name: example
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "CPU 使用率過高"
          description: "{{ $labels.instance }} 的 CPU 使用率已超過 80%"

      - alert: ServiceDown
        expr: up == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "服務停機"
          description: "{{ $labels.instance }} 服務已停機超過 5 分鐘"
```

## 效能和限制

### 訊息限制

- Slack 訊息長度限制：40,000 字符
- 富文本區塊限制：50 個區塊
- 附件限制：20 個附件

### 速率限制

- Slack API 有速率限制，建議：
  - 每秒不超過 1 個請求
  - 使用批次處理大量警報
  - 避免重複發送相同訊息

### 最佳實踐

1. 使用適當的等級路由避免訊息氾濫
2. 設定合理的 AlertManager 分組規則
3. 使用模板自訂來減少訊息冗餘
4. 定期檢查和清理不活躍的頻道配置

## 相關文檔

- [Slack 設定指南](./slack_setup.md)
- [服務啟用配置指南](./service-enable-config.md)
- [模板使用指南](./template_usage.md)
- [Kubernetes 環境變數配置](./kubernetes-env-vars.md)
