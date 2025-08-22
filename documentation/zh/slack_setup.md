# Slack 設定指南

本指南將引導您完成 Slack Bot 的設定和配置過程。

## 前置需求

- Slack 工作區管理員權限
- 能夠創建和安裝 Slack 應用程式

## 步驟 1: 創建 Slack 應用程式

### 1.1 前往 Slack API 控制台

1. 打開瀏覽器，前往 [Slack API](https://api.slack.com/apps)
2. 點擊 **"Create New App"**
3. 選擇 **"From scratch"**

### 1.2 填寫應用程式資訊

1. **App Name**: 輸入應用程式名稱（例如：`Alert Bot`）
2. **Pick a workspace**: 選擇您的工作區
3. 點擊 **"Create App"**

## 步驟 2: 配置 Bot 權限

### 2.1 設定 OAuth & Permissions

1. 在左側選單中，點擊 **"OAuth & Permissions"**
2. 滾動到 **"Scopes"** 區段
3. 在 **"Bot Token Scopes"** 中添加以下權限：

#### 必需權限：

```
chat:write          # 發送訊息
chat:write.public    # 發送訊息到公開頻道
channels:read        # 讀取頻道資訊
groups:read          # 讀取私人頻道資訊
im:read              # 讀取直接訊息
mpim:read            # 讀取群組訊息
```

#### 可選權限（建議添加）：

```
users:read           # 讀取用戶資訊（用於 @mentions）
channels:join        # 自動加入頻道
```

## 步驟 3: 安裝應用程式

### 3.1 安裝到工作區

1. 滾動到頁面頂部的 **"OAuth Tokens for Your Workspace"** 區段
2. 點擊 **"Install to Workspace"**
3. 審查權限並點擊 **"Allow"**

### 3.2 獲取 Bot Token

安裝完成後，您將看到 **"Bot User OAuth Token"**：

```
xoxb-xxxxxxxxxxxxx-xxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
```

**重要**: 請妥善保存這個 Token，它將用於應用程式配置。

## 步驟 4: 配置頻道

### 4.1 邀請 Bot 到頻道

對於每個要接收警報的頻道：

1. 前往該頻道
2. 輸入：`/invite @your-bot-name`
3. 或者在頻道資訊中點擊 **"Add apps"** 並選擇您的 Bot

### 4.2 獲取頻道 ID（可選）

如果需要使用頻道 ID 而不是頻道名稱：

1. 右鍵點擊頻道名稱
2. 選擇 **"Copy link"**
3. 頻道 ID 是 URL 中最後的部分：`https://yourworkspace.slack.com/archives/C1234567890`
4. 頻道 ID 格式：`C1234567890`

## 步驟 5: 應用程式配置

### 5.1 配置檔案設定

在 `config.yaml` 中添加 Slack 配置：

```yaml
slack:
  # 啟用 Slack 服務
  enable: true

  # Bot Token（也可通過環境變數 SLACK_TOKEN 設定）
  token: "xoxb-your-slack-bot-token"

  # 預設頻道（備用頻道）
  channel: "#alerts"

  # Bot 顯示設定
  username: "Alert Bot"
  icon_emoji: ":warning:" # 或使用 icon_url
  # icon_url: "https://example.com/bot-icon.png"

  # 多頻道配置（依警報等級分配）
  channels:
    chat_ids0: "資訊群組頻道ID" # L0 - Information Group
    chat_ids1: "一般訊息頻道ID" # L1 - General Message Group
    chat_ids2: "重要通知頻道ID" # L2 - Critical Notification Group
    chat_ids3: "緊急警報頻道ID" # L3 - Emergency Alert Group
    chat_ids4: "測試群組頻道ID" # L4 - Testing Group
    chat_ids5: "備用群組頻道ID" # L5 - Backup Group


  # 訊息選項
  link_names: true # 啟用 @mentions
  unfurl_links: false # 不展開連結預覽
  unfurl_media: false # 不展開媒體預覽

  # 模板設定
  template_mode: "full" # minimal 或 full
  template_language: "tw" # eng, tw, zh, ja, ko
```

### 5.2 環境變數設定（推薦用於生產環境）

```bash
# Slack Bot Token
export SLACK_TOKEN="xoxb-your-slack-bot-token"
```

在 Kubernetes 中：

```yaml
env:
  - name: SLACK_TOKEN
    valueFrom:
      secretKeyRef:
        name: alert-webhooks-secrets
        key: slack-token
```

## 步驟 6: 測試配置

### 6.1 啟動應用程式

確保配置正確後，啟動應用程式：

```bash
go run cmd/main.go
```

查看啟動日誌，確認 Slack 服務已啟用：

```
Service enable status - Webhooks: true, Telegram: false, Slack: true
```

### 6.2 測試 API 端點

#### 檢查服務狀態：

```bash
curl -u admin:admin http://localhost:9999/api/v1/slack/status
```

#### 測試發送訊息：

```bash
curl -X POST -H "Content-Type: application/json" -u admin:admin \
  -d '{"message": "測試訊息"}' \
  "http://localhost:9999/api/v1/slack/channel/alerts"
```

#### 測試等級路由：

```bash
curl -X POST -H "Content-Type: application/json" -u admin:admin \
  -d '{"message": "緊急警報測試"}' \
  "http://localhost:9999/api/v1/slack/chatid_L0"
```

## 頻道配置說明

### 頻道名稱格式

支援以下格式：

- **公開頻道**: `#channel-name`
- **私人頻道**: `#private-channel`（Bot 必須被邀請）
- **頻道 ID**: `C1234567890`

### 等級對應

| chat_ids  | Level | 群組用途 (Group Purpose)                    |
| --------- | ----- | ------------------------------------------- |
| chat_ids0 | 0     | Information group（資訊群組）               |
| chat_ids1 | 1     | General message group（一般訊息群組）       |
| chat_ids2 | 2     | Critical notification group（重要通知群組） |
| chat_ids3 | 3     | Emergency alert group（緊急警報群組）       |
| chat_ids4 | 4     | Testing group（測試群組）                   |
| chat_ids5 | 5     | Backup group（備用群組）                    |

## 常見問題解決

### 問題 1: Bot 無法發送訊息

**錯誤**: `not_in_channel`

**解決方案**:

1. 確保 Bot 已被邀請到目標頻道
2. 執行：`/invite @your-bot-name` 在該頻道中

### 問題 2: 權限不足

**錯誤**: `missing_scope`

**解決方案**:

1. 返回 Slack API 控制台
2. 檢查並添加必要的 Bot Token Scopes
3. 重新安裝應用程式到工作區

### 問題 3: Token 無效

**錯誤**: `invalid_auth`

**解決方案**:

1. 檢查 Token 格式是否正確（應以 `xoxb-` 開頭）
2. 確認 Token 沒有過期
3. 重新生成 Token

### 問題 4: 頻道不存在

**錯誤**: `channel_not_found`

**解決方案**:

1. 確認頻道名稱拼寫正確
2. 確認頻道存在且 Bot 有訪問權限
3. 使用頻道 ID 替代頻道名稱

## 進階配置

### 富文本訊息

使用富文本 API 發送格式化訊息：

```bash
curl -X POST -H "Content-Type: application/json" -u admin:admin \
  -d '{
    "blocks": [
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*警報通知*\n狀態: 觸發中\n嚴重程度: 高"
        }
      }
    ]
  }' \
  "http://localhost:9999/api/v1/slack/rich/alerts"
```

### 模板自訂

可以修改以下模板檔案來自訂 Slack 訊息格式：

- `templates/alerts/alert_template_tw.tmpl`（繁體中文）🇹🇼
- `templates/alerts/alert_template_eng.tmpl`（英文）🇺🇸
- `templates/alerts/alert_template_zh.tmpl`（簡體中文）🇨🇳
- `templates/alerts/alert_template_ja.tmpl`（日文）🇯🇵
- `templates/alerts/alert_template_ko.tmpl`（韓文）🇰🇷

### 多工作區支援

如需支援多個 Slack 工作區，可以：

1. 為每個工作區創建不同的配置檔案
2. 使用不同的環境變數設定不同的 Token

## 相關文檔

- [服務啟用配置指南](./service-enable-config.md)
- [Kubernetes 環境變數配置](./kubernetes-env-vars.md)
- [模板使用指南](./template_usage.md)
