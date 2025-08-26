# Configuration Guide

This guide provides detailed instructions for configuring the Alert Webhooks service.

## 📁 Configuration Files

The service uses YAML configuration files located in the `configs/` directory:

- `config.development.yaml` - Development environment configuration
- `config.production.yaml` - Production environment configuration  
- `config.test.yaml` - Test environment configuration
- `telegram_config.yaml` - Default template formatting options
- `telegram_config.minimal.yaml` - Minimal template formatting options

## 🔧 Main Configuration Structure

### Application Settings (`app`)

```yaml
app:
  name: "alert-webhooks"
  version: "1.0.0"
  mode: "development"    # development, production, test
  port: "9999"
  trusted_proxies: "127.0.0.1"
```

### Logging Configuration (`log`)

```yaml
log:
  level: "info"          # debug, info, warn, error
  format: "json"         # json, text
  output: "stdout"       # stdout, stderr, file path
```

### Webhook Authentication (`webhooks`)

```yaml
webhooks:
  enable: true           # Enable/disable webhook authentication
  username: "admin"      # Basic Auth username
  password: "admin"      # Basic Auth password
```

### Telegram Configuration (`telegram`)

```yaml
telegram:
  enable: true
  token: "YOUR_BOT_TOKEN"
  chat_ids0: "-1002465088995"    # Emergency alerts (Level 0)
  chat_ids1: "-1002465088995"    # Critical alerts (Level 1)
  chat_ids2: "-1002465088995"    # Important alerts (Level 2)
  chat_ids3: "-1002465088995"    # General alerts (Level 3)
  chat_ids4: "-1002465088995"    # Info alerts (Level 4)
  chat_ids5: "-1002465088995"    # Test alerts (Level 5)
  chat_ids6: ""                  # Reserved (Level 6)
  template_mode: "full"          # full, minimal
  template_language: "en"        # en, tw, zh, ja, ko
```

## 🤖 Telegram Bot Setup

### 1. Create Telegram Bot

1. Message [@BotFather](https://t.me/botfather) on Telegram
2. Send `/newbot` command
3. Follow instructions to create your bot
4. Save the bot token provided

### 2. Get Chat ID

For **private chats**:
1. Send a message to your bot
2. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
3. Look for the `chat.id` value

For **group chats**:
1. Add your bot to the group
2. Send a message mentioning the bot: `@your_bot_name hello`
3. Visit: `https://api.telegram.org/bot<YOUR_BOT_TOKEN>/getUpdates`
4. Look for the `chat.id` value (will be negative for groups)

### 3. Configure Chat IDs

Edit your configuration file and set the appropriate chat IDs:

```yaml
telegram:
  chat_ids0: "-1002465088995"    # Replace with your emergency group ID
  chat_ids1: "-1002465088995"    # Replace with your critical group ID
  # ... configure other levels as needed
```

## 🎨 Template Configuration

### Template Modes

- **`full`**: Complete alert information with all details
- **`minimal`**: Essential information only

### Template Languages

- **`en`**: English
- **`tw`**: Traditional Chinese (繁體中文)
- **`zh`**: Simplified Chinese (简体中文)
- **`ja`**: Japanese (日本語)
- **`ko`**: Korean (한국어)

### Template Format Options

Configure in `telegram_config.yaml` or `telegram_config.minimal.yaml`:

```yaml
format_options:
  show_links:
    enabled: true
  show_timestamps:
    enabled: true
  show_external_url:
    enabled: true
  show_generator_url:
    enabled: true
  show_emoji:
    enabled: true
  compact_mode:
    enabled: false
  max_summary_length:
    enabled: true
    value: 200
```

## 🔄 Hot Reload Configuration

The service supports hot reloading of configuration files. Changes to the following files will automatically reload:

- `configs/config.development.yaml`
- `configs/telegram_config.yaml`
- `configs/telegram_config.minimal.yaml`

### File Monitoring

The service uses `fsnotify` to monitor file changes. When a configuration file is modified:

1. **Main config changes**: Reload application configuration
2. **Template config changes**: Reinitialize template engine

## ⚠️ Important Notes

### Security Considerations

1. **Never commit sensitive data** like bot tokens to version control
2. **Use environment variables** for sensitive configuration in production
3. **Enable webhook authentication** in production environments
4. **Use HTTPS** in production deployments

### Chat ID Format

- **Private chats**: Positive integers (e.g., `123456789`)
- **Group chats**: Negative integers with `-100` prefix (e.g., `-1001234567890`)
- **Supergroups**: Also use the `-100` prefix format

### Environment-Specific Configuration

Use command-line flags to specify environment:

```bash
# Development
go run cmd/main.go -e development

# Production  
go run cmd/main.go -e production

# Test
go run cmd/main.go -e test
```

## 🔍 Troubleshooting

### Common Issues

1. **Bot token invalid**: Verify token from @BotFather
2. **Chat not found**: Check chat ID format and bot permissions
3. **Hot reload not working**: Verify file permissions and paths
4. **Template not loading**: Check template language code and file existence

### Debugging

Enable debug logging to troubleshoot configuration issues:

```yaml
log:
  level: "debug"
```

### Configuration Validation

The service validates configuration on startup. Check logs for any configuration errors or warnings.

## 📝 Example Configuration

Here's a complete example configuration file:

```yaml
# config.development.yaml
app:
  name: "alert-webhooks"
  version: "1.0.0"
  mode: "development"
  port: "9999"
  trusted_proxies: "127.0.0.1"

log:
  level: "debug"
  format: "json"
  output: "stdout"

webhooks:
  enable: true
  username: "admin"
  password: "secure_password_here"

telegram:
  enable: true
  token: "1234567890:ABCdefGHIjklMNOpqrSTUvwxYZ123456789"
  chat_ids0: "-1001234567890"
  chat_ids1: "-1001234567891"
  chat_ids2: "-1001234567892"
  chat_ids3: "-1001234567893"
  chat_ids4: "-1001234567894"
  chat_ids5: "-1001234567895"
  chat_ids6: ""
  template_mode: "full"
  template_language: "en"
```

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)
