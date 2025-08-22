# Discord Usage Guide

This document explains how to use Alert Webhooks Discord API endpoints to send notifications.

## 📋 Table of Contents
- [API Endpoints](#api-endpoints)
- [Usage Examples](#usage-examples)
- [Level Routing](#level-routing)
- [Template System](#template-system)
- [Error Handling](#error-handling)
- [AlertManager Integration](#alertmanager-integration)

## 🔗 API Endpoints

### Basic Endpoint Format
```
POST /api/v1/discord/channel/{channel_id}    - Send to specific channel
POST /api/v1/discord/chatid_L{level}         - Send to specific level
GET  /api/v1/discord/status                  - Get service status
POST /api/v1/discord/test/{channel_id}       - Test channel connection
POST /api/v1/discord/validate/{channel_id}   - Validate channel permissions
```

### Authentication
All API endpoints require Basic Authentication:
```bash
-u "username:password"
```

## 💡 Usage Examples

### 1. Simple Text Message

#### Send to Specific Channel
```bash
curl -X POST "http://localhost:9999/api/v1/discord/channel/987654321098765432" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{
    "message": "🚨 System Alert: Database connection failed"
  }'
```

#### Send to Level 0 (Critical)
```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L0" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{
    "message": "🔥 URGENT: Production service outage"
  }'
```

### 2. AlertManager Format Message

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L1" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "instance": "web-server-01",
          "severity": "warning",
          "env": "production"
        },
        "annotations": {
          "summary": "High CPU usage detected",
          "description": "CPU usage has exceeded 80%"
        },
        "startsAt": "2024-01-15T10:30:00Z"
      }
    ],
    "status": "firing",
    "externalURL": "http://alertmanager.example.com"
  }'
```

### 3. Service Status Check

```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin"
```

**Response Example:**
```json
{
  "service": "discord",
  "status": "healthy",
  "bot_info": {
    "id": "123456789012345678",
    "username": "Alert Webhooks Bot",
    "bot": true,
    "avatar": "abc123def456"
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### 4. Channel Validation

```bash
curl -X POST "http://localhost:9999/api/v1/discord/validate/987654321098765432" \
  -u "admin:admin"
```

### 5. Test Message

```bash
curl -X POST "http://localhost:9999/api/v1/discord/test/987654321098765432" \
  -u "admin:admin"
```

## 📊 Level Routing

### Level Mapping Table

| Level | API Endpoint | Channel Type | Config Key | Description |
|-------|--------------|--------------|------------|-------------|
| L0    | `/chatid_L0` | 🚨 Critical | `chat_ids0` | Critical alerts |
| L1    | `/chatid_L1` | ⚠️ High | `chat_ids1` | High priority |
| L2    | `/chatid_L2` | 📢 Normal | `chat_ids2` | Normal alerts |
| L3    | `/chatid_L3` | 📝 Info | `chat_ids3` | Info notifications |
| L4    | `/chatid_L4` | 🔧 Debug | `chat_ids4` | Debug messages |
| L5    | `/chatid_L5` | 📦 Backup | `chat_ids5` | Backup channel |

### Usage Recommendations

#### 🚨 Level 0 - Critical
- System failures
- Complete service outages
- Security incidents
- Requires immediate human intervention

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L0" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "🚨 **CRITICAL ALERT**: Primary database connection failed"}'
```

#### ⚠️ Level 1 - High Priority
- Performance degradation
- Partial service issues
- Error rate increases

```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L1" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "⚠️ **HIGH PRIORITY**: API response time anomaly detected"}'
```

#### 📢 Level 2 - Normal
- General monitoring warnings
- Resource usage alerts
- Configuration change notifications

## 🎨 Template System

### Template Configuration
Discord uses the same template system as Telegram/Slack:

```yaml
discord:
  template_mode: "full"        # minimal, full
  template_language: "eng"     # eng, tw, zh, ja, ko
```

### Supported Formatting

Discord supports standard Markdown formatting:

- **Bold text**: `**text**`
- *Italic text*: `*text*`
- `Inline code`: `` `code` ``
- Code blocks: ``` ```code block``` ```
- [Links](URL): `[Link text](URL)`

### Template Variables
Templates can use the following variables:

- `{{.alerts}}` - Alert array
- `{{.status}}` - Alert status
- `{{.externalURL}}` - External URL
- `{{.alertname}}` - Alert name
- `{{.env}}` - Environment
- `{{.severity}}` - Severity
- `{{.namespace}}` - Namespace

## 🚨 Error Handling

### Common Error Responses

#### 1. Missing Permissions
```json
{
  "success": false,
  "message": "bot lacks necessary permissions in channel 987654321098765432. Please ensure the bot has 'Send Messages' permission"
}
```

#### 2. Channel Not Found
```json
{
  "success": false,
  "message": "channel 987654321098765432 does not exist or bot cannot access it"
}
```

#### 3. Invalid Token
```json
{
  "success": false,
  "message": "invalid Discord token. Please check if the token in configuration is correct"
}
```

#### 4. Invalid Message Content
```json
{
  "success": false,
  "message": "message content is invalid or too long"
}
```

### Message Length Limits
- Discord message maximum length: **2000 characters**
- When exceeded, system automatically splits into multiple messages
- Splitting occurs at newlines to maintain formatting

## 🔗 AlertManager Integration

### Webhook Configuration
Configure in AlertManager's `alertmanager.yml`:

```yaml
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'discord-notifications'
  routes:
  - match:
      severity: critical
    receiver: 'discord-critical'
  - match:
      severity: warning
    receiver: 'discord-warning'

receivers:
- name: 'discord-critical'
  webhook_configs:
  - url: 'http://alert-webhooks:9999/api/v1/discord/chatid_L0'
    http_config:
      basic_auth:
        username: 'admin'
        password: 'admin'

- name: 'discord-warning'
  webhook_configs:
  - url: 'http://alert-webhooks:9999/api/v1/discord/chatid_L1'
    http_config:
      basic_auth:
        username: 'admin'
        password: 'admin'

- name: 'discord-notifications'
  webhook_configs:
  - url: 'http://alert-webhooks:9999/api/v1/discord/chatid_L2'
    http_config:
      basic_auth:
        username: 'admin'
        password: 'admin'
```

### Prometheus Rules Example

```yaml
groups:
- name: discord.rules
  rules:
  - alert: HighCPUUsage
    expr: cpu_usage_percent > 80
    for: 5m
    labels:
      severity: warning
      env: production
    annotations:
      summary: "High CPU usage detected"
      description: "Instance {{ $labels.instance }} CPU usage is {{ $value }}%"

  - alert: ServiceDown
    expr: up == 0
    for: 1m
    labels:
      severity: critical
      env: production
    annotations:
      summary: "Service is down"
      description: "Service {{ $labels.job }} on instance {{ $labels.instance }} is down"
```

## 📊 Monitoring and Logs

### Check Discord Service Status
```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin" | jq .
```

### View Discord-related Logs
```bash
grep "Discord" ./logs/server.log | tail -20
```

### Test End-to-End Flow
1. Check service status
2. Validate channel permissions
3. Send test message
4. Check Discord channel

```bash
# Complete test flow
curl -X GET "http://localhost:9999/api/v1/discord/status" -u "admin:admin"
curl -X POST "http://localhost:9999/api/v1/discord/validate/your-channel-id" -u "admin:admin"
curl -X POST "http://localhost:9999/api/v1/discord/test/your-channel-id" -u "admin:admin"
```

## 🔧 Advanced Configuration

### Role Mentions
Automatically mention specific roles in messages:

```yaml
discord:
  mention_roles:
    - "role-id-for-ops-team"
    - "role-id-for-on-call"
```

### Custom Message Formatting
Customize message format through template system, supporting:
- Markdown formatting
- Emojis
- Custom text and layout
- Multi-language support

## 📚 Related Documentation
- [Discord Setup Guide](discord_setup.md)
- [Template System Documentation](template-system.md)
- [Kubernetes Environment Variables](kubernetes-env-vars.md)
- [Troubleshooting Guide](troubleshooting.md)

## 💡 Best Practices
1. **Use Appropriate Levels** - Choose suitable levels based on alert severity
2. **Configure Role Mentions** - Set up role mentions for critical alerts
3. **Test Configuration** - Regularly test Discord integration
4. **Monitor Logs** - Regularly check service logs
5. **Backup Channels** - Configure backup channels for failover
