# Telegram Bot Setup Guide

This guide walks you through setting up a Telegram bot for use with the Alert Webhooks service.

## 🤖 Creating a Telegram Bot

### Step 1: Contact BotFather

1. Open Telegram and search for [@BotFather](https://t.me/botfather)
2. Start a chat with BotFather by clicking "Start" or sending `/start`
3. Send the command `/newbot` to create a new bot

### Step 2: Configure Your Bot

1. **Choose a name** for your bot (e.g., "Alert Webhook Bot")
2. **Choose a username** for your bot (must end with "bot", e.g., "alertwebhook_bot")
3. BotFather will provide you with a **bot token** - save this securely!

Example token format: `1234567890:ABCdefGHIjklMNOpqrSTUvwxYZ123456789`

### Step 3: Configure Bot Settings (Optional)

You can configure additional bot settings:

```
/setdescription - Set bot description
/setabouttext - Set bot about text
/setuserpic - Set bot profile picture
/setcommands - Set bot commands menu
```

## 🆔 Getting Chat IDs

### For Private Chats

1. Send a message to your bot
2. Open this URL in your browser (replace `<YOUR_BOT_TOKEN>`):
   ```
   https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates
   ```
3. Look for the `chat.id` value in the response

Example response:
```json
{
  "result": [
    {
      "message": {
        "chat": {
          "id": 123456789,
          "type": "private"
        }
      }
    }
  ]
}
```

### For Group Chats

1. **Add your bot to the group**:
   - Go to the group settings
   - Select "Add Members"
   - Search for your bot username
   - Add the bot

2. **Send a message mentioning the bot**:
   ```
   @your_bot_name Hello!
   ```

3. **Get the chat ID**:
   - Open: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
   - Look for the `chat.id` value (will be negative for groups)

Example group chat ID: `-1001234567890`

### For Supergroups

Supergroups use the same process as regular groups but have IDs starting with `-100`:

Example supergroup chat ID: `-1001234567890`

## 🔧 Configuration

### Update Configuration File

Edit your `configs/config.development.yaml` file:

```yaml
telegram:
  enable: true
  token: "YOUR_BOT_TOKEN_HERE"
  chat_ids0: "EMERGENCY_GROUP_ID"     # Level 0 - Emergency
  chat_ids1: "CRITICAL_GROUP_ID"      # Level 1 - Critical
  chat_ids2: "IMPORTANT_GROUP_ID"     # Level 2 - Important
  chat_ids3: "GENERAL_GROUP_ID"       # Level 3 - General
  chat_ids4: "INFO_GROUP_ID"          # Level 4 - Info
  chat_ids5: "TEST_GROUP_ID"          # Level 5 - Test
  chat_ids6: ""                       # Level 6 - Reserved
  template_mode: "full"               # full, minimal
  template_language: "en"             # en, tw, zh, ja, ko
```

### Chat Level Configuration

Configure different chat groups for different alert levels:

| Level | Purpose | Recommended Use |
|-------|---------|-----------------|
| L0 | Emergency | Critical system failures |
| L1 | Critical | High priority alerts |
| L2 | Important | Medium priority alerts |
| L3 | General | Standard notifications |
| L4 | Info | Informational messages |
| L5 | Test | Testing and development |
| L6 | Reserved | Future use |

## 🧪 Testing Your Setup

### Test Bot Communication

1. **Start the service**:
   ```bash
   go run cmd/main.go -e development
   ```

2. **Test the bot info endpoint**:
   ```bash
   curl -u admin:admin http://localhost:9999/api/v1/telegram/info
   ```

3. **Send a test message**:
   ```bash
   curl -X POST \
     -u admin:admin \
     -H "Content-Type: application/json" \
     -d '{
       "message": "Test message from Alert Webhooks!",
       "level": 5
     }' \
     http://localhost:9999/api/v1/telegram/chatid_L5
   ```

### Test AlertManager Integration

Send a test AlertManager webhook:

```bash
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "telegram-test",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {
          "alertname": "TestAlert",
          "severity": "warning"
        },
        "annotations": {
          "summary": "This is a test alert"
        },
        "startsAt": "2023-01-01T00:00:00.000Z"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

## 🔐 Security Best Practices

### Bot Token Security

1. **Never commit bot tokens** to version control
2. **Use environment variables** in production:
   ```bash
   export TELEGRAM_BOT_TOKEN="your_token_here"
   ```
3. **Regenerate tokens** if compromised via BotFather

### Bot Permissions

1. **Limit bot permissions** in groups:
   - Remove unnecessary admin rights
   - Only grant message sending permissions

2. **Use private groups** for sensitive alerts

3. **Monitor bot usage** through Telegram's bot analytics

### Network Security

1. **Use HTTPS** in production
2. **Implement rate limiting** for webhook endpoints
3. **Validate webhook payloads** before processing

## ⚠️ Troubleshooting

### Common Issues

**"Bot not found" error**:
- Verify bot token is correct
- Check if bot was created successfully via BotFather

**"Chat not found" error**:
- Verify chat ID format (negative for groups)
- Ensure bot is added to the group
- Check if group still exists

**"Forbidden" error**:
- Bot may have been removed from the group
- User may have blocked the bot
- Check bot permissions in the group

**Messages not being delivered**:
- Verify chat ID is correct
- Check if group has message restrictions
- Ensure bot has permission to send messages

### Debug Mode

Enable debug logging to troubleshoot issues:

```yaml
log:
  level: "debug"
```

### Telegram API Limits

Be aware of Telegram's rate limits:
- **30 messages per second** per bot
- **1 message per second** per chat for groups
- **20 messages per minute** for the same user in private chats

## 📝 Bot Commands (Optional)

You can set up bot commands for interactive use:

1. Send `/setcommands` to BotFather
2. Set up commands like:
   ```
   help - Show help information
   status - Check bot status
   test - Send test message
   ```

3. Implement command handlers in your application

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)
