# Template System Guide

This guide explains how to use and customize the template system for Alert Webhooks notifications.

## 🏗️ Template Architecture

The template system supports multiple languages and formats with hot-reloading capabilities.

### Template Engine Features

- **Multi-language support**: English, Traditional Chinese, Simplified Chinese, Japanese, Korean
- **Hot reload**: Templates update automatically when files change
- **Dual syntax support**: Native Go templates (`.tmpl`) and Jinja2-like syntax (`.j2`)
- **Format options**: Configurable display options per template mode
- **Template inheritance**: Shared formatting and structure

## 📁 Template Structure

### Template Files

Templates are located in `templates/telegram/`:

```
templates/telegram/
├── tg_template_en.tmpl     # English template
├── tg_template_tw.tmpl     # Traditional Chinese template  
├── tg_template_zh.tmpl     # Simplified Chinese template
├── tg_template_ja.tmpl     # Japanese template
└── tg_template_ko.tmpl     # Korean template
```

### Template Configuration

Template formatting is configured in:
- `configs/telegram_config.yaml` - Default/full mode settings
- `configs/telegram_config.minimal.yaml` - Minimal mode settings

## 🎨 Template Syntax

### Go Template Syntax (`.tmpl` files)

```go
{{/* Alert status header */}}
{{if eq .Status "firing"}}🚨 <b>Alert Firing</b>{{else}}✅ <b>Alert Resolved</b>{{end}}

{{/* Alert information */}}
<b>Alert Name:</b> {{.GroupLabels.alertname}}
<b>Environment:</b> {{.GroupLabels.env}}
<b>Severity:</b> {{.GroupLabels.severity}}

{{/* Conditional formatting */}}
{{if .FormatOptions.ShowTimestamps.Enabled}}
<b>Started:</b> {{.Alert.StartsAt}}
{{end}}

{{/* Loop through alerts */}}
{{range $index, $alert := .FiringAlerts}}
<b>Alert {{add $index 1}}:</b>
• Summary: {{$alert.Annotations.summary}}
{{end}}

{{/* Show links conditionally */}}
{{if and .FormatOptions.ShowLinks.Enabled .ExternalURL}}
<a href="{{.ExternalURL}}">View Details</a>
{{end}}
```

### Available Template Variables

#### Root Variables
- `.Status` - Overall status ("firing" or "resolved")
- `.GroupLabels` - Common labels across all alerts
- `.CommonAnnotations` - Common annotations
- `.ExternalURL` - AlertManager external URL
- `.FiringAlerts` - Array of currently firing alerts
- `.ResolvedAlerts` - Array of resolved alerts
- `.FormatOptions` - Template formatting configuration

#### Alert Variables (within loops)
- `.Status` - Alert status
- `.Labels` - Alert-specific labels
- `.Annotations` - Alert-specific annotations
- `.StartsAt` - Alert start time
- `.EndsAt` - Alert end time (if resolved)
- `.GeneratorURL` - Prometheus query URL
- `.Fingerprint` - Unique alert identifier

#### Format Options
- `.FormatOptions.ShowLinks.Enabled` - Show hyperlinks
- `.FormatOptions.ShowTimestamps.Enabled` - Show timestamps
- `.FormatOptions.ShowExternalURL.Enabled` - Show external URL
- `.FormatOptions.ShowGeneratorURL.Enabled` - Show generator URL
- `.FormatOptions.ShowEmoji.Enabled` - Show emoji icons
- `.FormatOptions.CompactMode.Enabled` - Use compact formatting

### Template Functions

#### Built-in Functions
- `add` - Add numbers: `{{add $index 1}}`
- `eq` - Equal comparison: `{{if eq .Status "firing"}}`
- `ne` - Not equal: `{{if ne .Status "resolved"}}`
- `and` - Logical AND: `{{if and .A .B}}`
- `or` - Logical OR: `{{if or .A .B}}`

#### Custom Functions
- `index` - Access array elements: `{{index .Alerts 0}}`

## 🌍 Multi-language Support

### Language Configuration

Set the language in your configuration:

```yaml
telegram:
  template_language: "en"  # en, tw, zh, ja, ko
```

### Language Files

Each language has its own template file with localized text:

#### English (`tg_template_en.tmpl`)
```go
{{if eq .Status "firing"}}🚨 <b>Alert Firing</b>{{else}}✅ <b>Alert Resolved</b>{{end}}
<b>Alert Name:</b> {{.GroupLabels.alertname}}
<b>Environment:</b> {{.GroupLabels.env}}
```

#### Traditional Chinese (`tg_template_tw.tmpl`)
```go
{{if eq .Status "firing"}}🚨 <b>警報觸發</b>{{else}}✅ <b>警報解除</b>{{end}}
<b>警報名稱:</b> {{.GroupLabels.alertname}}
<b>環境:</b> {{.GroupLabels.env}}
```

### Language Fallback

If a template for the configured language is not found, the system falls back to:
1. English (`en`)
2. Default hardcoded template

## 🎛️ Template Modes

### Full Mode (`full`)

Shows complete alert information including:
- Full alert details
- Timestamps
- External links
- Generator URLs
- All annotations

Configuration in `telegram_config.yaml`:
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
```

### Minimal Mode (`minimal`)

Shows essential information only:
- Alert name and status
- Critical annotations
- No links or timestamps

Configuration in `telegram_config.minimal.yaml`:
```yaml
format_options:
  show_links:
    enabled: false
  show_timestamps:
    enabled: false
  show_external_url:
    enabled: false
  show_generator_url:
    enabled: false
```

## 🔧 Customizing Templates

### Creating Custom Templates

1. **Copy existing template**:
   ```bash
   cp templates/telegram/tg_template_en.tmpl templates/telegram/tg_template_custom.tmpl
   ```

2. **Modify template content**:
   ```go
   {{/* Custom header */}}
   🔔 <b>Custom Alert Format</b>
   
   {{/* Your custom formatting here */}}
   <b>Alert:</b> {{.GroupLabels.alertname}}
   <b>Level:</b> {{.GroupLabels.severity}}
   
   {{if .FiringAlerts}}
   <b>Firing Alerts ({{len .FiringAlerts}}):</b>
   {{range .FiringAlerts}}
   • {{.Annotations.summary}}
   {{end}}
   {{end}}
   ```

3. **Update language configuration**:
   ```yaml
   telegram:
     template_language: "custom"
   ```

### Template Best Practices

1. **Use HTML formatting** for Telegram:
   - `<b>text</b>` - Bold text
   - `<i>text</i>` - Italic text
   - `<a href="url">text</a>` - Links
   - `<code>text</code>` - Monospace text

2. **Check for empty values**:
   ```go
   {{if .GroupLabels.env}}
   <b>Environment:</b> {{.GroupLabels.env}}
   {{end}}
   ```

3. **Use conditional formatting**:
   ```go
   {{if .FormatOptions.ShowEmoji.Enabled}}
   {{if eq .Status "firing"}}🚨{{else}}✅{{end}}
   {{end}}
   ```

4. **Limit message length**:
   ```go
   {{if .FormatOptions.MaxSummaryLength.Enabled}}
   {{if gt (len .Annotations.summary) .FormatOptions.MaxSummaryLength.Value}}
   {{slice .Annotations.summary 0 .FormatOptions.MaxSummaryLength.Value}}...
   {{else}}
   {{.Annotations.summary}}
   {{end}}
   {{end}}
   ```

## 🔄 Hot Reload

Templates support hot reloading:

1. **Modify template file**: Edit any `.tmpl` file in `templates/telegram/`
2. **Configuration changes**: Modify `telegram_config.yaml` or `telegram_config.minimal.yaml`
3. **Automatic reload**: Service detects changes and reloads templates
4. **No restart required**: Service continues running with new templates

### Monitoring Reload Events

Watch logs for reload events:
```
INFO Template engine reloaded successfully
INFO Config file changed, reloading...
```

## 🧪 Testing Templates

### Test Template Rendering

1. **Send test alert**:
   ```bash
   curl -X POST \
     -u admin:admin \
     -H "Content-Type: application/json" \
     -d '{
       "receiver": "test",
       "status": "firing",
       "alerts": [
         {
           "status": "firing",
           "labels": {"alertname": "TestAlert", "severity": "warning"},
           "annotations": {"summary": "Test alert summary"},
           "startsAt": "2023-01-01T00:00:00.000Z"
         }
       ]
     }' \
     http://localhost:9999/api/v1/telegram/chatid_L5
   ```

2. **Check Telegram output**: Verify formatting and content

3. **Test different languages**: Change `template_language` and test again

### Template Validation

The template engine validates templates on load:
- **Syntax errors**: Invalid Go template syntax
- **Missing variables**: References to undefined variables
- **Function errors**: Invalid function usage

Check logs for validation errors:
```
ERROR Template validation failed: template: parsing error
```

## 📋 Example Templates

### Minimal Alert Template
```go
{{if eq .Status "firing"}}🚨 ALERT{{else}}✅ RESOLVED{{end}}
<b>{{.GroupLabels.alertname}}</b>
{{range .FiringAlerts}}{{.Annotations.summary}}{{end}}
{{range .ResolvedAlerts}}{{.Annotations.summary}}{{end}}
```

### Detailed Alert Template
```go
{{if eq .Status "firing"}}🚨 <b>Alert Firing</b>{{else}}✅ <b>Alert Resolved</b>{{end}}

<b>Alert Name:</b> {{.GroupLabels.alertname}}
<b>Environment:</b> {{.GroupLabels.env | default "unknown"}}
<b>Severity:</b> {{.GroupLabels.severity}}
<b>Total Alerts:</b> {{add (len .FiringAlerts) (len .ResolvedAlerts)}}

{{if .FiringAlerts}}
<b>🔥 Firing ({{len .FiringAlerts}}):</b>
{{range $index, $alert := .FiringAlerts}}
<b>Alert {{add $index 1}}:</b>
• Summary: {{$alert.Annotations.summary}}
• Started: {{$alert.StartsAt}}
{{if $.FormatOptions.ShowGeneratorURL.Enabled}}
• <a href="{{$alert.GeneratorURL}}">View Query</a>
{{end}}
{{end}}
{{end}}

{{if .ResolvedAlerts}}
<b>✅ Resolved ({{len .ResolvedAlerts}}):</b>
{{range $index, $alert := .ResolvedAlerts}}
<b>Alert {{add $index 1}}:</b>
• Summary: {{$alert.Annotations.summary}}
• Ended: {{$alert.EndsAt}}
{{end}}
{{end}}

{{if and .FormatOptions.ShowExternalURL.Enabled .ExternalURL}}
<a href="{{.ExternalURL}}">AlertManager Console</a>
{{end}}
```

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)
