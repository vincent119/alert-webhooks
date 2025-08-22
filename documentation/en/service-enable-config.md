# Service Enable Configuration Guide

This guide explains how to properly configure the enable status of various services.

## Service Enable Logic

The Alert-Webhooks system includes three main services:
- **Webhooks**: HTTP API endpoint service
- **Telegram**: Telegram Bot notification service  
- **Slack**: Slack Bot notification service

### Default Behavior

**Important**: All services have `enable` field default to `false` (disabled).

This design follows the "security first" principle:
- Only explicitly configured services will run
- Avoid accidentally enabling unnecessary services
- Prevent security issues caused by configuration errors

### Configuration Syntax

```yaml
# Method 1: Explicitly enable service
telegram:
  enable: true
  token: "your-token"

# Method 2: Explicitly disable service  
telegram:
  enable: false
  token: "your-token"

# Method 3: Don't set enable field (defaults to false)
telegram:
  token: "your-token"
  # enable field not set, service will be disabled

# Method 4: Set enable to empty value (defaults to false)
telegram:
  enable: 
  token: "your-token"
```

## Service Configuration Examples

### Webhooks Service

```yaml
# Enable Webhooks API
webhooks:
  enable: true
  base_auth_user: "admin"
  base_auth_password: "password"

# Disable Webhooks API  
webhooks:
  enable: false
  base_auth_user: "admin"
  base_auth_password: "password"
```

### Telegram Service

```yaml
# Enable Telegram Bot
telegram:
  enable: true
  token: "your-telegram-bot-token"
  chat_ids0: "-1001234567890"
  template_mode: "full"
  template_language: "tw"

# Disable Telegram Bot
telegram:
  enable: false
  token: "your-telegram-bot-token"
```

### Slack Service

```yaml
# Enable Slack Bot
slack:
  enable: true
  token: "xoxb-your-slack-token"
  channel: "#alerts"
  template_mode: "full"
  template_language: "eng"

# Disable Slack Bot
slack:
  enable: false
  token: "xoxb-your-slack-token"
```

## Common Use Cases

### Use Case 1: Telegram Only

```yaml
webhooks:
  enable: false  # Disable Webhooks API

telegram:
  enable: true   # Enable Telegram
  token: "your-telegram-token"
  # ... other config

slack:
  enable: false  # Disable Slack
```

### Use Case 2: Slack Only

```yaml
webhooks:
  enable: false  # Disable Webhooks API

telegram:
  enable: false  # Disable Telegram

slack:
  enable: true   # Enable Slack
  token: "your-slack-token"
  # ... other config
```

### Use Case 3: All Services

```yaml
webhooks:
  enable: true   # Enable Webhooks API
  base_auth_user: "admin"
  base_auth_password: "password"

telegram:
  enable: true   # Enable Telegram
  token: "your-telegram-token"
  # ... other config

slack:
  enable: true   # Enable Slack
  token: "your-slack-token"
  # ... other config
```

### Use Case 4: Development Environment (All Disabled)

```yaml
# Development environment might only need testing, no actual notifications
webhooks:
  enable: false

telegram:
  enable: false

slack:
  enable: false
```

## Verify Service Status

### Check Startup Logs

When starting the application, the system will log the enable status of each service:

```
Service enable status - Webhooks: true, Telegram: true, Slack: false
```

### API Endpoint Check

You can check service status through the following API endpoints:

```bash
# Check Telegram service status
curl -u admin:admin http://localhost:9999/api/v1/telegram/status

# Check Slack service status  
curl -u admin:admin http://localhost:9999/api/v1/slack/status
```

## Best Practices

1. **Explicit Configuration**: Always explicitly set `enable: true` or `enable: false`, avoid relying on defaults
2. **Environment Separation**: Use different config files for different environments to precisely control service enable status
3. **Security Considerations**: Only enable necessary services in production to reduce attack surface
4. **Test Verification**: Check logs and API endpoints after deployment to confirm services are enabled as expected

## Troubleshooting

### Service Not Starting

Check the `enable` field in config file:
```yaml
# Make sure it's set to true
telegram:
  enable: true
```

### Service Started Unexpectedly

Check config file, confirm `enable` field setting:
```yaml
# Make sure it's set to false or remove enable field
telegram:
  enable: false
```

### Can't Determine Service Status

Check application startup logs for service status record:
```
Service enable status - Webhooks: true, Telegram: false, Slack: true
```

## Related Documentation

- [Kubernetes Environment Variables Configuration](./kubernetes-env-vars.md)
- [Telegram Setup Guide](./telegram_setup.md)
- [Slack Setup Guide](./slack_setup.md)
