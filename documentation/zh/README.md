# Alert Webhooks 項目文檔（繁體中文）

這個目錄包含了 Alert Webhooks 項目的完整繁體中文說明文檔。

## 📁 文檔結構

### 🔧 配置文檔
- **[config_guide.md](./config_guide.md)** - 配置文件詳細指南
- **[service-enable-config.md](./service-enable-config.md)** - 服務啟用配置指南
- **[kubernetes-env-vars.md](./kubernetes-env-vars.md)** - Kubernetes 環境變數配置
- **[Helm 安裝指南](../../kubernetes/helm/alert-webhooks-chart/README.md)** - Helm 前置條件、安裝、升級與回滾
- **[Kustomize 安裝指南](../../kubernetes/kustomize/README.md)** - Kustomize 前置條件、部署、驗證與清理
- **[telegram_setup.md](./telegram_setup.md)** - Telegram 機器人設置指南
- **[slack_setup.md](./slack_setup.md)** - Slack 機器人設置指南
- **[template_mode_config.md](./template_mode_config.md)** - 模板模式配置說明

### 📝 模板文檔
- **[template_guide.md](./template_guide.md)** - 模板系統使用指南
- **[template_usage.md](./template_usage.md)** - 模板使用範例

### 📱 使用範例
- **[telegram_usage.md](./telegram_usage.md)** - Telegram API 使用範例
- **[slack_usage.md](./slack_usage.md)** - Slack API 使用範例

## 🏗️ 項目結構

```
alert-webhooks/
├── cmd/                    # 主程序入口
├── config/                 # 配置管理
├── configs/                # 配置文件
├── docs/                   # Swagger API 文檔
├── documentation/          # 📖 項目說明文檔（當前目錄）
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

## 🚀 快速開始

1. **配置設置**: 閱讀 [config_guide.md](./config_guide.md)
2. **通知設置**: 閱讀 [telegram_setup.md](./telegram_setup.md) 或 [slack_setup.md](./slack_setup.md)
3. **模板配置**: 閱讀 [template_guide.md](./template_guide.md)
4. **API 文檔**: 訪問 `/swagger/index.html`

## 📊 API 文檔

項目提供完整的 Swagger API 文檔，包含：
- Telegram 訊息發送 API
- Slack 訊息發送 API
- 機器人資訊查詢 API
- 健康檢查端點
- AlertManager webhook 整合

訪問方式：
- **Swagger UI**: `http://localhost:9999/swagger/index.html`
- **JSON 格式**: `http://localhost:9999/swagger/doc.json`

## 🌟 主要功能

- ✅ **AlertManager 整合**: 直接接收 AlertManager webhook
- ✅ **多語言模板**: 支援英語、繁體中文、簡體中文、日語、韓語
- ✅ **熱重載**: 配置文件和模板動態重載
- ✅ **基本認證**: HTTP Basic Auth 安全保護
- ✅ **模板模式**: Full/Minimal 兩種顯示模式
- ✅ **分離通知**: 觸發中和已解決警報分開發送

## 🛠️ 開發工具

使用 Makefile 管理常用任務：
```bash
make dev              # 啟動開發環境
make swagger-generate # 生成 Swagger 文檔
make test            # 運行測試
make build           # 編譯項目
```

## 🌍 語言選項

- [English](../en/)
- [繁體中文](../zh/) (當前)

## 📞 支援

如有問題，請參考相關文檔或聯繫開發團隊。
