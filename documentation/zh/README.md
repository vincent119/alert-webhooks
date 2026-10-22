# Alert Webhooks 專案文件（繁體中文）

這個目錄包含了 Alert Webhooks 專案的完整繁體中文說明文件。

## 📁 文件結構

### 🔧 部署與設定文件

- **[docker-deployment.md](./docker-deployment.md)** - 公開 Docker image 部署、
  驗證、回滾與故障排除
- **[config_guide.md](./config_guide.md)** - 設定檔詳細指南
- **[service-enable-config.md](./service-enable-config.md)** - 服務啟用設定指南
- **[kubernetes-env-vars.md](./kubernetes-env-vars.md)** - Kubernetes 環境變數設定
- **[Helm 安裝指南](../../kubernetes/helm/alert-webhooks-chart/README.zh-TW.md)** - Helm 前置條件、安裝、升級與回滾
- **[Kustomize 安裝指南](../../kubernetes/kustomize/README.md)** - Kustomize 前置條件、部署、驗證與清理
- **[telegram_setup.md](./telegram_setup.md)** - Telegram 機器人設定指南
- **[slack_setup.md](./slack_setup.md)** - Slack 機器人設定指南
- **[template_mode_config.md](./template_mode_config.md)** - 模板模式設定說明

### 📝 模板文件

- **[template_guide.md](./template_guide.md)** - 模板系統使用指南
- **[template_usage.md](./template_usage.md)** - 模板使用範例

### 📱 使用範例

- **[telegram_usage.md](./telegram_usage.md)** - Telegram API 使用範例
- **[slack_usage.md](./slack_usage.md)** - Slack API 使用範例

## 🏗️ 專案結構

```
alert-webhooks/
├── cmd/                    # 主程式入口
├── config/                 # 設定管理
├── configs/                # 設定檔
├── docs/                   # Swagger API 文件
├── documentation/          # 📖 專案說明文件（目前目錄）
├── examples/               # 使用範例
├── pkg/                    # 核心功能套件
│   ├── logger/            # 日誌系統
│   ├── middleware/        # middleware
│   ├── service/           # 業務服務
│   ├── template/          # 模板引擎
│   └── watcher/           # 檔案監控
├── routes/                 # API 路由
├── scripts/               # 工具腳本
└── templates/             # 訊息模板
```

## 🚀 快速開始

1. **Docker 部署**：閱讀 [docker-deployment.md](./docker-deployment.md)
2. **準備設定檔**：閱讀 [config_guide.md](./config_guide.md)
3. **通知設定**：閱讀 [telegram_setup.md](./telegram_setup.md) 或 [slack_setup.md](./slack_setup.md)
4. **模板設定**：閱讀 [template_guide.md](./template_guide.md)
5. **API 文件**：存取 `/swagger/index.html`

## 📊 API 文件

專案提供完整的 Swagger API 文件，包含：
- Telegram 訊息傳送 API
- Slack 訊息傳送 API
- 機器人資訊查詢 API
- 健康檢查端點
- AlertManager webhook 整合

存取方式：
- **Swagger UI**: `http://localhost:9999/swagger/index.html`
- **JSON 格式**: `http://localhost:9999/swagger/doc.json`

## 🌟 主要功能

- ✅ **AlertManager 整合**: 直接接收 AlertManager webhook
- ✅ **多語言模板**: 支援英語、繁體中文、簡體中文、日語、韓語
- ✅ **Hot reload**: 設定檔和模板可動態重新載入
- ✅ **基本認證**: HTTP Basic Auth 安全保護
- ✅ **模板模式**: Full/Minimal 兩種顯示模式
- ✅ **分離通知**: 觸發中和已解決警報分開傳送

## 🛠️ 開發工具

使用 Makefile 管理常用任務：
```bash
make dev              # 啟動開發環境
make swagger-generate # 產生 Swagger 文件
make test            # 執行測試
make build           # 編譯專案
```

## 🌍 語言選項

- [English](../en/)
- [繁體中文](../zh/) (目前)

## 📞 支援

如有問題，請參考相關文件或聯絡開發團隊。
