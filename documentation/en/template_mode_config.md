# Template Mode Configuration

This guide explains how to configure and customize template modes for different notification requirements.

## 🎛️ Overview

The Alert Webhooks service supports two main template modes:
- **Full Mode**: Complete alert information with all details
- **Minimal Mode**: Essential information only for critical alerts

## 📋 Template Modes

### Full Mode (`full`)

**Purpose**: Comprehensive alert notifications with complete context

**Features**:
- Complete alert details
- Timestamps (start/end times)
- External links to AlertManager
- Generator URLs to Prometheus queries
- Full annotations and labels
- Emoji indicators
- Rich formatting

**Use Cases**:
- Development and testing environments
- Detailed monitoring dashboards
- Post-incident analysis
- Training and educational purposes

### Minimal Mode (`minimal`)

**Purpose**: Essential alert information for rapid response

**Features**:
- Alert name and status only
- Critical annotations (summary)
- No timestamps or links
- Compact formatting
- Reduced message length

**Use Cases**:
- Production emergency alerts
- Mobile notifications
- High-frequency alerting
- Bandwidth-limited environments

## 🔧 Configuration

### Setting Template Mode

Configure the template mode in your main configuration file:

```yaml
# configs/config.development.yaml
telegram:
  template_mode: "full"    # or "minimal"
  template_language: "en"
```

### Template Configuration Files

#### Full Mode Configuration (`telegram_config.yaml`)

```yaml
naming_convention:
  prefix: "tg_template_"
  suffix: ".tmpl"
  directory: "templates/telegram"

supported_languages:
  - "en"    # English
  - "tw"    # Traditional Chinese
  - "zh"    # Simplified Chinese
  - "ja"    # Japanese
  - "ko"    # Korean

fallback_order:
  - "en"
  - "tw"

format_options:
  show_links:
    enabled: true
    description: "Display hyperlinks to external resources"
    
  show_timestamps:
    enabled: true
    description: "Show alert start and end times"
    
  show_external_url:
    enabled: true
    description: "Display link to AlertManager console"
    
  show_generator_url:
    enabled: true
    description: "Display link to Prometheus query"
    
  show_emoji:
    enabled: true
    description: "Use emoji indicators for alert status"
    
  compact_mode:
    enabled: false
    description: "Use compact formatting to reduce message length"
    
  max_summary_length:
    enabled: true
    value: 200
    description: "Maximum length of alert summary text"
```

#### Minimal Mode Configuration (`telegram_config.minimal.yaml`)

```yaml
naming_convention:
  prefix: "tg_template_"
  suffix: ".tmpl"
  directory: "templates/telegram"

supported_languages:
  - "en"
  - "tw"
  - "zh" 
  - "ja"
  - "ko"

fallback_order:
  - "en"
  - "tw"

format_options:
  show_links:
    enabled: false
    description: "Hide hyperlinks for cleaner messages"
    
  show_timestamps:
    enabled: false
    description: "Hide timestamps to reduce message length"
    
  show_external_url:
    enabled: false
    description: "Hide AlertManager console link"
    
  show_generator_url:
    enabled: false
    description: "Hide Prometheus query link"
    
  show_emoji:
    enabled: true
    description: "Keep emoji for quick visual identification"
    
  compact_mode:
    enabled: true
    description: "Use compact formatting"
    
  max_summary_length:
    enabled: true
    value: 100
    description: "Shorter summary for minimal mode"
```

## 🔄 Hot Reloading

Template modes support hot reloading for dynamic configuration changes:

### Automatic Detection

The service monitors these files for changes:
- `configs/config.development.yaml` (template_mode changes)
- `configs/telegram_config.yaml` (full mode settings)
- `configs/telegram_config.minimal.yaml` (minimal mode settings)

### Reload Process

1. **File Change Detection**: Service detects file modification
2. **Configuration Reload**: New settings are loaded
3. **Template Engine Restart**: Template engine reinitializes with new config
4. **Service Continues**: No service restart required

### Monitoring Reloads

Watch for reload events in logs:
```
INFO Config file changed, reloading...
INFO Template engine reloaded successfully with mode: minimal
INFO Format options updated: show_links=false, show_timestamps=false
```

## 🎨 Format Options

### Link Display Options

```yaml
show_links:
  enabled: true/false
```
- **Enabled**: Shows clickable links in messages
- **Disabled**: Text-only messages without hyperlinks

Example output:
```
# Enabled
View Details: https://alertmanager.example.com

# Disabled  
(Links hidden for minimal display)
```

### Timestamp Options

```yaml
show_timestamps:
  enabled: true/false
```
- **Enabled**: Shows alert start and end times
- **Disabled**: Hides timestamp information

Example output:
```
# Enabled
Started: 2023-01-01T10:30:00Z
Ended: 2023-01-01T10:45:00Z

# Disabled
(Timestamps hidden)
```

### External URL Options

```yaml
show_external_url:
  enabled: true/false
```
- **Enabled**: Shows link to AlertManager console
- **Disabled**: Hides AlertManager link

### Generator URL Options

```yaml
show_generator_url:
  enabled: true/false
```
- **Enabled**: Shows link to Prometheus query
- **Disabled**: Hides Prometheus link

### Emoji Options

```yaml
show_emoji:
  enabled: true/false
```
- **Enabled**: Uses emoji indicators (🚨, ✅, 🔥)
- **Disabled**: Text-only status indicators

### Compact Mode

```yaml
compact_mode:
  enabled: true/false
```
- **Enabled**: Single-line formatting where possible
- **Disabled**: Multi-line detailed formatting

Example comparison:
```
# Compact Mode
🚨 ALERT: DatabaseDown | prod | critical | Started: 10:30

# Normal Mode
🚨 Alert Firing
Alert Name: DatabaseDown
Environment: prod
Severity: critical
Started: 2023-01-01T10:30:00Z
```

### Summary Length Limits

```yaml
max_summary_length:
  enabled: true/false
  value: 100
```
- **Enabled**: Truncates summaries to specified length
- **Disabled**: Shows full summary text

## 📊 Mode Comparison

| Feature | Full Mode | Minimal Mode |
|---------|-----------|--------------|
| Alert Details | Complete | Essential only |
| Timestamps | Yes | No |
| External Links | Yes | No |
| Generator URLs | Yes | No |
| Rich Formatting | Yes | Compact |
| Message Length | Long | Short |
| Response Time | Slower | Faster |
| Bandwidth Usage | Higher | Lower |

## 🚀 Usage Examples

### Development Environment (Full Mode)

```yaml
telegram:
  template_mode: "full"
  template_language: "en"
```

**Message Output**:
```
🚨 Alert Firing

Alert Name: HighCPUUsage
Environment: development
Severity: warning
Namespace: monitoring
Total Alerts: 1
Firing: 1

🔥 Firing Alerts:
Alert 1:
• Summary: CPU usage above 80% for 5 minutes
• Started: 2023-01-01T10:30:00Z
• View Query: https://prometheus.example.com/graph?...

AlertManager Console: https://alertmanager.example.com
```

### Production Environment (Minimal Mode)

```yaml
telegram:
  template_mode: "minimal"
  template_language: "en"
```

**Message Output**:
```
🚨 HighCPUUsage
CPU usage above 80% for 5 minutes
```

## 🔧 Custom Mode Configuration

### Creating Custom Modes

You can create custom template configurations:

1. **Create custom config file**:
   ```bash
   cp configs/telegram_config.yaml configs/telegram_config.custom.yaml
   ```

2. **Modify format options**:
   ```yaml
   format_options:
     show_links:
       enabled: true
     show_timestamps:
       enabled: false
     # ... customize other options
   ```

3. **Update service configuration**:
   ```yaml
   telegram:
     template_mode: "custom"
   ```

### Dynamic Mode Switching

You can switch modes at runtime by modifying the configuration file:

```bash
# Switch to minimal mode
sed -i 's/template_mode: "full"/template_mode: "minimal"/' configs/config.development.yaml

# Service automatically reloads with new mode
```

## 🧪 Testing Different Modes

### Test Full Mode

```bash
# Set full mode
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "test-full",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {"alertname": "TestAlert", "severity": "warning"},
        "annotations": {"summary": "Test alert for full mode"},
        "startsAt": "2023-01-01T00:00:00.000Z",
        "generatorURL": "https://prometheus.example.com/graph"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

### Test Minimal Mode

```bash
# Switch to minimal mode in config, then send same alert
# Observe difference in message format
```

## ⚠️ Best Practices

### Mode Selection Guidelines

1. **Production Critical Alerts**: Use minimal mode
   - Faster delivery
   - Reduced notification fatigue
   - Essential information only

2. **Development/Testing**: Use full mode
   - Complete debugging information
   - All context available
   - Educational value

3. **Monitoring Dashboards**: Use full mode
   - Rich information for analysis
   - Links for further investigation

### Performance Considerations

- **Minimal mode**: Lower bandwidth, faster delivery
- **Full mode**: Higher bandwidth, more processing time
- **Hot reload**: Small performance impact during reload

### Message Length Limits

Telegram has message length limits:
- **Maximum message length**: 4096 characters
- **Recommended length**: < 1000 characters for readability

Use `max_summary_length` to control message size.

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)
