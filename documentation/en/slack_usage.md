# Slack Usage Guide

This guide explains how to use Slack for alert notifications in the Alert-Webhooks system.

## API Endpoints Overview

Alert-Webhooks provides the following Slack API endpoints:

| Endpoint                           | Method | Description                                |
| ---------------------------------- | ------ | ------------------------------------------ |
| `/api/v1/slack/channel/{channel}`  | POST   | Send message to specific channel           |
| `/api/v1/slack/chatid_L{level}`    | POST   | Send message to specific level channel     |
| `/api/v1/slack/rich/{channel}`     | POST   | Send rich text message to specific channel |
| `/api/v1/slack/status`             | GET    | Get Slack service status                   |
| `/api/v1/slack/channels`           | GET    | Get configured channel list                |
| `/api/v1/slack/test`               | POST   | Test Slack connection                      |
| `/api/v1/slack/validate/{channel}` | GET    | Validate if Bot is in specified channel    |

## Authentication

All API endpoints require HTTP Basic authentication:

- **Username**: `config.webhooks.base_auth_user`
- **Password**: `config.webhooks.base_auth_password`

## Usage Examples

### 1. Send Simple Message

#### Send to specific channel

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{"message": "System maintenance notice: Maintenance will be performed tonight at 10 PM"}' \
  "http://localhost:9999/api/v1/slack/channel/general"
```

#### Send to level channel

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{"message": "Critical alert: Database connection failed"}' \
  "http://localhost:9999/api/v1/slack/chatid_L0"
```

### 2. Send AlertManager Alerts

#### Using AlertManager data format

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{
    "receiver": "web.hook",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "instance": "server-01",
          "severity": "warning"
        },
        "annotations": {
          "summary": "High CPU usage",
          "description": "CPU usage on server-01 has exceeded 80%"
        },
        "startsAt": "2024-01-15T10:30:00Z",
        "generatorURL": "http://prometheus:9090/graph?g0.expr=cpu_usage"
      }
    ],
    "groupLabels": {"alertname": "HighCPUUsage"},
    "commonLabels": {"alertname": "HighCPUUsage", "severity": "warning"},
    "externalURL": "http://alertmanager:9093"
  }' \
  "http://localhost:9999/api/v1/slack/chatid_L1"
```

### 3. Send Rich Text Message

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -u admin:admin \
  -d '{
    "blocks": [
      {
        "type": "header",
        "text": {
          "type": "plain_text",
          "text": "🚨 System Alert"
        }
      },
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*Status*: Firing\n*Severity*: High\n*Service*: Web API"
        }
      },
      {
        "type": "actions",
        "elements": [
          {
            "type": "button",
            "text": {
              "type": "plain_text",
              "text": "View Details"
            },
            "url": "http://monitoring.example.com/alerts"
          }
        ]
      }
    ]
  }' \
  "http://localhost:9999/api/v1/slack/rich/alerts"
```

### 4. Service Status Check

#### Get Slack service status

```bash
curl -u admin:admin "http://localhost:9999/api/v1/slack/status"
```

Response example:

```json
{
  "success": true,
  "service": "slack",
  "status": "active",
  "config": {
    "enabled": true,
    "default_channel": "#alerts",
    "bot_username": "Alert Bot"
  },
  "channels": {
    "L0": "#critical-alerts",
    "L1": "#warning-alerts",
    "L2": "#info-alerts"
  }
}
```

#### Get channel configuration

```bash
curl -u admin:admin "http://localhost:9999/api/v1/slack/channels"
```

#### Test connection

```bash
curl -X POST \
  -u admin:admin \
  "http://localhost:9999/api/v1/slack/test"
```

#### Validate channel

```bash
curl -u admin:admin "http://localhost:9999/api/v1/slack/validate/alerts"
```

## Level Routing System

### Level Configuration

The system supports 6 levels (L0-L5), each corresponding to different channels:

| Level | Route        | Suggested Use                                  | Config Key  |
| ----- | ------------ | ---------------------------------------------- | ----------- |
| L0    | `/chatid_L0` | Critical alerts, system outages                | `chat_ids0` |
| L1    | `/chatid_L1` | Important warnings, performance issues         | `chat_ids1` |
| L2    | `/chatid_L2` | General notifications, status changes          | `chat_ids2` |
| L3    | `/chatid_L3` | Information messages, deployment notifications | `chat_ids3` |
| L4    | `/chatid_L4` | Debug messages, test alerts                    | `chat_ids4` |
| L5    | `/chatid_L5` | Other, backup                                  | `chat_ids5` |

### Configuration Example

```yaml
slack:
  channels:
    chat_ids0: "#critical-alerts" # Critical alerts
    chat_ids1: "#warning-alerts" # Warnings
    chat_ids2: "#info-alerts" # Information
    chat_ids3: "#debug-alerts" # Debug
    chat_ids4: "#test-alerts" # Testing
    chat_ids5: "#other-alerts" # Other
```

## Message Format

### Template System

The system uses templates to format AlertManager alert messages. Templates are located at:

- `templates/alerts/alert_template_tw.tmpl` (Traditional Chinese) 🇹🇼
- `templates/alerts/alert_template_eng.tmpl` (English) 🇺🇸
- `templates/alerts/alert_template_zh.tmpl` (Simplified Chinese) 🇨🇳
- `templates/alerts/alert_template_ja.tmpl` (Japanese) 🇯🇵
- `templates/alerts/alert_template_ko.tmpl` (Korean) 🇰🇷

### Template Configuration

Set in configuration file:

```yaml
slack:
  template_mode: "full" # minimal or full
  template_language: "en" # eng, tw, zh, ja, ko
```

### Format Examples

#### Full mode message example:

```
🚨 Alert Notification

Status: firing
Alert Name: HighCPUUsage
Environment: production
Severity: warning
Namespace: default
Total Alerts: 1
Firing: 1

🚨 Firing Alerts:

Alert 1:
• Summary: High CPU usage
• Pod: web-server-01
• Started: 2024-01-15 10:30:00
• View Details: http://prometheus:9090/graph?g0.expr=cpu_usage

View All Alert Details: http://alertmanager:9093
```

#### Minimal mode message example:

```
🚨 HighCPUUsage - warning
High CPU usage (web-server-01)
```

## Error Handling

### Common Errors and Solutions

#### 1. Authentication Failed

```json
{
  "error": "Unauthorized",
  "message": "Invalid username or password",
  "code": 401
}
```

**Solution**: Check Basic Auth username and password

#### 2. Service Not Enabled

```json
{
  "success": false,
  "message": "Slack service is not enabled"
}
```

**Solution**: Set `slack.enable: true` in configuration file

#### 3. Channel Not Found

```json
{
  "success": false,
  "message": "Channel not found: #nonexistent"
}
```

**Solution**: Confirm channel name is correct, Bot has been invited to the channel

#### 4. Invalid Token

```json
{
  "success": false,
  "message": "Invalid token"
}
```

**Solution**: Check if Slack Bot Token is correctly configured

## AlertManager Integration

### AlertManager Configuration

Configure webhook in AlertManager's `alertmanager.yml`:

```yaml
global:
  # Global settings

route:
  group_by: ["alertname"]
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: "web.hook"
  routes:
    - match:
        severity: critical
      receiver: "slack-critical"
    - match:
        severity: warning
      receiver: "slack-warning"

receivers:
  - name: "web.hook"
    slack_configs:
      - send_resolved: true
        api_url: "http://alert-webhooks:9999/api/v1/slack/chatid_L2"
        username: "Alert Bot"
        channel: "#info-alerts"

  - name: "slack-critical"
    slack_configs:
      - send_resolved: true
        api_url: "http://alert-webhooks:9999/api/v1/slack/chatid_L0"
        username: "Alert Bot"
        channel: "#critical-alerts"

  - name: "slack-warning"
    slack_configs:
      - send_resolved: true
        api_url: "http://alert-webhooks:9999/api/v1/slack/chatid_L1"
        username: "Alert Bot"
        channel: "#warning-alerts"
```

### Prometheus Rules Example

```yaml
groups:
  - name: example
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 80
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage"
          description: "CPU usage on {{ $labels.instance }} has exceeded 80%"

      - alert: ServiceDown
        expr: up == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Service down"
          description: "{{ $labels.instance }} service has been down for more than 5 minutes"
```

## Performance and Limitations

### Message Limits

- Slack message length limit: 40,000 characters
- Rich text block limit: 50 blocks
- Attachment limit: 20 attachments

### Rate Limits

- Slack API has rate limits, recommendations:
  - No more than 1 request per second
  - Use batch processing for large number of alerts
  - Avoid sending duplicate messages

### Best Practices

1. Use appropriate level routing to avoid message flooding
2. Configure reasonable AlertManager grouping rules
3. Use template customization to reduce message redundancy
4. Regularly check and clean up inactive channel configurations

## Related Documentation

- [Slack Setup Guide](./slack_setup.md)
- [Service Enable Configuration Guide](./service-enable-config.md)
- [Template Usage Guide](./template_usage.md)
- [Kubernetes Environment Variables Configuration](./kubernetes-env-vars.md)
