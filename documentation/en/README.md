# Alert Webhooks Documentation (English)

This directory contains comprehensive English documentation for the Alert Webhooks project.

## 📁 Documentation Structure

### 🔧 Configuration Documentation

- **[docker-deployment.md](./docker-deployment.md)** - Public Docker image
  deployment, verification, rollback, and troubleshooting
- **[config_guide.md](./config_guide.md)** - Detailed configuration file guide
- **[service-enable-config.md](./service-enable-config.md)** - Service enable configuration guide
- **[kubernetes-env-vars.md](./kubernetes-env-vars.md)** - Kubernetes environment variables configuration
- **[Helm installation](../../kubernetes/helm/alert-webhooks-chart/README.md)** -
  Helm prerequisites, installation, upgrade, and rollback
- **[Kustomize installation](../../kubernetes/kustomize/README.en.md)** -
  Kustomize prerequisites, deployment, verification, and cleanup
- **[telegram_setup.md](./telegram_setup.md)** - Telegram bot setup guide
- **[slack_setup.md](./slack_setup.md)** - Slack bot setup guide
- **[template_mode_config.md](./template_mode_config.md)** - Template mode configuration

### 📝 Template Documentation

- **[template_guide.md](./template_guide.md)** - Template system usage guide
- **[template_usage.md](./template_usage.md)** - Template usage examples

### 📱 Usage Examples

- **[telegram_usage.md](./telegram_usage.md)** - Telegram API usage examples
- **[slack_usage.md](./slack_usage.md)** - Slack API usage examples

## 🏗️ Project Structure

```
alert-webhooks/
├── cmd/                    # Main program entry
├── config/                 # Configuration management
├── configs/                # Configuration files
├── docs/                   # Swagger API documentation
├── documentation/          # 📖 Project documentation (current directory)
├── examples/               # Usage examples
├── pkg/                    # Core functionality packages
│   ├── logger/            # Logging system
│   ├── middleware/        # Middleware
│   ├── service/           # Business services
│   ├── template/          # Template engine
│   └── watcher/           # File monitoring
├── routes/                 # API routes
├── scripts/               # Utility scripts
└── templates/             # Message templates
```

## 🚀 Quick Start

1. **Docker Deployment**: Read [docker-deployment.md](./docker-deployment.md)
2. **Configuration Setup**: Read [config_guide.md](./config_guide.md)
3. **Notification Setup**: Read [telegram_setup.md](./telegram_setup.md) or [slack_setup.md](./slack_setup.md)
4. **Template Configuration**: Read [template_guide.md](./template_guide.md)
5. **API Documentation**: Visit `/swagger/index.html`

## 📊 API Documentation

The project provides complete Swagger API documentation including:

- Telegram message sending API
- Slack message sending API
- Bot information query API
- Health check endpoint
- AlertManager webhook integration

Access methods:

- **Swagger UI**: `http://localhost:9999/swagger/index.html`
- **JSON format**: `http://localhost:9999/swagger/doc.json`

## 🌟 Main Features

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

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)

## 📞 Support

For questions, please refer to the relevant documentation or contact the development team.
