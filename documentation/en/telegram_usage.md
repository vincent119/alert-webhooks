# Telegram API Usage Examples

## Overview

This project integrates the [go-telegram/bot](https://github.com/go-telegram/bot) framework to send Telegram messages.

## Configuration

### 1. Set Telegram Bot Token

Configure your Telegram Bot Token in the configuration file:

```yaml
# configs/config.development.yaml
app:
  telegram_token: "YOUR_TELEGRAM_BOT_TOKEN_HERE"
```

### 2. Set Chat ID

Configure the corresponding chat_id in `configs/config.{env}.yaml`:

```go
  chat_ids0: "-1002465088995" # information group (level 0)
  chat_ids1: "-1002465088995" # general message group (level 1)
  chat_ids2: "-1002465088995" # critical notification group (level 2)
  chat_ids3: "-1002465088995" # emergency alert group (level 3)
  chat_ids4: "-1002465088995" # testing group (level 4)
  chat_ids5: "-1002465088995" # backup group (level 5)
  chat_ids6: "" # backup group (level 6)
```

## API Endpoints

### 1. Send Message

**Endpoint:** `POST /api/v1/telegram/chatid_L{level}`

**Parameters:**

- `level`: Chat level (0-5)

**Request Example:**

```bash
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L0 \
  -H "Content-Type: application/json" \
  -d '{
    "message": "This is a test message"
  }'
```

**Response Example:**

```json
{
  "success": true,
  "message": "Message sent successfully",
  "level": 0
}
```

### 2. Get Bot Information

**Endpoint:** `GET /api/v1/telegram/info`

**Request Example:**

```bash
curl http://localhost:9999/api/v1/telegram/info
```

**Response Example:**

```json
{
  "success": true,
  "bot_info": {
    "id": 123456789,
    "username": "your_bot_username",
    "first_name": "Your Bot Name",
    "can_join_groups": true,
    "can_read_all_group_messages": false
  }
}
```

## Usage Examples

### 1. Send Messages at Different Levels

```bash
# Send to level 0
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
  http://localhost:9999/api/v1/telegram/chatid_L0

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

### 2. Error Handling

**Invalid chatid format:**

```bash
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L5 \
  -H "Content-Type: application/json" \
  -d '{"message": "test"}'
```

**Response:**

```json
{
  "success": false,
  "message": "Invalid chatid format. Must be L0, L1, L2, L3, or L4"
}
```

**Empty message:**

```bash
curl -X POST http://localhost:9999/api/v1/telegram/chatid_L0 \
  -H "Content-Type: application/json" \
  -d '{"message": ""}'
```

**Response:**

```json
{
  "success": false,
  "message": "Message cannot be empty"
}
```

## Logging

The system logs all Telegram-related operations:

### Successful Message Sending

```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123+08:00",
  "msg": "Telegram message sent successfully",
  "category": "telegram",
  "level": 0,
  "chat_id": -1001234567890,
  "message": "Test message"
}
```

### Sending Failed

```json
{
  "level": "error",
  "ts": "2024-01-15T10:30:45.123+08:00",
  "msg": "Failed to send Telegram message",
  "category": "telegram",
  "level": 0,
  "chat_id": -1001234567890,
  "message": "Test message",
  "error": "Forbidden: bot was blocked by the user"
}
```

## Notes

1. **Token Security**: Ensure your Telegram Bot Token is not exposed in public code
2. **Chat ID**: Ensure the configured chat_id is correct and the bot has permission to send messages
3. **Message Format**: Supports plain text messages, use Markdown or HTML for richer formatting
4. **Rate Limiting**: Telegram API has rate limits, avoid sending messages too frequently

## Extended Features

You can extend the following features based on your needs:

1. **Message Format**: Support for Markdown, HTML formats
2. **File Sending**: Support for sending images, documents, etc.
3. **Buttons**: Support for inline keyboard buttons
4. **Group Management**: Support for group management features
5. **Message Editing**: Support for editing sent messages
