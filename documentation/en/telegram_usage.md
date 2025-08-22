# Telegram API Usage Examples

This guide provides practical examples of using the Telegram API endpoints in the Alert Webhooks service.

## 📡 API Endpoints

### 1. Send Message Endpoint

**Endpoint**: `POST /api/v1/telegram/chatid_{level}`  
**Authentication**: HTTP Basic Auth  
**Content-Type**: `application/json`

### 2. Bot Info Endpoint

**Endpoint**: `GET /api/v1/telegram/info`  
**Authentication**: HTTP Basic Auth

## 🔐 Authentication

All Telegram endpoints require HTTP Basic Authentication:

```bash
# Using curl
curl -u username:password [endpoint]

# Using credentials from config
curl -u admin:admin [endpoint]
```

## 💬 Sending Messages

### Simple Text Message

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Hello from Alert Webhooks!",
    "level": 5
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

**Response**:
```json
{
  "success": true,
  "message": "Message sent successfully"
}
```

### AlertManager Webhook Format

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "telegram-alerts",
    "status": "firing",
    "groupLabels": {
      "alertname": "HighCPUUsage",
      "env": "production",
      "severity": "warning"
    },
    "commonAnnotations": {
      "summary": "High CPU usage detected"
    },
    "externalURL": "https://alertmanager.example.com",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "HighCPUUsage",
          "env": "production",
          "severity": "warning",
          "instance": "server-01"
        },
        "annotations": {
          "summary": "CPU usage above 80% for 5 minutes",
          "description": "Server experiencing high load"
        },
        "startsAt": "2023-01-01T10:30:00.000Z",
        "generatorURL": "https://prometheus.example.com/graph?expr=cpu_usage"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L1
```

### Multi-Language Support

```bash
# English message
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "alertmanager_data": {
      "status": "firing",
      "alerts": [...]
    },
    "template_language": "en"
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L0

# Traditional Chinese message
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "alertmanager_data": {
      "status": "firing", 
      "alerts": [...]
    },
    "template_language": "tw"
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L0
```

## 🤖 Bot Information

### Get Bot Details

```bash
curl -u admin:admin http://localhost:9999/api/v1/telegram/info
```

**Response**:
```json
{
  "success": true,
  "user": {
    "id": 1234567890,
    "is_bot": true,
    "first_name": "Alert Webhook Bot",
    "username": "alertwebhook_bot",
    "can_join_groups": true,
    "can_read_all_group_messages": false,
    "supports_inline_queries": false
  }
}
```

## 📊 Chat Level Configuration

### Chat Levels

| Level | Purpose | Example Use Case |
|-------|---------|------------------|
| L0 | Emergency | System down, data loss |
| L1 | Critical | Service unavailable |
| L2 | Important | Performance issues |
| L3 | General | Regular monitoring |
| L4 | Info | System updates |
| L5 | Test | Development testing |
| L6 | Reserved | Future use |

### Level-Specific Examples

#### Emergency Alert (L0)

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "alertmanager_data": {
      "receiver": "emergency",
      "status": "firing",
      "groupLabels": {
        "alertname": "SystemDown",
        "env": "production",
        "severity": "critical"
      },
      "alerts": [
        {
          "status": "firing",
          "labels": {
            "alertname": "SystemDown",
            "severity": "critical"
          },
          "annotations": {
            "summary": "Production system is completely down"
          },
          "startsAt": "2023-01-01T10:00:00.000Z"
        }
      ]
    }
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L0
```

#### Test Alert (L5)

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "message": "This is a test message",
    "level": 5
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

## 🔄 Resolved Alerts

### Send Resolution Notification

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "resolved-alerts",
    "status": "resolved",
    "groupLabels": {
      "alertname": "HighCPUUsage",
      "env": "production"
    },
    "alerts": [
      {
        "status": "resolved",
        "labels": {
          "alertname": "HighCPUUsage",
          "env": "production",
          "instance": "server-01"
        },
        "annotations": {
          "summary": "CPU usage returned to normal"
        },
        "startsAt": "2023-01-01T10:30:00.000Z",
        "endsAt": "2023-01-01T10:45:00.000Z"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L2
```

## 🧪 Testing and Development

### Health Check Integration

Before sending alerts, verify the service is running:

```bash
# Check service health
curl http://localhost:9999/api/v1/healthz

# Expected response
{
  "status": "ok",
  "version": "1.0.0"
}
```

### Template Testing

Test different template languages and modes:

```bash
# Test English template
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "alertmanager_data": {...},
    "template_language": "en"
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5

# Test Chinese template  
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "alertmanager_data": {...},
    "template_language": "tw"
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

### Error Handling Examples

#### Authentication Failure

```bash
curl -X POST \
  -u wrong:credentials \
  -H "Content-Type: application/json" \
  -d '{"message": "test"}' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

**Response** (401):
```json
{
  "success": false,
  "message": "Authentication failed: Invalid credentials provided"
}
```

#### Invalid Chat Level

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{"message": "test"}' \
  http://localhost:9999/api/v1/telegram/chatid_L9
```

**Response** (400):
```json
{
  "success": false,
  "message": "Invalid chat level: L9"
}
```

#### Missing Message Content

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{}' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

**Response** (400):
```json
{
  "success": false,
  "message": "Either message or alertmanager_data must be provided"
}
```

## 🔧 Integration Examples

### AlertManager Configuration

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'telegram-alerts'

receivers:
- name: 'telegram-alerts'
  webhook_configs:
  - url: 'http://localhost:9999/api/v1/telegram/chatid_L1'
    http_config:
      basic_auth:
        username: 'admin'
        password: 'admin'
    send_resolved: true

inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  equal: ['alertname', 'dev', 'instance']
```

### Prometheus Alert Rules

```yaml
# alerts.yml
groups:
- name: system
  rules:
  - alert: HighCPUUsage
    expr: 100 - (avg by(instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 80
    for: 5m
    labels:
      severity: warning
      env: production
    annotations:
      summary: "High CPU usage on {{ $labels.instance }}"
      description: "CPU usage is above 80% for more than 5 minutes"

  - alert: SystemDown
    expr: up == 0
    for: 1m
    labels:
      severity: critical
      env: production
    annotations:
      summary: "System {{ $labels.instance }} is down"
      description: "System has been down for more than 1 minute"
```

### Script Integration

```bash
#!/bin/bash
# send_telegram_alert.sh

WEBHOOK_URL="http://localhost:9999/api/v1/telegram/chatid_L2"
USERNAME="admin"
PASSWORD="admin"

# Function to send alert
send_alert() {
  local message="$1"
  local level="${2:-2}"
  
  curl -X POST \
    -u "$USERNAME:$PASSWORD" \
    -H "Content-Type: application/json" \
    -d "{\"message\": \"$message\", \"level\": $level}" \
    "$WEBHOOK_URL" \
    -s | jq .
}

# Usage examples
send_alert "Backup completed successfully" 4
send_alert "Disk space low on server" 2
send_alert "Database connection failed" 1
```

### Python Integration

```python
#!/usr/bin/env python3
import requests
import json
from datetime import datetime

class TelegramAlerter:
    def __init__(self, base_url, username, password):
        self.base_url = base_url
        self.auth = (username, password)
    
    def send_simple_alert(self, message, level=5):
        """Send a simple text message"""
        url = f"{self.base_url}/telegram/chatid_L{level}"
        data = {
            "message": message,
            "level": level
        }
        
        response = requests.post(url, json=data, auth=self.auth)
        return response.json()
    
    def send_alertmanager_alert(self, alertname, severity, summary, level=2):
        """Send AlertManager format alert"""
        url = f"{self.base_url}/telegram/chatid_L{level}"
        data = {
            "receiver": "python-client",
            "status": "firing",
            "groupLabels": {
                "alertname": alertname,
                "severity": severity
            },
            "alerts": [
                {
                    "status": "firing",
                    "labels": {
                        "alertname": alertname,
                        "severity": severity
                    },
                    "annotations": {
                        "summary": summary
                    },
                    "startsAt": datetime.utcnow().isoformat() + "Z"
                }
            ]
        }
        
        response = requests.post(url, json=data, auth=self.auth)
        return response.json()

# Usage
alerter = TelegramAlerter("http://localhost:9999/api/v1", "admin", "admin")

# Send simple alert
result = alerter.send_simple_alert("Python test message", 5)
print(result)

# Send AlertManager format alert
result = alerter.send_alertmanager_alert(
    "PythonAlert", 
    "warning", 
    "Test alert from Python script",
    3
)
print(result)
```

## 📈 Monitoring and Logging

### Enable Debug Logging

To troubleshoot API calls, enable debug logging:

```yaml
# config.development.yaml
log:
  level: "debug"
```

### Monitor API Calls

Watch logs for API activity:
```bash
# Follow logs
tail -f logs/app.log | grep "telegram"

# Or if logging to stdout
go run cmd/main.go -e development | grep "telegram"
```

### Prometheus Metrics

The service exposes metrics at `/api/v1/metrics`:

```bash
curl -u metric_user:metric_password http://localhost:9999/api/v1/metrics
```

Key metrics:
- `telegram_messages_sent_total` - Total messages sent
- `telegram_errors_total` - Total errors
- `http_requests_total` - HTTP request counts
- `http_request_duration_seconds` - Request latency

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)
