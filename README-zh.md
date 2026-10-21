# Alert Webhooks

一個高效的 AlertManager webhook 處理服務，支援多平台通知和多語言模板。

[![GitHub](https://img.shields.io/badge/github-vincent119%2Falert--webhooks-blue?logo=github)](https://github.com/vincent119/alert-webhooks)
![License](https://img.shields.io/github/license/awslabs/mcp)
[![Go Version](https://img.shields.io/badge/go-1.19%2B-blue?logo=go)](go.mod)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/vincent119/alert-webhooks/actions)
[![Coverage](https://codecov.io/gh/vincent119/alert-webhooks/branch/main/graph/badge.svg)](https://codecov.io/gh/vincent119/alert-webhooks)
[![Score](https://img.shields.io/badge/score-9.2-brightgreen)](https://goreportcard.com/report/github.com/vincent119/alert-webhooks)
[![GitHub stars](https://img.shields.io/github/stars/vincent119/alert-webhooks?style=flat&color=yellow&logo=github&logoColor=white)](https://github.com/vincent119/alert-webhooks/stargazers)
![GitHub release](https://img.shields.io/github/v/release/awslabs/mcp)

## 📋 目錄

- [🌟 主要功能](#-主要功能)
- [🚀 支援平台](#-支援平台)
- [⚡ 快速開始](#-快速開始)
- [公開 container image](#公開-container-image)
- [📊 Chat IDs 群組用途對照表](#-chat-ids-群組用途對照表)
- [📖 詳細文件](#-詳細文件)
- [🛠️ 開發](#️-開發)
- [📄 AlertManager Webhook 樣本](#-alertmanager-webhook-樣本)
- [📁 專案結構](#-專案結構)
- [🌍 語言版本](#-語言版本)

## 🌟 主要功能

- 🔗 **AlertManager 整合**: 直接接收和處理 AlertManager webhook
- 📱 **多平台通知**: 支援 Telegram、Slack、Discord 三大通訊平台
- 🎯 **多等級通知**: 支援依警報等級分送至不同群組
- 🌍 **多語言模板**: 英語、繁體中文、簡體中文、日語、韓語
- 🔄 **Hot reload**: 設定檔與模板變更後不必重新啟動服務
- 🔐 **安全認證**: HTTP Basic Auth 保護
- 📋 **雙模板模式**: Full/Minimal 兩種顯示格式
- 📨 **分離通知**: 觸發中和已解決警報分別傳送
- 🎨 **自訂模板**: 支援自訂訊息模板和格式

## 🚀 支援平台

本服務支援以下通訊平台的警報通知：

### 📱 Telegram

- ✅ 支援多個聊天群組
- ✅ 支援依警報等級分送通知
- ✅ 支援機器人資訊查詢
- ✅ 支援自訂訊息格式

### 💬 Slack

- ✅ 支援 Webhook 通知
- ✅ 支援頻道訊息傳送
- ✅ 支援自訂訊息格式
- ✅ 支援附件和格式化訊息

### 🎮 Discord

- ✅ 支援伺服器頻道通知
- ✅ 支援 Webhook 訊息
- ✅ 支援豐富的訊息格式
- ✅ 支援嵌入式訊息

## ⚡ 快速開始

### 1. 安裝依賴

```bash
go mod download
```

### 2. 準備設定檔

```bash
# 複製設定檔
cp examples/config.expamle configs/config.development.yaml

# 編輯設定檔（設定 Telegram token 與 chat IDs）
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

### 4. 存取 API 文件

開啟瀏覽器前往：<http://localhost:9999/swagger/index.html>

## 公開 container image

公開發布的 container image 位於 Docker Hub：
`docker.io/vincent119/alert-webhooks:v2.0.8`，支援 `linux/amd64` 與
`linux/arm64`。

發布流程只建立不可變的固定版本 tag，不發布 Docker `latest`；部署時必須明確指定版本。
Docker Hub 是公開發布來源，既有 private ECR production 部署流程維持不變。

```bash
docker pull docker.io/vincent119/alert-webhooks:v2.0.8

config_file="$(pwd)/examples/config.expamle.yaml"
docker run --detach \
  --name alert-webhooks \
  --publish 127.0.0.1:9999:9999 \
  --volume "${config_file}:/app/configs/config.production.yaml:ro" \
  --env APP_ENV=production \
  docker.io/vincent119/alert-webhooks:v2.0.8

curl --fail http://127.0.0.1:9999/healthy
```

容器以非 root UID `1500` 執行。發布契約與部署細節請參考
[v2.0.8 Release](https://github.com/vincent119/alert-webhooks/releases/tag/v2.0.8)
與 [Docker 部署指南](./documentation/zh/docker-deployment.md)。

## 📊 Chat IDs 群組用途對照表

| Chat IDs  | Level | 群組用途                                    | 說明                   |
| --------- | ----- | ------------------------------------------- | ---------------------- |
| chat_ids0 | L0    | Information Group（資訊群組）               | 一般資訊和狀態更新通知 |
| chat_ids1 | L1    | General Message Group（一般訊息群組）       | 標準警報和日常監控通知 |
| chat_ids2 | L2    | Critical Notification Group（重要通知群組） | 重要警報和關鍵系統通知 |
| chat_ids3 | L3    | Emergency Alert Group（緊急警報群組）       | 緊急事件和嚴重故障通知 |
| chat_ids4 | L4    | Testing Group（測試群組）                   | 測試和開發環境通知     |
| chat_ids5 | L5    | Backup Group（備用群組）                    | 備用和災難復原通知群組 |

## 📖 詳細文件

完整的設定與使用指南提供多語言版本，請查看 **[documentation](./documentation/)** 目錄：

### 🌍 語言選擇

- **[English Documentation](./documentation/en/)** - 完整的英文文件
- **[繁體中文文件](./documentation/zh/)** - 完整的繁體中文文件

### 📋 快速連結

#### 🔧 基礎設定

- **[設定指南](./documentation/zh/config_guide.md)** - 詳細設定說明
- **[服務啟用設定](./documentation/zh/service-enable-config.md)** - 服務啟用設定
- **[Kubernetes 環境變數](./documentation/zh/kubernetes-env-vars.md)** - K8s 部署設定
- **[Helm 安裝指南](./kubernetes/helm/alert-webhooks-chart/README.md)** - 使用 Helm 安裝、升級、驗證與回滾
- **[Kustomize 安裝指南](./kubernetes/kustomize/README.md)** - 使用 Kustomize render、部署、驗證與清理
- **[Swagger 疑難排解](./documentation/zh/swagger-troubleshooting.md)** - API 文件問題解決

#### 📝 模板系統

- **[模板指南](./documentation/zh/template_guide.md)** - 自訂模板說明
- **[模板模式設定](./documentation/zh/template_mode_config.md)** - Full/Minimal 模式設定
- **[模板使用範例](./documentation/zh/template_usage.md)** - 模板實際應用範例

#### 📱 平台設定指南

- **[Telegram 設定](./documentation/zh/telegram_setup.md)** - Telegram 機器人設定
- **[Slack 設定](./documentation/zh/slack_setup.md)** - Slack 應用程式設定
- **[Discord 設定](./documentation/zh/discord_setup.md)** - Discord 機器人設定

#### 📚 平台使用範例

- **[Telegram 使用範例](./documentation/zh/telegram_usage.md)** - Telegram API 使用範例
- **[Slack 使用範例](./documentation/zh/slack_usage.md)** - Slack API 使用範例
- **[Discord 使用範例](./documentation/zh/discord_usage.md)** - Discord API 使用範例

## 🛠️ 開發

### Makefile 命令

```bash
make dev              # 啟動開發環境
make build            # 編譯專案
make test             # 執行測試
make swagger-generate # 重新產生 Swagger 文件
make fmt              # 格式化程式碼
make lint             # 程式碼品質檢查
```

### API 端點

#### 📱 Telegram API

| 方法   | 路徑                              | 描述               | 認證          |
| ------ | --------------------------------- | ------------------ | ------------- |
| `POST` | `/api/v1/telegram/chatid_{level}` | 傳送 Telegram 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/telegram/info`           | 取得機器人資訊     | ✅ Basic Auth |

#### 💬 Slack API

| 方法   | 路徑                           | 描述            | 認證          |
| ------ | ------------------------------ | --------------- | ------------- |
| `POST` | `/api/v1/slack/chatid_{level}` | 傳送 Slack 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/slack/info`           | 取得 Slack 資訊 | ✅ Basic Auth |

#### 🎮 Discord API

| 方法   | 路徑                             | 描述              | 認證          |
| ------ | -------------------------------- | ----------------- | ------------- |
| `POST` | `/api/v1/discord/chatid_{level}` | 傳送 Discord 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/discord/info`           | 取得 Discord 資訊 | ✅ Basic Auth |

#### 🔧 系統 API

| 方法  | 路徑              | 描述     | 認證          |
| ----- | ----------------- | -------- | ------------- |
| `GET` | `/api/v1/healthz` | 健康檢查 | ❌            |
| `GET` | `/swagger/*`      | API 文件 | ✅ Basic Auth |

### 可觀測性

#### 📊 Prometheus Metrics

服務在 `/api/v1/metrics` 提供 Prometheus 指標。

| 方法  | 路徑               | 描述             | 認證                    |
| ----- | ------------------ | ---------------- | ----------------------- |
| `GET` | `/api/v1/metrics`  | Prometheus 指標  | 可選 Basic Auth         |

當 `metric.user` 和 `metric.password` 同時設定時，該端點需要 Basic Auth 認證。否則為公開存取。

```yaml
# config.yaml - Metrics 設定
metric:
  user: "prometheus"
  password: "secret"
```

環境變數覆蓋：`METRIC_USER`、`METRIC_PASSWORD`

Prometheus 抓取範例：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: "alert-webhooks"
    metrics_path: "/api/v1/metrics"
    basic_auth:
      username: "prometheus"
      password: "secret"
    static_configs:
      - targets: ["localhost:9999"]
```

#### 🔭 OpenTelemetry Tracing

服務支援透過 OpenTelemetry（OTLP over HTTP）進行分散式追蹤。啟用後，所有 HTTP 請求和通知傳送（Telegram、Slack、Discord）都會自動追蹤。

```yaml
# config.yaml - Trace 設定
trace:
  enable: true
  url: "otel-collector.monitoring.svc.cluster.local"
  port: "4318"
  urlPath: "/v1/traces"
  insecure: true          # true = HTTP，false = HTTPS
  tlsSkipVerify: false    # 跳過 TLS 憑證驗證（自簽憑證）
  authUser: ""            # OTLP collector 的 Basic Auth
  authPasswd: ""
  sampleRate: 1.0         # 1.0 = 100%，0.5 = 50%，0.0 = 停用
```

| 欄位            | 說明                                         | 預設值        |
| --------------- | -------------------------------------------- | ------------- |
| `enable`        | 啟用/停用追蹤                                | `false`       |
| `url`           | OTLP collector 主機位址                      | `""`          |
| `port`          | OTLP HTTP receiver port                      | `"4318"`      |
| `urlPath`       | OTLP 端點路徑                                | `/v1/traces`  |
| `insecure`      | `true` = HTTP，`false` = HTTPS               | `true`        |
| `tlsSkipVerify` | 跳過 TLS 憑證驗證                            | `false`       |
| `authUser`      | collector 的 Basic Auth 帳號                 | `""`          |
| `authPasswd`    | collector 的 Basic Auth 密碼                 | `""`          |
| `sampleRate`    | 取樣率（0.0 ~ 1.0）                         | `1.0`         |

環境變數覆蓋：`TRACE_AUTH_USER`、`TRACE_AUTH_PASSWD`

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

專案根目錄中的 `raw_alertmanager.json` 檔案提供完整的 Prometheus
AlertManager webhook payload 範例，包含：

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
- **模板開發**: 開發自訂警報模板時的參考資料
- **除錯分析**: 分析 AlertManager 傳送的 webhook 結構
- **格式參考**: 了解完整的 AlertManager webhook payload 格式

#### 📊 包含的警報類型

樣本檔案包含兩種狀態的警報：

1. **🔥 Firing 警報** (3 個)

   - Pod CPU 使用率超過 80% 的作用中警報
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

## 📁 專案結構

```
alert-webhooks/
├── cmd/                           # 主程式入口
│   └── main.go                   # 應用程式入口點
├── config/                        # 設定管理模組
│   ├── app.go                    # 應用程式設定
│   ├── config.go                 # 主設定檔
│   ├── discord.go                # Discord 設定
│   ├── logger.go                 # 日誌設定
│   ├── manager.go                # 設定 manager
│   ├── metric.go                 # 監控指標設定
│   ├── slack.go                  # Slack 設定
│   ├── telgram.go                # Telegram 設定
│   └── webhooks.go               # Webhook 設定
├── configs/                       # 設定檔目錄
│   ├── alert_config.minimal.yaml # 最小警報設定
│   ├── alert_config.yaml         # 完整警報設定
│   └── config.development.yaml   # 開發環境設定
├── docs/                          # Swagger API 文件
│   ├── docs.go                   # 文件產生器
│   ├── swagger.json              # JSON 格式 API 文件
│   └── swagger.yaml              # YAML 格式 API 文件
├── documentation/                 # 📖 專案說明文件
│   ├── en/                       # 英文文件
│   │   ├── config_guide.md       # 設定指南
│   │   ├── discord_setup.md      # Discord 設定
│   │   ├── discord_usage.md      # Discord 使用範例
│   │   ├── slack_setup.md        # Slack 設定
│   │   ├── slack_usage.md        # Slack 使用範例
│   │   ├── telegram_setup.md     # Telegram 設定
│   │   ├── telegram_usage.md     # Telegram 使用範例
│   │   └── template_guide.md     # 模板指南
│   └── zh/                       # 中文文件
│       ├── config_guide.md       # 設定指南
│       ├── discord_setup.md      # Discord 設定
│       ├── discord_usage.md      # Discord 使用範例
│       ├── slack_setup.md        # Slack 設定
│       ├── slack_usage.md        # Slack 使用範例
│       ├── telegram_setup.md     # Telegram 設定
│       ├── telegram_usage.md     # Telegram 使用範例
│       └── template_guide.md     # 模板指南
├── examples/                      # 使用範例
│   ├── config_usage.go           # 設定使用範例
│   └── config.expamle.yaml       # 設定範例檔案
├── pkg/                          # 核心功能套件
│   ├── logcore/                  # 日誌核心
│   │   └── core.go              # 日誌核心實作
│   ├── logger/                   # 日誌系統
│   │   ├── logger.go            # logger 實作
│   │   ├── middleware.go        # logging middleware
│   │   └── utils.go             # 日誌工具函式
│   ├── logutil/                  # 日誌工具
│   │   └── context.go           # 上下文日誌工具
│   ├── middleware/               # HTTP middleware
│   │   ├── basic_auth.go        # Basic Auth middleware
│   │   ├── cors.go              # CORS middleware
│   │   ├── logger.go            # logging middleware
│   │   └── recovery.go          # recovery middleware
│   ├── notification/             # 通知系統
│   │   ├── manager.go           # 通知 manager
│   │   ├── providers/           # 通知提供者
│   │   │   ├── discord.go       # Discord 通知實作
│   │   │   ├── slack.go         # Slack 通知實作
│   │   │   └── telegram.go      # Telegram 通知實作
│   │   └── types/               # 通知類型定義
│   │       └── types.go         # 通知類型結構
│   ├── service/                  # 業務服務層
│   │   ├── discord.go           # Discord 服務
│   │   ├── service.go           # 通用服務介面
│   │   ├── slack.go             # Slack 服務
│   │   └── telegram.go          # Telegram 服務
│   ├── template/                 # 模板引擎
│   │   └── engine.go            # 模板引擎實作
│   └── watcher/                  # 檔案監控
│       └── config_watcher.go    # 設定檔監控器
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
│   └── mainRoute.go             # 主路由設定
├── scripts/                      # 工具腳本
│   ├── fix_swagger_docs.go      # Swagger 文件修復腳本
│   └── regenerate_swagger.sh    # Swagger 重新產生腳本
├── templates/                    # 訊息模板
│   └── alerts/                  # 警報模板
│       ├── alert_template_eng.tmpl  # 英文警報模板
│       ├── alert_template_ja.tmpl   # 日文警報模板
│       ├── alert_template_ko.tmpl   # 韓文警報模板
│       ├── alert_template_tw.tmpl   # 繁體中文警報模板
│       └── alert_template_zh.tmpl   # 簡體中文警報模板
├── kubernetes/                   # Kubernetes 部署設定
│   ├── helm/                    # Helm chart 與安裝指南
│   └── kustomize/               # Kustomize manifests 與安裝指南
├── docker-compose.yml           # Docker Compose 設定
├── Dockerfile                   # Docker image 建置檔案
├── Makefile                     # 建置和管理腳本
├── go.mod                       # Go 模組依賴
├── go.sum                       # Go 模組校驗和
├── raw_alertmanager.json        # AlertManager webhook payload 範例
├── README.md                    # 英文專案說明文件
└── README-zh.md                 # 中文專案說明文件
```

## 🌍 語言版本

- [English](./README.md)
- [繁體中文](./README-zh.md) (目前)

## 🤝 貢獻

歡迎提交 Issues 和 Pull Requests 來改進這個專案！

## 📄 授權

本專案採用 MIT 授權條款。
