# Slack Setup Guide

This guide will walk you through the process of setting up and configuring a Slack Bot.

## Prerequisites

- Slack workspace admin permissions
- Ability to create and install Slack applications

## Step 1: Create Slack Application

### 1.1 Go to Slack API Console

1. Open your browser and go to [Slack API](https://api.slack.com/apps)
2. Click **"Create New App"**
3. Select **"From scratch"**

### 1.2 Fill Application Information

1. **App Name**: Enter application name (e.g., `Alert Bot`)
2. **Pick a workspace**: Select your workspace
3. Click **"Create App"**

## Step 2: Configure Bot Permissions

### 2.1 Set OAuth & Permissions

1. In the left sidebar, click **"OAuth & Permissions"**
2. Scroll to **"Scopes"** section
3. Add the following permissions under **"Bot Token Scopes"**:

#### Required Permissions:

```
chat:write          # Send messages
chat:write.public    # Send messages to public channels
channels:read        # Read channel information
groups:read          # Read private channel information
im:read              # Read direct messages
mpim:read            # Read group messages
```

#### Optional Permissions (Recommended):

```
users:read           # Read user information (for @mentions)
channels:join        # Auto-join channels
```

## Step 3: Install Application

### 3.1 Install to Workspace

1. Scroll to **"OAuth Tokens for Your Workspace"** section at the top
2. Click **"Install to Workspace"**
3. Review permissions and click **"Allow"**

### 3.2 Get Bot Token

After installation, you'll see the **"Bot User OAuth Token"**:

```
xoxb-xxxxxxxxxxxxx-xxxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
```

**Important**: Save this token securely, it will be used in application configuration.

## Step 4: Configure Channels

### 4.1 Invite Bot to Channels

For each channel that should receive alerts:

1. Go to the channel
2. Type: `/invite @your-bot-name`
3. Or click **"Add apps"** in channel info and select your Bot

### 4.2 Get Channel ID (Optional)

If you need to use channel ID instead of channel name:

1. Right-click on the channel name
2. Select **"Copy link"**
3. Channel ID is the last part of the URL: `https://yourworkspace.slack.com/archives/C1234567890`
4. Channel ID format: `C1234567890`

## Step 5: Application Configuration

### 5.1 Configuration File Setup

Add Slack configuration to `config.yaml`:

```yaml
slack:
  # Enable Slack service
  enable: true

  # Bot Token (can also be set via SLACK_TOKEN environment variable)
  token: "xoxb-your-slack-bot-token"

  # Default channel (fallback channel)
  channel: "#alerts"

  # Bot display settings
  username: "Alert Bot"
  icon_emoji: ":warning:" # or use icon_url
  # icon_url: "https://example.com/bot-icon.png"

  # Multi-channel configuration (assign by alert level)
  channels:
    chat_ids0: "#critical-alerts" # Level 0 - Critical alerts
    chat_ids1: "#warning-alerts" # Level 1 - Warnings
    chat_ids2: "#info-alerts" # Level 2 - Information
    chat_ids3: "#debug-alerts" # Level 3 - Debug
    chat_ids4: "#test-alerts" # Level 4 - Testing
    chat_ids5: "#other-alerts" # Level 5 - Other

  # Message options
  link_names: true # Enable @mentions
  unfurl_links: false # Don't unfurl link previews
  unfurl_media: false # Don't unfurl media previews

  # Template settings
  template_mode: "full" # minimal or full
  template_language: "en" # eng, tw, zh, ja, ko
```

### 5.2 Environment Variable Setup (Recommended for Production)

```bash
# Slack Bot Token
export SLACK_TOKEN="xoxb-your-slack-bot-token"
```

In Kubernetes:

```yaml
env:
  - name: SLACK_TOKEN
    valueFrom:
      secretKeyRef:
        name: alert-webhooks-secrets
        key: slack-token
```

## Step 6: Test Configuration

### 6.1 Start Application

After ensuring correct configuration, start the application:

```bash
go run cmd/main.go
```

Check startup logs to confirm Slack service is enabled:

```
Service enable status - Webhooks: true, Telegram: false, Slack: true
```

### 6.2 Test API Endpoints

#### Check service status:

```bash
curl -u admin:admin http://localhost:9999/api/v1/slack/status
```

#### Test send message:

```bash
curl -X POST -H "Content-Type: application/json" -u admin:admin \
  -d '{"message": "Test message"}' \
  "http://localhost:9999/api/v1/slack/channel/alerts"
```

#### Test level routing:

```bash
curl -X POST -H "Content-Type: application/json" -u admin:admin \
  -d '{"message": "Critical alert test"}' \
  "http://localhost:9999/api/v1/slack/chatid_L0"
```

## Channel Configuration

### Channel Name Formats

Supports the following formats:

- **Public channels**: `#channel-name`
- **Private channels**: `#private-channel` (Bot must be invited)
- **Channel ID**: `C1234567890`

### Level Mapping

| chat_ids  | Level | 群組用途 (Group Purpose)                    |
| --------- | ----- | ------------------------------------------- |
| chat_ids0 | 0     | Information group（資訊群組）               |
| chat_ids1 | 1     | General message group（一般訊息群組）       |
| chat_ids2 | 2     | Critical notification group（重要通知群組） |
| chat_ids3 | 3     | Emergency alert group（緊急警報群組）       |
| chat_ids4 | 4     | Testing group（測試群組）                   |
| chat_ids5 | 5     | Backup group（備用群組）                    |

## Troubleshooting

### Issue 1: Bot Cannot Send Messages

**Error**: `not_in_channel`

**Solution**:

1. Ensure Bot is invited to target channel
2. Run: `/invite @your-bot-name` in that channel

### Issue 2: Insufficient Permissions

**Error**: `missing_scope`

**Solution**:

1. Return to Slack API console
2. Check and add required Bot Token Scopes
3. Reinstall application to workspace

### Issue 3: Invalid Token

**Error**: `invalid_auth`

**Solution**:

1. Check if token format is correct (should start with `xoxb-`)
2. Confirm token hasn't expired
3. Regenerate token

### Issue 4: Channel Not Found

**Error**: `channel_not_found`

**Solution**:

1. Confirm channel name spelling is correct
2. Confirm channel exists and Bot has access
3. Use channel ID instead of channel name

## Advanced Configuration

### Rich Text Messages

Use rich text API to send formatted messages:

```bash
curl -X POST -H "Content-Type: application/json" -u admin:admin \
  -d '{
    "blocks": [
      {
        "type": "section",
        "text": {
          "type": "mrkdwn",
          "text": "*Alert Notification*\nStatus: Firing\nSeverity: High"
        }
      }
    ]
  }' \
  "http://localhost:9999/api/v1/slack/rich/alerts"
```

### Template Customization

You can modify the following template files to customize Slack message format:

- `templates/alerts/alert_template_tw.tmpl` (Traditional Chinese) 🇹🇼
- `templates/alerts/alert_template_eng.tmpl` (English) 🇺🇸
- `templates/alerts/alert_template_zh.tmpl` (Simplified Chinese) 🇨🇳
- `templates/alerts/alert_template_ja.tmpl` (Japanese) 🇯🇵
- `templates/alerts/alert_template_ko.tmpl` (Korean) 🇰🇷

### Multi-Workspace Support

To support multiple Slack workspaces:

1. Create different configuration files for each workspace
2. Use different environment variables for different tokens

## Related Documentation

- [Service Enable Configuration Guide](./service-enable-config.md)
- [Kubernetes Environment Variables Configuration](./kubernetes-env-vars.md)
- [Template Usage Guide](./template_usage.md)
