# Template Usage Examples

This guide provides practical examples of using and customizing templates in the Alert Webhooks service.

## 🎯 Basic Usage

### Simple Alert Template

```go
{{/* Basic alert notification */}}
{{if eq .Status "firing"}}🚨 ALERT{{else}}✅ RESOLVED{{end}}

<b>{{.GroupLabels.alertname}}</b>
{{if .GroupLabels.severity}}<b>Severity:</b> {{.GroupLabels.severity}}{{end}}

{{range .FiringAlerts}}• {{.Annotations.summary}}{{end}}
{{range .ResolvedAlerts}}• {{.Annotations.summary}} (Resolved){{end}}
```

**Output Example**:
```
🚨 ALERT

DatabaseConnectionFailed
Severity: critical

• Database connection failed for user service
```

### Conditional Formatting

```go
{{/* Alert with conditional formatting */}}
{{if eq .Status "firing"}}
{{if eq .GroupLabels.severity "critical"}}🔴{{else if eq .GroupLabels.severity "warning"}}🟡{{else}}🔵{{end}}
{{else}}✅{{end}} 

<b>{{.GroupLabels.alertname}}</b>

{{if .GroupLabels.env}}
<b>Environment:</b> {{.GroupLabels.env}}
{{end}}

{{if .GroupLabels.namespace}}
<b>Namespace:</b> {{.GroupLabels.namespace}}
{{end}}

{{/* Show different content based on alert count */}}
{{$totalAlerts := add (len .FiringAlerts) (len .ResolvedAlerts)}}
{{if gt $totalAlerts 1}}
<b>Multiple Alerts ({{$totalAlerts}}):</b>
{{else}}
<b>Single Alert:</b>
{{end}}

{{range $index, $alert := .FiringAlerts}}
{{add $index 1}}. {{$alert.Annotations.summary}}
{{if $.FormatOptions.ShowTimestamps.Enabled}}
   Started: {{$alert.StartsAt}}
{{end}}
{{end}}
```

## 🌍 Multi-language Examples

### English Template (`tg_template_en.tmpl`)

```go
{{if eq .Status "firing"}}🚨 <b>Alert Firing</b>{{else}}✅ <b>Alert Resolved</b>{{end}}

<b>Alert Name:</b> {{.GroupLabels.alertname}}
<b>Environment:</b> {{.GroupLabels.env}}
<b>Severity:</b> {{.GroupLabels.severity}}
<b>Total Alerts:</b> {{add (len .FiringAlerts) (len .ResolvedAlerts)}}

{{if .FiringAlerts}}
<b>🔥 Firing Alerts ({{len .FiringAlerts}}):</b>
{{range $index, $alert := .FiringAlerts}}
<b>Alert {{add $index 1}}:</b>
• Summary: {{$alert.Annotations.summary}}
{{if $.FormatOptions.ShowTimestamps.Enabled}}
• Started: {{$alert.StartsAt}}
{{end}}
{{if and $.FormatOptions.ShowGeneratorURL.Enabled $alert.GeneratorURL}}
• <a href="{{$alert.GeneratorURL}}">View Query</a>
{{end}}
{{end}}
{{end}}

{{if .ResolvedAlerts}}
<b>✅ Resolved Alerts ({{len .ResolvedAlerts}}):</b>
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

### Traditional Chinese Template (`tg_template_tw.tmpl`)

```go
{{if eq .Status "firing"}}🚨 <b>警報觸發</b>{{else}}✅ <b>警報解除</b>{{end}}

<b>警報名稱:</b> {{.GroupLabels.alertname}}
<b>環境:</b> {{.GroupLabels.env}}
<b>嚴重程度:</b> {{.GroupLabels.severity}}
<b>總警報數:</b> {{add (len .FiringAlerts) (len .ResolvedAlerts)}}

{{if .FiringAlerts}}
<b>🔥 觸發中的警報 ({{len .FiringAlerts}}):</b>
{{range $index, $alert := .FiringAlerts}}
<b>警報 {{add $index 1}}:</b>
• 摘要: {{$alert.Annotations.summary}}
{{if $.FormatOptions.ShowTimestamps.Enabled}}
• 開始時間: {{$alert.StartsAt}}
{{end}}
{{if and $.FormatOptions.ShowGeneratorURL.Enabled $alert.GeneratorURL}}
• <a href="{{$alert.GeneratorURL}}">查看查詢</a>
{{end}}
{{end}}
{{end}}

{{if .ResolvedAlerts}}
<b>✅ 已解除警報 ({{len .ResolvedAlerts}}):</b>
{{range $index, $alert := .ResolvedAlerts}}
<b>警報 {{add $index 1}}:</b>
• 摘要: {{$alert.Annotations.summary}}
• 結束時間: {{$alert.EndsAt}}
{{end}}
{{end}}

{{if and .FormatOptions.ShowExternalURL.Enabled .ExternalURL}}
<a href="{{.ExternalURL}}">AlertManager 控制台</a>
{{end}}
```

## 🎨 Advanced Formatting Examples

### Compact Mode Template

```go
{{/* Ultra-compact format for minimal mode */}}
{{if eq .Status "firing"}}🚨{{else}}✅{{end}} <b>{{.GroupLabels.alertname}}</b>
{{if .GroupLabels.env}} | {{.GroupLabels.env}}{{end}}
{{if .GroupLabels.severity}} | {{.GroupLabels.severity}}{{end}}

{{if .FiringAlerts}}
{{range .FiringAlerts}}• {{.Annotations.summary}}{{end}}
{{end}}
{{if .ResolvedAlerts}}
{{range .ResolvedAlerts}}• {{.Annotations.summary}} ✓{{end}}
{{end}}
```

**Output**:
```
🚨 DatabaseDown | prod | critical
• Database connection lost for 5 minutes
```

### Rich Information Template

```go
{{/* Detailed template with all information */}}
{{if eq .Status "firing"}}
🚨 <b>ALERT FIRING</b>
{{else}}
✅ <b>ALERT RESOLVED</b>
{{end}}

┌─ <b>Alert Information</b>
│ <b>Name:</b> {{.GroupLabels.alertname}}
│ <b>Environment:</b> {{.GroupLabels.env | default "unknown"}}
│ <b>Severity:</b> {{.GroupLabels.severity | default "unknown"}}
{{if .GroupLabels.namespace}}│ <b>Namespace:</b> {{.GroupLabels.namespace}}{{end}}
{{if .GroupLabels.service}}│ <b>Service:</b> {{.GroupLabels.service}}{{end}}
└─ <b>Total:</b> {{add (len .FiringAlerts) (len .ResolvedAlerts)}} alert(s)

{{if .FiringAlerts}}
┌─ <b>🔥 FIRING ALERTS ({{len .FiringAlerts}})</b>
{{range $index, $alert := .FiringAlerts}}
│ <b>Alert {{add $index 1}}:</b>
│ ├─ <b>Summary:</b> {{$alert.Annotations.summary}}
{{if $alert.Annotations.description}}│ ├─ <b>Description:</b> {{$alert.Annotations.description}}{{end}}
{{if $.FormatOptions.ShowTimestamps.Enabled}}│ ├─ <b>Started:</b> {{$alert.StartsAt}}{{end}}
{{if $alert.Labels.instance}}│ ├─ <b>Instance:</b> {{$alert.Labels.instance}}{{end}}
{{if and $.FormatOptions.ShowGeneratorURL.Enabled $alert.GeneratorURL}}│ └─ <a href="{{$alert.GeneratorURL}}">🔍 View Query</a>{{else}}│ └─ (Query link hidden){{end}}
{{if ne $index (sub (len $.FiringAlerts) 1)}}│{{end}}
{{end}}
└─
{{end}}

{{if .ResolvedAlerts}}
┌─ <b>✅ RESOLVED ALERTS ({{len .ResolvedAlerts}})</b>
{{range $index, $alert := .ResolvedAlerts}}
│ <b>Alert {{add $index 1}}:</b>
│ ├─ <b>Summary:</b> {{$alert.Annotations.summary}}
│ ├─ <b>Started:</b> {{$alert.StartsAt}}
│ └─ <b>Ended:</b> {{$alert.EndsAt}}
{{if ne $index (sub (len $.ResolvedAlerts) 1)}}│{{end}}
{{end}}
└─
{{end}}

{{if and .FormatOptions.ShowExternalURL.Enabled .ExternalURL}}
🔗 <a href="{{.ExternalURL}}">Open AlertManager Console</a>
{{end}}

<i>Generated at {{.Timestamp | default "unknown"}}</i>
```

### Table Format Template

```go
{{/* Table-like format for structured data */}}
{{if eq .Status "firing"}}🚨 <b>ALERT FIRING</b>{{else}}✅ <b>RESOLVED</b>{{end}}

<code>
╭─────────────────────────────────────╮
│ Alert: {{.GroupLabels.alertname | printf "%-25s"}} │
│ Env:   {{.GroupLabels.env | printf "%-25s"}} │  
│ Sev:   {{.GroupLabels.severity | printf "%-25s"}} │
╰─────────────────────────────────────╯
</code>

{{if .FiringAlerts}}
<b>Firing:</b>
{{range $index, $alert := .FiringAlerts}}
<code>{{add $index 1}}. {{$alert.Annotations.summary}}</code>
{{end}}
{{end}}

{{if .ResolvedAlerts}}
<b>Resolved:</b>
{{range $index, $alert := .ResolvedAlerts}}
<code>{{add $index 1}}. {{$alert.Annotations.summary}}</code>
{{end}}
{{end}}
```

## 🔧 Custom Functions and Helpers

### String Manipulation

```go
{{/* Text truncation */}}
{{define "truncate"}}
{{if gt (len .) 50}}
{{slice . 0 50}}...
{{else}}
{{.}}
{{end}}
{{end}}

{{/* Usage */}}
<b>Summary:</b> {{template "truncate" .Annotations.summary}}
```

### Time Formatting

```go
{{/* Custom time format */}}
{{define "formatTime"}}
{{if .}}
{{.Format "2006-01-02 15:04:05 UTC"}}
{{else}}
N/A
{{end}}
{{end}}

{{/* Usage */}}
<b>Started:</b> {{template "formatTime" .StartsAt}}
```

### Status Icons

```go
{{/* Status icon mapping */}}
{{define "statusIcon"}}
{{if eq . "firing"}}🚨{{else if eq . "resolved"}}✅{{else}}❓{{end}}
{{end}}

{{define "severityIcon"}}
{{if eq . "critical"}}🔴{{else if eq . "warning"}}🟡{{else if eq . "info"}}🔵{{else}}⚪{{end}}
{{end}}

{{/* Usage */}}
{{template "statusIcon" .Status}} {{template "severityIcon" .GroupLabels.severity}} <b>{{.GroupLabels.alertname}}</b>
```

## 📊 Template Testing Examples

### Test Data Structure

```json
{
  "receiver": "telegram-test",
  "status": "firing",
  "groupLabels": {
    "alertname": "HighCPUUsage",
    "env": "production",
    "severity": "warning",
    "namespace": "monitoring",
    "service": "api-server"
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
        "summary": "CPU usage above 80% for 5 minutes on server-01",
        "description": "The server is experiencing high CPU load"
      },
      "startsAt": "2023-01-01T10:30:00.000Z",
      "generatorURL": "https://prometheus.example.com/graph?g0.expr=cpu_usage"
    }
  ]
}
```

### Testing Commands

```bash
# Test basic template
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
        "annotations": {"summary": "Test alert message"},
        "startsAt": "2023-01-01T00:00:00.000Z"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5

# Test multi-alert template
curl -X POST \
  -u admin:admin \
  -H "Content-Type: application/json" \
  -d '{
    "receiver": "test-multi",
    "status": "firing",
    "alerts": [
      {
        "status": "firing",
        "labels": {"alertname": "CPUHigh", "severity": "warning"},
        "annotations": {"summary": "CPU usage high on server-01"},
        "startsAt": "2023-01-01T00:00:00.000Z"
      },
      {
        "status": "resolved",
        "labels": {"alertname": "DiskSpace", "severity": "info"},
        "annotations": {"summary": "Disk space recovered on server-02"},
        "startsAt": "2023-01-01T00:00:00.000Z",
        "endsAt": "2023-01-01T00:15:00.000Z"
      }
    ]
  }' \
  http://localhost:9999/api/v1/telegram/chatid_L5
```

## 🎯 Best Practices

### Template Organization

1. **Use comments** to document template sections:
   ```go
   {{/* Alert header section */}}
   {{/* Alert details section */}}
   {{/* Footer with links */}}
   ```

2. **Group related logic**:
   ```go
   {{/* Status and severity indicators */}}
   {{template "statusIcon" .Status}}
   {{template "severityIcon" .GroupLabels.severity}}
   
   {{/* Alert information */}}
   {{template "alertInfo" .}}
   
   {{/* Action links */}}
   {{template "actionLinks" .}}
   ```

3. **Use consistent formatting**:
   ```go
   {{/* Always use the same pattern for labels */}}
   <b>{{.Field}}:</b> {{.Value}}
   ```

### Performance Optimization

1. **Minimize template complexity**:
   - Avoid nested loops where possible
   - Use simple conditionals
   - Cache complex calculations

2. **Limit message length**:
   ```go
   {{/* Truncate long summaries */}}
   {{if gt (len .Annotations.summary) 100}}
   {{slice .Annotations.summary 0 100}}...
   {{else}}
   {{.Annotations.summary}}
   {{end}}
   ```

3. **Use format options effectively**:
   ```go
   {{/* Only process when needed */}}
   {{if .FormatOptions.ShowTimestamps.Enabled}}
   {{/* Timestamp processing */}}
   {{end}}
   ```

### Error Handling

1. **Check for nil values**:
   ```go
   {{if .GroupLabels.env}}
   <b>Environment:</b> {{.GroupLabels.env}}
   {{end}}
   ```

2. **Provide defaults**:
   ```go
   <b>Environment:</b> {{.GroupLabels.env | default "unknown"}}
   ```

3. **Handle empty arrays**:
   ```go
   {{if .FiringAlerts}}
   {{/* Process firing alerts */}}
   {{else}}
   No firing alerts
   {{end}}
   ```

## 🌍 Language Options

- [English](../en/) (Current)
- [繁體中文](../zh/)
