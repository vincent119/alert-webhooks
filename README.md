# Alert Webhooks

An efficient AlertManager webhook processing service with multi-platform notifications and multilingual template support.

[![GitHub](https://img.shields.io/badge/github-vincent119%2Falert--webhooks-blue?logo=github)](https://github.com/vincent119/alert-webhooks)
![License](https://img.shields.io/github/license/awslabs/mcp)
[![Go Version](https://img.shields.io/badge/go-1.19%2B-blue?logo=go)](go.mod)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/vincent119/alert-webhooks/actions)
[![Coverage](https://img.shields.io/badge/coverage-89%25-yellow)](https://codecov.io/gh/vincent119/alert-webhooks)
[![Score](https://img.shields.io/badge/score-9.2-brightgreen)](https://goreportcard.com/report/github.com/vincent119/alert-webhooks)
[![GitHub stars](https://img.shields.io/github/stars/vincent119/alert-webhooks?style=flat&color=yellow&logo=github&logoColor=white)](https://github.com/vincent119/alert-webhooks/stargazers)
![GitHub release](https://img.shields.io/github/v/release/awslabs/mcp)

## 📋 Table of Contents

- [🌟 Key Features](#-key-features)
- [🚀 Supported Platforms](#-supported-platforms)
- [⚡ Quick Start](#-quick-start)
- [📊 Chat IDs Level Mapping](#-chat-ids-level-mapping)
- [📖 Documentation](#-documentation)
- [🛠️ Development](#️-development)
- [📄 AlertManager Webhook Sample](#-alertmanager-webhook-sample)
- [📁 Project Structure](#-project-structure)
- [🌍 Languages](#-languages)

## 🌟 Key Features

- 🔗 **AlertManager Integration**: Direct webhook processing from AlertManager
- 📱 **Multi-Platform Notifications**: Support for Telegram, Slack, and Discord
- 🎯 **Multi-Level Notifications**: Support for different alert level group distribution
- 🌍 **Multilingual Templates**: English, Traditional Chinese, Simplified Chinese, Japanese, Korean
- 🔄 **Hot Reload**: Dynamic configuration and template reloading without service restart
- 🔐 **Secure Authentication**: HTTP Basic Auth protection
- 📋 **Dual Template Modes**: Full/Minimal display formats
- 📨 **Separate Notifications**: Firing and resolved alerts sent separately
- 🎨 **Custom Templates**: Support for custom message templates and formats

## 🚀 Supported Platforms

This service supports alert notifications for the following communication platforms:

### 📱 Telegram

- ✅ Support for multiple chat groups
- ✅ Support for different level notification distribution
- ✅ Support for bot information queries
- ✅ Support for custom message formats

### 💬 Slack

- ✅ Support for Webhook notifications
- ✅ Support for channel message sending
- ✅ Support for custom message formats
- ✅ Support for attachments and formatted messages

### 🎮 Discord

- ✅ Support for server channel notifications
- ✅ Support for Webhook messages
- ✅ Support for rich message formats
- ✅ Support for embedded messages

## ⚡ Quick Start

### 1. Install Dependencies

```bash
go mod download
```

### 2. Configuration Setup

```bash
# Copy configuration file
cp examples/config.expamle.yaml configs/config.development.yaml

# Edit configuration (set Telegram token and chat IDs)
vim configs/config.development.yaml
```

### 3. Start Service

```bash
# Development environment
make dev
# or
go run cmd/main.go -e development

# Production environment
make run
# or
go run cmd/main.go -e production
```

### 4. Access API Documentation

Open browser: <http://localhost:9999/swagger/index.html>

### 📊 Chat IDs Level Mapping

| Chat IDs  | Level | Group Purpose               | Description                                        |
| --------- | ----- | --------------------------- | -------------------------------------------------- |
| chat_ids0 | L0    | Information Group           | General information and status updates             |
| chat_ids1 | L1    | General Message Group       | Standard alerts and daily monitoring               |
| chat_ids2 | L2    | Critical Notification Group | Important alerts and critical system notifications |
| chat_ids3 | L3    | Emergency Alert Group       | Emergency events and severe failure notifications  |
| chat_ids4 | L4    | Testing Group               | Testing and development environment notifications  |
| chat_ids5 | L5    | Backup Group                | Backup and disaster recovery notification group    |

## 📖 Documentation

Complete setup and usage guides are available in multiple languages in the **[documentation](./documentation/)** directory:

### 🌍 Language Selection

- **[English Documentation](./documentation/en/)** - Complete English documentation
- **[繁體中文文檔](./documentation/zh/)** - Complete Traditional Chinese documentation

### 📋 Quick Links

#### 🔧 Basic Configuration

- **[Configuration Guide](./documentation/en/config_guide.md)** - Detailed configuration instructions
- **[Service Enable Configuration](./documentation/en/service-enable-config.md)** - Service enablement settings
- **[Kubernetes Environment Variables](./documentation/en/kubernetes-env-vars.md)** - K8s deployment configuration
- **[Swagger Troubleshooting](./documentation/en/swagger-troubleshooting.md)** - API documentation issue resolution

#### 📝 Template System

- **[Template Guide](./documentation/en/template_guide.md)** - Custom template instructions
- **[Template Mode Configuration](./documentation/en/template_mode_config.md)** - Full/Minimal mode settings
- **[Template Usage Examples](./documentation/en/template_usage.md)** - Template practical application examples

#### 📱 Platform Setup Guides

- **[Telegram Setup](./documentation/en/telegram_setup.md)** - Telegram bot configuration
- **[Slack Setup](./documentation/en/slack_setup.md)** - Slack application configuration
- **[Discord Setup](./documentation/en/discord_setup.md)** - Discord bot configuration

#### 📚 Platform Usage Examples

- **[Telegram Usage Examples](./documentation/en/telegram_usage.md)** - Telegram API usage examples
- **[Slack Usage Examples](./documentation/en/slack_usage.md)** - Slack API usage examples
- **[Discord Usage Examples](./documentation/en/discord_usage.md)** - Discord API usage examples

## 🛠️ Development

### Makefile Commands

```bash
make dev              # Start development environment
make build            # Build project
make test             # Run tests
make swagger-generate # Regenerate Swagger documentation
make fmt              # Format code
make lint             # Code quality check
```

### API Endpoints

#### 📱 Telegram API

| Method | Path                              | Description           | Authentication |
| ------ | --------------------------------- | --------------------- | -------------- |
| `POST` | `/api/v1/telegram/chatid_{level}` | Send Telegram message | ✅ Basic Auth  |
| `GET`  | `/api/v1/telegram/info`           | Get bot information   | ✅ Basic Auth  |

#### 💬 Slack API

| Method | Path                           | Description        | Authentication |
| ------ | ------------------------------ | ------------------ | -------------- |
| `POST` | `/api/v1/slack/chatid_{level}` | Send Slack message | ✅ Basic Auth  |
| `GET`  | `/api/v1/slack/info`           | Get Slack info     | ✅ Basic Auth  |

#### 🎮 Discord API

| Method | Path                             | Description          | Authentication |
| ------ | -------------------------------- | -------------------- | -------------- |
| `POST` | `/api/v1/discord/chatid_{level}` | Send Discord message | ✅ Basic Auth  |
| `GET`  | `/api/v1/discord/info`           | Get Discord info     | ✅ Basic Auth  |

#### 🔧 System API

| Method | Path              | Description       | Authentication |
| ------ | ----------------- | ----------------- | -------------- |
| `GET`  | `/api/v1/healthz` | Health check      | ❌             |
| `GET`  | `/swagger/*`      | API documentation | ✅ Basic Auth  |

### AlertManager Integration Example

#### 📱 Telegram Notification Setup

```yaml
# alertmanager.yml - Telegram Setup
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

#### 💬 Slack Notification Setup

```yaml
# alertmanager.yml - Slack Setup
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

#### 🎮 Discord Notification Setup

```yaml
# alertmanager.yml - Discord Setup
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

#### 🔄 Multi-Platform Simultaneous Notifications

```yaml
# alertmanager.yml - Multi-Platform Setup
route:
  receiver: "multi-platform-notifications"

receivers:
  - name: "multi-platform-notifications"
    webhook_configs:
      # Telegram notification
      - url: "http://localhost:9999/api/v1/telegram/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
      # Slack notification
      - url: "http://localhost:9999/api/v1/slack/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
      # Discord notification
      - url: "http://localhost:9999/api/v1/discord/chatid_L2"
        http_config:
          basic_auth:
            username: "admin"
            password: "admin"
```

### 📄 AlertManager Webhook Sample

The `raw_alertmanager.json` file in the project root provides a complete Prometheus AlertManager webhook payload sample, including:

#### 🔍 Sample Content Description

```json
{
  "receiver": "test-telegram-webhook", // Receiver name
  "status": "firing", // Alert status: firing/resolved
  "alerts": [
    // Alert array
    {
      "status": "firing", // Individual alert status
      "labels": {
        // Alert labels
        "alertname": "TEST_Pod_CPU_Usage High", // Alert name
        "env": "uat", // Environment label
        "namespace": "hcgateconsole", // Kubernetes namespace
        "pod": "hcgateconsole-deploy-xxx", // Pod name
        "severity": "test-alert" // Severity level
      },
      "annotations": {
        // Alert annotations
        "summary": "Pod CPU usage exceeds 80%" // Alert summary
      },
      "startsAt": "2022-11-18T08:17:31.745Z", // Alert start time
      "endsAt": "0001-01-01T00:00:00Z", // Alert end time (empty when firing)
      "generatorURL": "http://prometheus...", // Prometheus query link
      "fingerprint": "2da0690c63cf9cd3" // Alert fingerprint ID
    }
  ],
  "groupLabels": {
    // Group labels
    "alertname": "TEST_Pod_CPU_Usage High",
    "env": "uat",
    "severity": "test-alert"
  },
  "commonLabels": {
    // Common labels
    "alertname": "TEST_Pod_CPU_Usage High",
    "env": "uat",
    "severity": "test-alert"
  },
  "commonAnnotations": {}, // Common annotations
  "externalURL": "http://prometheus-alertmanager:9093", // AlertManager external URL
  "version": "4", // AlertManager version
  "groupKey": "...", // Group identification key
  "truncatedAlerts": 0 // Number of truncated alerts
}
```

#### 🎯 Use Cases

- **Development Testing**: Test webhook endpoint functionality
- **Template Development**: Reference data for developing custom alert templates
- **Debug Analysis**: Analyze webhook structure sent by AlertManager
- **Documentation Reference**: Understand complete AlertManager webhook payload format

#### 📊 Included Alert Types

The sample file contains alerts in two states:

1. **🔥 Firing Alerts** (3 alerts)

   - Active alerts for Pod CPU usage exceeding 80%
   - CPU high usage alerts for different Pods

2. **✅ Resolved Alerts** (1 alert)
   - Resolved CPU usage alerts
   - Includes complete start and end times

#### 🧪 Testing Usage

```bash
# Test webhook endpoint using curl
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L4 \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic $(echo -n admin:admin | base64)" \
  -d @raw_alertmanager.json

# Test other platforms
curl -X POST http://localhost:9999/api/v1/slack/chatid_L4 \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic $(echo -n admin:admin | base64)" \
  -d @raw_alertmanager.json

curl -X POST http://localhost:9999/api/v1/discord/chatid_L4 \
  -H "Content-Type: application/json" \
  -H "Authorization: Basic $(echo -n admin:admin | base64)" \
  -d @raw_alertmanager.json
```

## 📁 Project Structure

```text
alert-webhooks/
├── cmd/                           # Main program entry
│   └── main.go                   # Application entry point
├── config/                        # Configuration management modules
│   ├── app.go                    # Application configuration
│   ├── config.go                 # Main configuration file
│   ├── discord.go                # Discord configuration
│   ├── logger.go                 # Logger configuration
│   ├── manager.go                # Configuration manager
│   ├── metric.go                 # Monitoring metrics configuration
│   ├── slack.go                  # Slack configuration
│   ├── telgram.go                # Telegram configuration
│   └── webhooks.go               # Webhook configuration
├── configs/                       # Configuration file directory
│   ├── alert_config.minimal.yaml # Minimal alert configuration
│   ├── alert_config.yaml         # Complete alert configuration
│   └── config.development.yaml   # Development environment configuration
├── docs/                          # Swagger API documentation
│   ├── docs.go                   # Documentation generator
│   ├── swagger.json              # JSON format API documentation
│   └── swagger.yaml              # YAML format API documentation
├── documentation/                 # 📖 Project documentation
│   ├── en/                       # English documentation
│   │   ├── config_guide.md       # Configuration guide
│   │   ├── discord_setup.md      # Discord setup
│   │   ├── discord_usage.md      # Discord usage examples
│   │   ├── slack_setup.md        # Slack setup
│   │   ├── slack_usage.md        # Slack usage examples
│   │   ├── telegram_setup.md     # Telegram setup
│   │   ├── telegram_usage.md     # Telegram usage examples
│   │   └── template_guide.md     # Template guide
│   └── zh/                       # Chinese documentation
│       ├── config_guide.md       # Configuration guide
│       ├── discord_setup.md      # Discord setup
│       ├── discord_usage.md      # Discord usage examples
│       ├── slack_setup.md        # Slack setup
│       ├── slack_usage.md        # Slack usage examples
│       ├── telegram_setup.md     # Telegram setup
│       ├── telegram_usage.md     # Telegram usage examples
│       └── template_guide.md     # Template guide
├── examples/                      # Usage examples
│   ├── config_usage.go           # Configuration usage example
│   └── config.expamle.yaml       # Configuration example file
├── pkg/                          # Core functionality packages
│   ├── logcore/                  # Logging core
│   │   └── core.go              # Logging core implementation
│   ├── logger/                   # Logging system
│   │   ├── logger.go            # Logger implementation
│   │   ├── middleware.go        # Logging middleware
│   │   └── utils.go             # Logging utility functions
│   ├── logutil/                  # Logging utilities
│   │   └── context.go           # Context logging utilities
│   ├── middleware/               # HTTP middleware
│   │   ├── basic_auth.go        # Basic authentication middleware
│   │   ├── cors.go              # CORS middleware
│   │   ├── logger.go            # Logging middleware
│   │   └── recovery.go          # Recovery middleware
│   ├── notification/             # Notification system
│   │   ├── manager.go           # Notification manager
│   │   ├── providers/           # Notification providers
│   │   │   ├── discord.go       # Discord notification implementation
│   │   │   ├── slack.go         # Slack notification implementation
│   │   │   └── telegram.go      # Telegram notification implementation
│   │   └── types/               # Notification type definitions
│   │       └── types.go         # Notification type structures
│   ├── service/                  # Business service layer
│   │   ├── discord.go           # Discord service
│   │   ├── service.go           # Common service interface
│   │   ├── slack.go             # Slack service
│   │   └── telegram.go          # Telegram service
│   ├── template/                 # Template engine
│   │   └── engine.go            # Template engine implementation
│   └── watcher/                  # File monitoring
│       └── config_watcher.go    # Configuration file monitor
├── routes/                       # API routing system
│   ├── api/                     # API routes
│   │   └── v1/                  # API v1 version
│   │       ├── discord/         # Discord API routes
│   │       │   ├── handler.go   # Discord handler
│   │       │   └── routes.go    # Discord route definitions
│   │       ├── slack/           # Slack API routes
│   │       │   ├── handler.go   # Slack handler
│   │       │   └── routes.go    # Slack route definitions
│   │       ├── telegram/        # Telegram API routes
│   │       │   ├── handler.go   # Telegram handler
│   │       │   └── routes.go    # Telegram route definitions
│   │       ├── healthCheck.go   # Health check endpoint
│   │       └── register.go      # Route registrar
│   └── mainRoute.go             # Main route configuration
├── scripts/                      # Utility scripts
│   ├── fix_swagger_docs.go      # Swagger documentation fix script
│   └── regenerate_swagger.sh    # Swagger regeneration script
├── templates/                    # Message templates
│   └── alerts/                  # Alert templates
│       ├── alert_template_eng.tmpl  # English alert template
│       ├── alert_template_ja.tmpl   # Japanese alert template
│       ├── alert_template_ko.tmpl   # Korean alert template
│       ├── alert_template_tw.tmpl   # Traditional Chinese alert template
│       └── alert_template_zh.tmpl   # Simplified Chinese alert template
├── kubernetes/                   # Kubernetes deployment configuration
│   └── deployment-example.yaml  # Deployment example configuration
├── docker-compose.yml           # Docker Compose configuration
├── docker-compose.dev.yml       # Development environment Docker Compose
├── Dockerfile                   # Docker image build file
├── Makefile                     # Build and management scripts
├── go.mod                       # Go module dependencies
├── go.sum                       # Go module checksums
├── raw_alertmanager.json        # AlertManager webhook payload sample
├── README.md                    # English project documentation
└── README-zh.md                 # Chinese project documentation
```

## 🌍 Languages

- [English](./README.md) (Current)
- [繁體中文](./README-zh.md)

## 🤝 Contributing

Issues and Pull Requests are welcome to improve this project!

## 📄 License

This project is licensed under the MIT License.
