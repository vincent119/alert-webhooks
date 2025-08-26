# Discord 整合設定指南

此文件說明如何設定 Discord 機器人來接收 AlertManager 通知。

## 📋 目錄

- [創建 Discord 應用程式](#創建-discord-應用程式)
- [設定機器人權限](#設定機器人權限)
- [獲取必要資訊](#獲取必要資訊)
- [配置 Alert Webhooks](#配置-alert-webhooks)
- [測試設定](#測試設定)
- [故障排除](#故障排除)

## 🚀 創建 Discord 應用程式

### 步驟 1: 建立應用程式

1. 前往 [Discord Developer Portal](https://discord.com/developers/applications)
2. 點擊 **"New Application"**
3. 輸入應用程式名稱 (例如: "Alert Webhooks Bot")
4. 點擊 **"Create"**

### 步驟 2: 創建機器人

1. 在左側選單點擊 **"Bot"**
2. 點擊 **"Add Bot"**
3. 確認創建機器人

### 步驟 3: 配置機器人設定

1. 在 **"Bot"** 頁面中：
   - 設定機器人名稱和頭像
   - 複製 **Bot Token** (這是您需要的 `DISCORD_TOKEN`)
   - ⚠️ **重要**: 保持 Token 機密，不要公開分享

## 🔐 設定機器人權限

### 必要權限

機器人需要以下權限才能正常運作：

- ✅ **Send Messages** - 發送訊息
- ✅ **View Channels** - 查看頻道
- ✅ **Use External Emojis** - 使用外部表情符號
- ✅ **Read Message History** - 讀取訊息歷史

### 邀請機器人到伺服器

1. 在 **"OAuth2"** > **"URL Generator"** 中：
   - **Scopes**: 選擇 `bot`
   - **Bot Permissions**: 選擇上述必要權限
2. 複製生成的 URL 並在瀏覽器中開啟
3. 選擇您的 Discord 伺服器
4. 確認權限並授權

## 📝 獲取必要資訊

### 啟用開發者模式

1. 在 Discord 中，進入 **用戶設定** > **進階**
2. 啟用 **"開發者模式"**

### 獲取 Guild ID (伺服器 ID)

1. 右鍵點擊伺服器名稱
2. 選擇 **"複製 ID"**
3. 這就是您的 `guild_id`

### 獲取 Channel IDs (頻道 ID)

1. 右鍵點擊頻道名稱
2. 選擇 **"複製 ID"**
3. 重複此步驟獲取所有需要的頻道 ID

### 建議的伺服器結構

```
📁 您的 Discord 伺服器
├── 📝 alerts-info         (L0) - 資訊群組
├── 📢 alerts-general      (L1) - 一般訊息群組
├── 🚨 alerts-critical     (L2) - 重要通知群組
├── ⚠️  alerts-emergency   (L3) - 緊急警報群組
├── 🔧 alerts-testing      (L4) - 測試群組
└── 📦 alerts-backup       (L5) - 備用群組
```

## ⚙️ 配置 Alert Webhooks

### 配置文件設定

編輯您的 `config.yaml` 文件：

```yaml
discord:
  enable: true
  token: "${使用環境變數或是直接設定}" # 使用環境變數
  guild_id: "您的伺服器ID"
  username: "Alert Webhooks Bot"

  # 頻道對應 Alert Level
  channels:
    chat_ids0: "資訊群組頻道ID" # L0 - Information Group
    chat_ids1: "一般訊息頻道ID" # L1 - General Message Group
    chat_ids2: "重要通知頻道ID" # L2 - Critical Notification Group
    chat_ids3: "緊急警報頻道ID" # L3 - Emergency Alert Group
    chat_ids4: "測試群組頻道ID" # L4 - Testing Group
    chat_ids5: "備用群組頻道ID" # L5 - Backup Group

  # Discord 特定選項
  message_format: "markdown"
  mention_roles: [] # 可選: 需要 @mention 的角色 ID

  # 模板配置
  template_mode: "full" # minimal, full
  template_language: "tw" # eng, tw, zh, ja, ko
```

### 環境變數設定

設定 Discord Bot Token 環境變數：

```bash
export DISCORD_TOKEN="your-discord-bot-token-here"
```

### Kubernetes 設定

如果使用 Kubernetes，在 Secret 中設定：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: discord-secrets
type: Opaque
stringData:
  DISCORD_TOKEN: "your-discord-bot-token-here"
```

## 等級對應表

| Level | 頻道類型                 | 說明         | 範例用途               |
| ----- | ------------------------ | ------------ | ---------------------- |
| L0    | 📝 Information           | 資訊群組     | 一般資訊和狀態更新通知 |
| L1    | 📢 General Message       | 一般訊息群組 | 標準警報和日常監控通知 |
| L2    | 🚨 Critical Notification | 重要通知群組 | 重要警報和關鍵系統通知 |
| L3    | ⚠️ Emergency Alert       | 緊急警報群組 | 緊急事件和嚴重故障通知 |
| L4    | 🔧 Testing               | 測試群組     | 測試和開發環境通知     |
| L5    | 📦 Backup                | 備用群組     | 備用和容災通知群組     |

## 🧪 測試設定

### 測試機器人連接

```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin"
```

### 測試頻道驗證

```bash
curl -X POST "http://localhost:9999/api/v1/discord/validate/您的頻道ID" \
  -u "admin:admin"
```

### 發送測試訊息

```bash
curl -X POST "http://localhost:9999/api/v1/discord/test/您的頻道ID" \
  -u "admin:admin"
```

### 測試 Level 路由

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L0" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "測試 L0 訊息"}'
```

## 🔧 故障排除

### 常見錯誤

#### 1. "Missing Permissions" 錯誤

**原因**: 機器人缺少必要權限
**解決方案**:

- 確認機器人有 "Send Messages" 權限
- 檢查頻道特定權限設定
- 重新邀請機器人並確認權限

#### 2. "Unknown Channel" 錯誤

**原因**: 頻道 ID 不正確或機器人無法存取
**解決方案**:

- 檢查頻道 ID 是否正確
- 確認機器人已加入伺服器
- 檢查頻道是否為私人頻道

#### 3. "Unauthorized" 錯誤

**原因**: Discord Bot Token 無效
**解決方案**:

- 檢查 Token 是否正確
- 確認 Token 前有 "Bot " 前綴 (程式會自動添加)
- 重新生成 Bot Token

#### 4. "Bot is not in channel" 錯誤

**原因**: 機器人未加入特定頻道
**解決方案**:

- 確認機器人已加入伺服器
- 檢查頻道權限設定
- 嘗試手動 @mention 機器人

### 日誌檢查

查看 Discord 相關日誌：

```bash
grep "Discord" ./logs/server.log
```

### 驗證配置

檢查配置是否正確載入：

```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin" | jq .
```

## 📚 相關文件

- [Discord 使用指南](discord_usage.md)
- [Kubernetes 環境變數配置](kubernetes-env-vars.md)
- [服務啟用配置](service-enable-config.md)
- [模板系統說明](../en/template-system.md)

## 🆘 需要幫助？

如果遇到問題，請檢查：

1. Discord Bot Token 是否有效
2. 機器人是否有適當權限
3. 頻道 ID 是否正確
4. 網路連接是否正常
5. 應用程式日誌中的錯誤訊息
