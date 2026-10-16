# Alert Webhooks Documentation

This directory contains comprehensive documentation for the Alert Webhooks project in multiple languages.

## 🌍 Language Selection

- **[English Documentation](./en/)** - Complete English documentation
- **[繁體中文文檔](./zh/)** - 完整的繁體中文文檔

## 📁 Documentation Structure

### English Documentation (`/en/`)
- **[README.md](./en/README.md)** - Documentation overview and index
- **[config_guide.md](./en/config_guide.md)** - Configuration file detailed guide
- **[telegram_setup.md](./en/telegram_setup.md)** - Telegram bot setup guide
- **[template_mode_config.md](./en/template_mode_config.md)** - Template mode configuration
- **[template_guide.md](./en/template_guide.md)** - Template system guide
- **[template_usage.md](./en/template_usage.md)** - Template usage examples
- **[telegram_usage.md](./en/telegram_usage.md)** - Telegram API usage examples
- **[Helm installation](../kubernetes/helm/alert-webhooks-chart/README.md)** - Helm installation and lifecycle guide
- **[Kustomize installation](../kubernetes/kustomize/README.md)** - Kustomize deployment guide

### 繁體中文文檔 (`/zh/`)
- **[README.md](./zh/README.md)** - 文檔概覽和索引
- **[config_guide.md](./zh/config_guide.md)** - 配置文件詳細指南
- **[telegram_setup.md](./zh/telegram_setup.md)** - Telegram 機器人設置指南
- **[template_mode_config.md](./zh/template_mode_config.md)** - 模板模式配置說明
- **[template_guide.md](./zh/template_guide.md)** - 模板系統使用指南
- **[template_usage.md](./zh/template_usage.md)** - 模板使用範例
- **[telegram_usage.md](./zh/telegram_usage.md)** - Telegram API 使用範例
- **[Helm 安裝指南](../kubernetes/helm/alert-webhooks-chart/README.md)** - Helm 安裝與生命週期指南
- **[Kustomize 安裝指南](../kubernetes/kustomize/README.md)** - Kustomize 部署指南

## 🏗️ Project Structure

```
alert-webhooks/
├── cmd/                    # Main program entry / 主程序入口
├── config/                 # Configuration management / 配置管理
├── configs/                # Configuration files / 配置文件
├── docs/                   # Swagger API documentation / Swagger API 文檔
├── documentation/          # 📖 Project documentation / 項目說明文檔
│   ├── en/                # English documentation / 英文文檔
│   └── zh/                # Chinese documentation / 中文文檔
├── examples/               # Usage examples / 使用範例
├── pkg/                    # Core functionality packages / 核心功能包
├── routes/                 # API routes / API 路由
├── scripts/               # Utility scripts / 工具腳本
└── templates/             # Message templates / 消息模板
```

## 🚀 Quick Start

1. **Configuration**: Read [English config guide](./en/config_guide.md) or [中文配置指南](./zh/config_guide.md)
2. **Telegram Setup**: Read [English Telegram guide](./en/telegram_setup.md) or [中文 Telegram 指南](./zh/telegram_setup.md)
3. **Template Configuration**: Read [English template guide](./en/template_guide.md) or [中文模板指南](./zh/template_guide.md)
4. **API Documentation**: Visit `/swagger/index.html`

## 📊 API Documentation

The project provides complete Swagger API documentation including:
- Telegram message sending API
- Bot information query API
- Health check endpoint
- AlertManager webhook integration

Access methods:
- **Swagger UI**: `http://localhost:9999/swagger/index.html`
- **JSON format**: `http://localhost:9999/swagger/doc.json`

## 🌟 Key Features

- ✅ **AlertManager Integration**: Direct AlertManager webhook processing
- ✅ **Multilingual Templates**: Support for English, Traditional Chinese, Simplified Chinese, Japanese, Korean
- ✅ **Hot Reload**: Dynamic configuration and template reloading
- ✅ **Basic Authentication**: HTTP Basic Auth security protection
- ✅ **Template Modes**: Full/Minimal display modes
- ✅ **Separate Notifications**: Firing and resolved alerts sent separately

## 🛠️ Development Tools

Use Makefile to manage common tasks:
```bash
make dev              # Start development environment
make swagger-generate # Generate Swagger documentation
make test            # Run tests
make build           # Build project
```

## 📞 Support

For questions, please refer to the relevant documentation or contact the development team.
