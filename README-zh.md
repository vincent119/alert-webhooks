# Alert Webhooks

一個高效的 AlertManager webhook 處理服務，支援多平台通知和多語言模板。

[![GitHub](https://img.shields.io/badge/github-vincent119%2Falert--webhooks-blue?logo=github)](https://github.com/vincent119/alert-webhooks)
![License](https://img.shields.io/github/license/awslabs/mcp)
[![Go Version](https://img.shields.io/badge/go-1.19%2B-blue?logo=go)](go.mod)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/vincent119/alert-webhooks/actions)
[![Coverage](https://img.shields.io/badge/coverage-89%25-yellow)](https://codecov.io/gh/vincent119/alert-webhooks)
[![Score](https://img.shields.io/badge/score-9.2-brightgreen)](https://goreportcard.com/report/github.com/vincent119/alert-webhooks)
[![GitHub stars](https://img.shields.io/github/stars/vincent119/alert-webhooks?style=flat&color=yellow&logo=github&logoColor=white)](https://github.com/vincent119/alert-webhooks/stargazers)
![GitHub release](https://img.shields.io/github/v/release/awslabs/mcp)

## 📋 目錄

- [🌟 主要功能](#-主要功能)
- [🚀 支援平台](#-支援平台)
- [⚡ 快速開始](#-快速開始)
- [📊 Chat IDs 群組用途對照表](#-chat-ids-群組用途對照表)
- [📖 詳細文檔](#-詳細文檔)
- [🛠️ 開發](#️-開發)
- [📄 AlertManager Webhook 樣本](#-alertmanager-webhook-樣本)
- [📁 項目結構](#-項目結構)
- [🌍 語言版本](#-語言版本)

## 🌟 主要功能

- 🔗 **AlertManager 整合**: 直接接收和處理 AlertManager webhook
- 📱 **多平台通知**: 支援 Telegram、Slack、Discord 三大通訊平台
- 🎯 **多等級通知**: 支援不同等級的群組通知分發
- 🌍 **多語言模板**: 英語、繁體中文、簡體中文、日語、韓語
- 🔄 **熱重載**: 配置文件和模板動態重載，無需重啟服務
- 🔐 **安全認證**: HTTP Basic Auth 保護
- 📋 **雙模板模式**: Full/Minimal 兩種顯示格式
- 📨 **分離通知**: 觸發中和已解決警報分別發送
- 🎨 **自定義模板**: 支援自定義消息模板和格式

## 🚀 支援平台

本服務支援以下通訊平台的警報通知：

### 📱 Telegram

- ✅ 支援多個聊天群組
- ✅ 支援不同等級的通知分發
- ✅ 支援機器人資訊查詢
- ✅ 支援自定義消息格式

### 💬 Slack

- ✅ 支援 Webhook 通知
- ✅ 支援頻道消息發送
- ✅ 支援自定義消息格式
- ✅ 支援附件和格式化消息

### 🎮 Discord

- ✅ 支援伺服器頻道通知
- ✅ 支援 Webhook 消息
- ✅ 支援豐富的消息格式
- ✅ 支援嵌入式消息

## ⚡ 快速開始

### 1. 安裝依賴

```bash
go mod download
```

### 2. 配置設置

```bash
# 複製配置文件
cp examples/config.expamle configs/config.development.yaml

# 編輯配置（設置 Telegram token 和 chat IDs）
vim configs/config.development.yaml
```

### 3. 啟動服務

```bash
# 開發環境
make dev
# 或
go run cmd/main.go -e development

# 生產環境
make run
# 或
go run cmd/main.go -e production
```

### 4. 訪問 API 文檔

打開瀏覽器訪問: http://localhost:9999/swagger/index.html

### 📊 Chat IDs 群組用途對照表

| Chat IDs  | Level | 群組用途                                    | 說明                   |
| --------- | ----- | ------------------------------------------- | ---------------------- |
| chat_ids0 | L0    | Information Group（資訊群組）               | 一般資訊和狀態更新通知 |
| chat_ids1 | L1    | General Message Group（一般訊息群組）       | 標準警報和日常監控通知 |
| chat_ids2 | L2    | Critical Notification Group（重要通知群組） | 重要警報和關鍵系統通知 |
| chat_ids3 | L3    | Emergency Alert Group（緊急警報群組）       | 緊急事件和嚴重故障通知 |
| chat_ids4 | L4    | Testing Group（測試群組）                   | 測試和開發環境通知     |
| chat_ids5 | L5    | Backup Group（備用群組）                    | 備用和容災通知群組     |

## 📖 詳細文檔

完整的設置和使用指南提供多語言版本，請查看 **[documentation](./documentation/)** 目錄：

### 🌍 語言選擇

- **[English Documentation](./documentation/en/)** - 完整的英文文檔
- **[繁體中文文檔](./documentation/zh/)** - 完整的繁體中文文檔

### 📋 快速連結

#### 🔧 基礎配置

- **[配置指南](./documentation/zh/config_guide.md)** - 詳細配置說明
- **[服務啟用配置](./documentation/zh/service-enable-config.md)** - 服務啟用設定
- **[Kubernetes 環境變數](./documentation/zh/kubernetes-env-vars.md)** - K8s 部署配置
- **[Swagger 疑難排解](./documentation/zh/swagger-troubleshooting.md)** - API 文檔問題解決

#### 📝 模板系統

- **[模板指南](./documentation/zh/template_guide.md)** - 自定義模板說明
- **[模板模式配置](./documentation/zh/template_mode_config.md)** - Full/Minimal 模式設定
- **[模板使用範例](./documentation/zh/template_usage.md)** - 模板實際應用範例

#### 📱 平台設置指南

- **[Telegram 設置](./documentation/zh/telegram_setup.md)** - Telegram 機器人配置
- **[Slack 設置](./documentation/zh/slack_setup.md)** - Slack 應用程式配置
- **[Discord 設置](./documentation/zh/discord_setup.md)** - Discord 機器人配置

#### 📚 平台使用範例

- **[Telegram 使用範例](./documentation/zh/telegram_usage.md)** - Telegram API 使用範例
- **[Slack 使用範例](./documentation/zh/slack_usage.md)** - Slack API 使用範例
- **[Discord 使用範例](./documentation/zh/discord_usage.md)** - Discord API 使用範例

## 🛠️ 開發

### Makefile 命令

```bash
make dev              # 啟動開發環境
make build            # 編譯項目
make test             # 運行測試
make swagger-generate # 重新生成 Swagger 文檔
make fmt              # 格式化代碼
make lint             # 代碼質量檢查
```

### API 端點

#### 📱 Telegram API

| 方法   | 路徑                              | 描述               | 認證          |
| ------ | --------------------------------- | ------------------ | ------------- |
| `POST` | `/api/v1/telegram/chatid_{level}` | 發送 Telegram 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/telegram/info`           | 獲取機器人資訊     | ✅ Basic Auth |

#### 💬 Slack API

| 方法   | 路徑                           | 描述            | 認證          |
| ------ | ------------------------------ | --------------- | ------------- |
| `POST` | `/api/v1/slack/chatid_{level}` | 發送 Slack 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/slack/info`           | 獲取 Slack 資訊 | ✅ Basic Auth |

#### 🎮 Discord API

| 方法   | 路徑                             | 描述              | 認證          |
| ------ | -------------------------------- | ----------------- | ------------- |
| `POST` | `/api/v1/discord/chatid_{level}` | 發送 Discord 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/discord/info`           | 獲取 Discord 資訊 | ✅ Basic Auth |

#### 🔧 系統 API

| 方法  | 路徑              | 描述     | 認證          |
| ----- | ----------------- | -------- | ------------- |
| `GET` | `/api/v1/healthz` | 健康檢查 | ❌            |
| `GET` | `/swagger/*`      | API 文檔 | ✅ Basic Auth |

### AlertManager 整合範例

#### 📱 Telegram 通知設定

```yaml
# alertmanager.yml - Telegram 設定
route:
  receiver: "telegram-notifications"

receivers:
  - name: "telegram-notifications"
    webhook_configs:
      - url: "http://localhost:9999/api/v1/telegram/chatid_L0"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
```

#### 💬 Slack 通知設定

```yaml
# alertmanager.yml - Slack 設定
route:
  receiver: "slack-notifications"

receivers:
  - name: "slack-notifications"
    webhook_configs:
      - url: "http://localhost:9999/api/v1/slack/chatid_L1"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
```

#### 🎮 Discord 通知設定

```yaml
# alertmanager.yml - Discord 設定
route:
  receiver: "discord-notifications"

receivers:
  - name: "discord-notifications"
    webhook_configs:
      - url: "http://localhost:9999/api/v1/discord/chatid_L1"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
```

#### 🔄 多平台同時通知

```yaml
# alertmanager.yml - 多平台設定
route:
  receiver: "multi-platform-notifications"

receivers:
  - name: "multi-platform-notifications"
    webhook_configs:
      # Telegram 通知
      - url: "http://localhost:9999/api/v1/telegram/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
      # Slack 通知
      - url: "http://localhost:9999/api/v1/slack/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
      # Discord 通知
      - url: "http://localhost:9999/api/v1/discord/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
```

### 📄 AlertManager Webhook 樣本

項目根目錄中的 `raw_alertmanager.json` 文件提供了完整的 Prometheus AlertManager webhook 負載樣本，包含：

#### 🔍 樣本內容說明

```json
{
  "receiver": "test-telegram-webhook", // 接收器名稱
  "status": "firing", // 警報狀態: firing/resolved
  "alerts": [
    // 警報陣列
    {
      "status": "firing", // 個別警報狀態
      "labels": {
        // 警報標籤
        "alertname": "TEST_Pod_CPU_Usage High", // 警報名稱
        "env": "uat", // 環境標籤
        "namespace": "hcgateconsole", // Kubernetes 命名空間
        "pod": "hcgateconsole-deploy-xxx", // Pod 名稱
        "severity": "test-alert" // 嚴重性等級
      },
      "annotations": {
        // 警報註釋
        "summary": "Pod CPU 使用率超過 80%" // 警報摘要
      },
      "startsAt": "2022-11-18T08:17:31.745Z", // 警報開始時間
      "endsAt": "0001-01-01T00:00:00Z", // 警報結束時間 (firing 時為空)
      "generatorURL": "http://prometheus...", // Prometheus 查詢連結
      "fingerprint": "2da0690c63cf9cd3" // 警報指紋識別碼
    }
  ],
  "groupLabels": {
    // 群組標籤
    "alertname": "TEST_Pod_CPU_Usage High",
    "env": "uat",
    "severity": "test-alert"
  },
  "commonLabels": {
    // 共同標籤
    "alertname": "TEST_Pod_CPU_Usage High",
    "env": "uat",
    "severity": "test-alert"
  },
  "commonAnnotations": {}, // 共同註釋
  "externalURL": "http://prometheus-alertmanager:9093", // AlertManager 外部 URL
  "version": "4", // AlertManager 版本
  "groupKey": "...", // 群組識別鍵
  "truncatedAlerts": 0 // 截斷的警報數量
}
```

#### 🎯 使用場景

- **開發測試**: 用於測試 webhook 端點的功能
- **模板開發**: 開發自定義警報模板時的參考數據
- **調試分析**: 分析 AlertManager 發送的 webhook 結構
- **文檔參考**: 了解完整的 AlertManager webhook 負載格式

#### 📊 包含的警報類型

樣本文件包含兩種狀態的警報：

1. **🔥 Firing 警報** (3 個)

   - Pod CPU 使用率超過 80% 的活躍警報
   - 不同 Pod 的 CPU 高使用率警報

2. **✅ Resolved 警報** (1 個)
   - 已解決的 CPU 使用率警報
   - 包含完整的開始和結束時間

#### 🧪 測試使用方法

```bash
# 使用 curl 測試 webhook 端點
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L4 \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic $(echo -n admin:admin | base64)" \
  -d @raw_alertmanager.json

# 測試其他平台
curl -X POST http://localhost:9999/api/v1/slack/chatid_L4 \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic $(echo -n admin:admin | base64)" \
  -d @raw_alertmanager.json

curl -X POST http://localhost:9999/api/v1/discord/chatid_L4 \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic $(echo -n admin:admin | base64)" \
  -d @raw_alertmanager.json
```

## 📁 項目結構

```
alert-webhooks/
├── cmd/                           # 主程序入口
│   └── main.go                   # 應用程式入口點
├── config/                        # 配置管理模組
│   ├── app.go                    # 應用程式配置
│   ├── config.go                 # 主配置文件
│   ├── discord.go                # Discord 配置
│   ├── logger.go                 # 日誌配置
│   ├── manager.go                # 配置管理器
│   ├── metric.go                 # 監控指標配置
│   ├── slack.go                  # Slack 配置
│   ├── telgram.go                # Telegram 配置
│   └── webhooks.go               # Webhook 配置
├── configs/                       # 配置文件目錄
│   ├── alert_config.minimal.yaml # 最小警報配置
│   ├── alert_config.yaml         # 完整警報配置
│   └── config.development.yaml   # 開發環境配置
├── docs/                          # Swagger API 文檔
│   ├── docs.go                   # 文檔生成器
│   ├── swagger.json              # JSON 格式 API 文檔
│   └── swagger.yaml              # YAML 格式 API 文檔
├── documentation/                 # 📖 項目說明文檔
│   ├── en/                       # 英文文檔
│   │   ├── config_guide.md       # 配置指南
│   │   ├── discord_setup.md      # Discord 設置
│   │   ├── discord_usage.md      # Discord 使用範例
│   │   ├── slack_setup.md        # Slack 設置
│   │   ├── slack_usage.md        # Slack 使用範例
│   │   ├── telegram_setup.md     # Telegram 設置
│   │   ├── telegram_usage.md     # Telegram 使用範例
│   │   └── template_guide.md     # 模板指南
│   └── zh/                       # 中文文檔
│       ├── config_guide.md       # 配置指南
│       ├── discord_setup.md      # Discord 設置
│       ├── discord_usage.md      # Discord 使用範例
│       ├── slack_setup.md        # Slack 設置
│       ├── slack_usage.md        # Slack 使用範例
│       ├── telegram_setup.md     # Telegram 設置
│       ├── telegram_usage.md     # Telegram 使用範例
│       └── template_guide.md     # 模板指南
├── examples/                      # 使用範例
│   ├── config_usage.go           # 配置使用範例
│   └── config.expamle.yaml       # 配置範例文件
├── pkg/                          # 核心功能包
│   ├── logcore/                  # 日誌核心
│   │   └── core.go              # 日誌核心實現
│   ├── logger/                   # 日誌系統
│   │   ├── logger.go            # 日誌器實現
│   │   ├── middleware.go        # 日誌中間件
│   │   └── utils.go             # 日誌工具函數
│   ├── logutil/                  # 日誌工具
│   │   └── context.go           # 上下文日誌工具
│   ├── middleware/               # HTTP 中間件
│   │   ├── basic_auth.go        # 基礎認證中間件
│   │   ├── cors.go              # CORS 中間件
│   │   ├── logger.go            # 日誌中間件
│   │   └── recovery.go          # 恢復中間件
│   ├── notification/             # 通知系統
│   │   ├── manager.go           # 通知管理器
│   │   ├── providers/           # 通知提供者
│   │   │   ├── discord.go       # Discord 通知實現
│   │   │   ├── slack.go         # Slack 通知實現
│   │   │   └── telegram.go      # Telegram 通知實現
│   │   └── types/               # 通知類型定義
│   │       └── types.go         # 通知類型結構
│   ├── service/                  # 業務服務層
│   │   ├── discord.go           # Discord 服務
│   │   ├── service.go           # 通用服務介面
│   │   ├── slack.go             # Slack 服務
│   │   └── telegram.go          # Telegram 服務
│   ├── template/                 # 模板引擎
│   │   └── engine.go            # 模板引擎實現
│   └── watcher/                  # 文件監控
│       └── config_watcher.go    # 配置文件監控器
├── routes/                       # API 路由系統
│   ├── api/                     # API 路由
│   │   └── v1/                  # API v1 版本
│   │       ├── discord/         # Discord API 路由
│   │       │   ├── handler.go   # Discord 處理器
│   │       │   └── routes.go    # Discord 路由定義
│   │       ├── slack/           # Slack API 路由
│   │       │   ├── handler.go   # Slack 處理器
│   │       │   └── routes.go    # Slack 路由定義
│   │       ├── telegram/        # Telegram API 路由
│   │       │   ├── handler.go   # Telegram 處理器
│   │       │   └── routes.go    # Telegram 路由定義
│   │       ├── healthCheck.go   # 健康檢查端點
│   │       └── register.go      # 路由註冊器
│   └── mainRoute.go             # 主路由配置
├── scripts/                      # 工具腳本
│   ├── fix_swagger_docs.go      # Swagger 文檔修復腳本
│   └── regenerate_swagger.sh    # Swagger 重新生成腳本
├── templates/                    # 消息模板
│   └── alerts/                  # 警報模板
│       ├── alert_template_eng.tmpl  # 英文警報模板
│       ├── alert_template_ja.tmpl   # 日文警報模板
│       ├── alert_template_ko.tmpl   # 韓文警報模板
│       ├── alert_template_tw.tmpl   # 繁體中文警報模板
│       └── alert_template_zh.tmpl   # 簡體中文警報模板
├── kubernetes/                   # Kubernetes 部署配置
│   └── deployment-example.yaml  # 部署範例配置
├── docker-compose.yml           # Docker Compose 配置
├── docker-compose.dev.yml       # 開發環境 Docker Compose
├── Dockerfile                   # Docker 映像構建文件
├── Makefile                     # 構建和管理腳本
├── go.mod                       # Go 模組依賴
├── go.sum                       # Go 模組校驗和
├── raw_alertmanager.json        # AlertManager webhook 負載樣本
├── README.md                    # 英文項目說明文件
└── README-zh.md                 # 中文項目說明文件
```

## 🌍 語言版本

- [English](./README.md)
- [繁體中文](./README-zh.md) (當前)

## 🤝 貢獻

歡迎提交 Issues 和 Pull Requests 來改進這個項目！

## 📄 授權

本項目採用 MIT 授權條款。
