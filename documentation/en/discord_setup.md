# Discord Integration Setup Guide

This document explains how to set up a Discord bot to receive AlertManager notifications.

## 📋 Table of Contents
- [Creating Discord Application](#creating-discord-application)
- [Setting Bot Permissions](#setting-bot-permissions)
- [Obtaining Required Information](#obtaining-required-information)
- [Configuring Alert Webhooks](#configuring-alert-webhooks)
- [Testing Setup](#testing-setup)
- [Troubleshooting](#troubleshooting)

## 🚀 Creating Discord Application

### Step 1: Create Application
1. Go to [Discord Developer Portal](https://discord.com/developers/applications)
2. Click **"New Application"**
3. Enter application name (e.g., "Alert Webhooks Bot")
4. Click **"Create"**

### Step 2: Create Bot
1. Click **"Bot"** in the left sidebar
2. Click **"Add Bot"**
3. Confirm bot creation

### Step 3: Configure Bot Settings
1. In the **"Bot"** page:
   - Set bot name and avatar
   - Copy the **Bot Token** (this is your `DISCORD_TOKEN`)
   - ⚠️ **Important**: Keep the Token secret, do not share publicly

## 🔐 Setting Bot Permissions

### Required Permissions
The bot needs the following permissions to function properly:

- ✅ **Send Messages** - Send messages
- ✅ **View Channels** - View channels
- ✅ **Use External Emojis** - Use external emojis
- ✅ **Read Message History** - Read message history

### Invite Bot to Server
1. In **"OAuth2"** > **"URL Generator"**:
   - **Scopes**: Select `bot`
   - **Bot Permissions**: Select the required permissions above
2. Copy the generated URL and open in browser
3. Select your Discord server
4. Confirm permissions and authorize

## 📝 Obtaining Required Information

### Enable Developer Mode
1. In Discord, go to **User Settings** > **Advanced**
2. Enable **"Developer Mode"**

### Get Guild ID (Server ID)
1. Right-click on server name
2. Select **"Copy ID"**
3. This is your `guild_id`

### Get Channel IDs
1. Right-click on channel name
2. Select **"Copy ID"**
3. Repeat for all required channels

### Recommended Server Structure
```
📁 Your Discord Server
├── 🚨 alerts-critical     (Level 0) - Critical alerts
├── ⚠️  alerts-high        (Level 1) - High priority
├── 📢 alerts-normal       (Level 2) - Normal alerts  
├── 📝 alerts-info         (Level 3) - Info notifications
├── 🔧 alerts-debug        (Level 4) - Debug messages
└── 📦 alerts-backup       (Level 5) - Backup channel
```

## ⚙️ Configuring Alert Webhooks

### Configuration File Setup
Edit your `config.yaml` file:

```yaml
discord:
  enable: true
  token: "${DISCORD_TOKEN}" # Use environment variable
  guild_id: "your-server-id"
  username: "Alert Webhooks Bot"
  
  # Channel mapping for Alert Levels
  channels:
    chat_ids0: "critical-alerts-channel-id"    # Level 0 - Critical
    chat_ids1: "high-priority-channel-id"      # Level 1 - High  
    chat_ids2: "normal-alerts-channel-id"      # Level 2 - Normal
    chat_ids3: "info-notifications-channel-id" # Level 3 - Info
    chat_ids4: "debug-messages-channel-id"     # Level 4 - Debug
    chat_ids5: "backup-channel-id"             # Level 5 - Backup
  
  # Discord specific options
  message_format: "markdown"
  mention_roles: [] # Optional: Role IDs to @mention
  
  # Template configuration
  template_mode: "full"        # minimal, full
  template_language: "eng"     # eng, tw, zh, ja, ko
```

### Environment Variable Setup
Set the Discord Bot Token environment variable:

```bash
export DISCORD_TOKEN="your-discord-bot-token-here"
```

### Kubernetes Setup
If using Kubernetes, set in Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: discord-secrets
type: Opaque
stringData:
  DISCORD_TOKEN: "your-discord-bot-token-here"
```

## Level Mapping Table

| Level | Channel Type | Description | Example Usage |
|-------|--------------|-------------|---------------|
| L0    | 🚨 Critical | Critical alerts | System failures, service outages |
| L1    | ⚠️ High | High priority | Performance degradation, error rate increases |
| L2    | 📢 Normal | Normal alerts | General monitoring warnings |
| L3    | 📝 Info | Info notifications | Status updates, deployment notifications |
| L4    | 🔧 Debug | Debug messages | Debug information, detailed logs |
| L5    | 📦 Backup | Backup channel | Testing, backup notifications |

## 🧪 Testing Setup

### Test Bot Connection
```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin"
```

### Test Channel Validation
```bash
curl -X POST "http://localhost:9999/api/v1/discord/validate/your-channel-id" \
  -u "admin:admin"
```

### Send Test Message
```bash
curl -X POST "http://localhost:9999/api/v1/discord/test/your-channel-id" \
  -u "admin:admin"
```

### Test Level Routing
```bash
curl -X POST "http://localhost:9999/api/v1/discord/chatid_L0" \
  -H "Content-Type: application/json" \
  -u "admin:admin" \
  -d '{"message": "Test Level 0 message"}'
```

## 🔧 Troubleshooting

### Common Errors

#### 1. "Missing Permissions" Error
**Cause**: Bot lacks necessary permissions
**Solution**: 
- Ensure bot has "Send Messages" permission
- Check channel-specific permission settings
- Re-invite bot with proper permissions

#### 2. "Unknown Channel" Error  
**Cause**: Channel ID is incorrect or bot cannot access it
**Solution**:
- Verify channel ID is correct
- Ensure bot has joined the server
- Check if channel is private

#### 3. "Unauthorized" Error
**Cause**: Invalid Discord Bot Token
**Solution**:
- Verify Token is correct
- Ensure Token has "Bot " prefix (automatically added by code)
- Regenerate Bot Token if needed

#### 4. "Bot is not in channel" Error
**Cause**: Bot hasn't joined the specific channel
**Solution**:
- Confirm bot has joined the server
- Check channel permission settings
- Try manually @mentioning the bot

### Log Checking
Check Discord-related logs:
```bash
grep "Discord" ./logs/server.log
```

### Verify Configuration
Check if configuration is loaded correctly:
```bash
curl -X GET "http://localhost:9999/api/v1/discord/status" \
  -u "admin:admin" | jq .
```

## 📚 Related Documentation
- [Discord Usage Guide](discord_usage.md)
- [Kubernetes Environment Variables](kubernetes-env-vars.md)
- [Service Enable Configuration](service-enable-config.md)
- [Template System Documentation](template-system.md)

## 🆘 Need Help?
If you encounter issues, please check:
1. Discord Bot Token validity
2. Bot permissions are correct
3. Channel IDs are accurate
4. Network connectivity
5. Application logs for error messages
