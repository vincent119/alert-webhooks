# Alert Webhooks

一個高效的 AlertManager webhook 處理服務，支援多平台通知和多語言模板。

[![GitHub](https://img.shields.io/badge/github-vincent119%2Falert--webhooks-blue?logo=github)](https://github.com/vincent119/alert-webhooks)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.19%2B-blue?logo=go)](go.mod)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/vincent119/alert-webhooks/actions)
[![Coverage](https://img.shields.io/badge/coverage-89%25-yellow)](https://codecov.io/gh/vincent119/alert-webhooks)
[![Score](https://img.shields.io/badge/score-9.2-brightgreen)](https://goreportcard.com/report/github.com/vincent119/alert-webhooks)

## 🌟 主要功能

- 🔗 **AlertManager 整合**: 直接接收和處理 AlertManager webhook
- 📱 **Telegram 通知**: 支援多等級聊天群組通知
- 🌍 **多語言模板**: 英語、繁體中文、簡體中文、日語、韓語
- 🔄 **熱重載**: 配置文件和模板動態重載，無需重啟服務
- 🔐 **安全認證**: HTTP Basic Auth 保護
- 📋 **雙模板模式**: Full/Minimal 兩種顯示格式
- 📨 **分離通知**: 觸發中和已解決警報分別發送

## 🚀 快速開始

### 1. 安裝依賴

```bash
go mod download
```

### 2. 配置設置

```bash
# 複製配置文件
cp configs/config.development.yaml.example configs/config.development.yaml

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

# 📊 Chat IDs 群組用途對照表

| chat_ids  | Level | 群組用途 (Group Purpose)                    |
| --------- | ----- | ------------------------------------------- |
| chat_ids0 | 0     | Information group（資訊群組）               |
| chat_ids1 | 1     | General message group（一般訊息群組）       |
| chat_ids2 | 2     | Critical notification group（重要通知群組） |
| chat_ids3 | 3     | Emergency alert group（緊急警報群組）       |
| chat_ids4 | 4     | Testing group（測試群組）                   |
| chat_ids5 | 5     | Backup group（備用群組）                    |

## 📖 詳細文檔

完整的設置和使用指南提供多語言版本，請查看 **[documentation](./documentation/)** 目錄：

### 🌍 語言選擇

- **[English Documentation](./documentation/en/)** - 完整的英文文檔
- **[繁體中文文檔](./documentation/zh/)** - 完整的繁體中文文檔

### 📋 快速連結

- 🔧 **[配置指南](./documentation/zh/config_guide.md)** - 詳細配置說明
- 📱 **[Telegram 設置](./documentation/zh/telegram_setup.md)** - Telegram 機器人配置
- 📝 **[模板指南](./documentation/zh/template_guide.md)** - 自定義模板說明
- 📚 **[使用範例](./documentation/zh/telegram_usage.md)** - API 使用範例

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

| 方法   | 路徑                              | 描述               | 認證          |
| ------ | --------------------------------- | ------------------ | ------------- |
| `POST` | `/api/v1/telegram/chatid_{level}` | 發送 Telegram 訊息 | ✅ Basic Auth |
| `GET`  | `/api/v1/telegram/info`           | 獲取機器人資訊     | ✅ Basic Auth |
| `GET`  | `/api/v1/healthz`                 | 健康檢查           | ❌            |
| `GET`  | `/swagger/*`                      | API 文檔           | ❌            |

### AlertManager 整合範例

```yaml
# alertmanager.yml
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

## 📁 項目結構

```
alert-webhooks/
├── cmd/                    # 主程序入口
├── config/                 # 配置管理
├── configs/                # 配置文件
├── docs/                   # Swagger API 文檔
├── documentation/          # 📖 項目說明文檔
├── examples/               # 使用範例
├── pkg/                    # 核心功能包
│   ├── logger/            # 日誌系統
│   ├── middleware/        # 中間件
│   ├── service/           # 業務服務
│   ├── template/          # 模板引擎
│   └── watcher/           # 文件監控
├── routes/                 # API 路由
├── scripts/               # 工具腳本
└── templates/             # 消息模板
```

## 🌍 語言版本

- [English](./README.md)
- [繁體中文](./README-zh.md) (當前)

## 🤝 貢獻

歡迎提交 Issues 和 Pull Requests 來改進這個項目！

## 📄 授權

本項目採用 MIT 授權條款。
