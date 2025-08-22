# Telegram 模板系統

這個目錄包含了用於生成 Telegram 警報訊息的 Jinja2 風格模板。

## 模板檔案

系統支援兩種模板格式：

### 支援的檔案格式

1. **`.tmpl` 檔案** - Go template 語法 (推薦)
2. **`.j2` 檔案** - Jinja2 語法 (自動轉換)

### 目前可用的模板

- `tg_template_eng.tmpl` - 英文模板
- `tg_template_tw.tmpl` - 繁體中文模板
- `tg_template_zh.tmpl` - 簡體中文模板

### 優先順序

如果同一語言有多種格式的模板檔案，系統會按以下順序選擇：

1. `.tmpl` 檔案 (Go template 語法)
2. `.j2` 檔案 (Jinja2 語法)

## 模板語法

支援兩種語法：

### Go Template 語法 (.tmpl 檔案，推薦)

```go
{{ if gt .FiringCount 0 }}
🚨 *警報通知*
{{ end }}
*狀態:* {{ .Status }}
{{ range $index, $alert := .Alerts }}
*警報 {{ add $index 1 }}:*
{{ end }}
```

### Jinja2 語法 (.j2 檔案)

模板使用 Jinja2 風格的語法，系統會自動轉換為 Go template 語法。

**注意**: Jinja2 轉換功能支援基本的語法轉換，包括：

- 變數引用：`{{ variable }}` → `{{ .Variable }}`
- 條件語句：`{% if condition %}` → `{{ if condition }}`
- 迴圈語句：`{% for item in items %}` → `{{ range $index, $item := .Items }}`

**限制**: 轉換功能是基於字串替換的簡化實現，複雜的 Jinja2 語法可能無法正確轉換。建議使用 Go template (.tmpl) 格式以獲得最佳相容性。

### 可用的變數

- `status` - 警報狀態 (firing/resolved)
- `alert_name` - 警報名稱
- `env` - 環境
- `severity` - 嚴重程度
- `namespace` - 命名空間
- `total_alerts` - 總警報數
- `firing_count` - 觸發中的警報數
- `resolved_count` - 已解決的警報數
- `alerts` - 警報列表
- `externalURL` - 外部連結

### 警報物件結構

每個警報物件包含：

- `status` - 警報狀態
- `labels` - 標籤 (包含 pod, env, namespace 等)
- `annotations` - 註解 (包含 summary 等)
- `startsAt` - 開始時間
- `endsAt` - 結束時間
- `generatorURL` - 生成器 URL

### 條件語法

```jinja2
{% if firing_count > 0 %}
🚨 *Alert Notification*
{% elif resolved_count > 0 %}
✅ *Alert Resolved*
{% endif %}
```

### 迴圈語法

```jinja2
{% for alert in alerts %}
{% if alert.status == "firing" %}
*Alert {{ loop.index }}:*
• Summary: {{ alert.annotations.summary }}
{% endif %}
{% endfor %}
```

### 函數

- `format_time(time_str)` - 格式化時間字符串

## 自定義模板

1. 複製現有模板檔案
2. 修改內容以符合您的需求
3. 重新啟動服務

## 範例

### 英文模板片段

```jinja2
{% if firing_count > 0 %}
🚨 *Alert Notification*

{% elif resolved_count > 0 %}
✅ *Alert Resolved*

{% endif %}
*Status:* {{ status }}
*Alert Name:* {{ alert_name }}
*Environment:* {{ env }}
*Severity:* {{ severity }}
*Namespace:* {{ namespace }}
*Total Alerts:* {{ total_alerts }}

{% if firing_count > 0 %}
*Firing:* {{ firing_count }}
{% endif %}
{% if resolved_count > 0 %}
*Resolved:* {{ resolved_count }}
{% endif %}
```

### 繁體中文模板片段

```jinja2
{% if firing_count > 0 %}
🚨 *警報通知*

{% elif resolved_count > 0 %}
✅ *警報已解決*

{% endif %}
*狀態:* {{ status }}
*警報名稱:* {{ alert_name }}
*環境:* {{ env }}
*嚴重程度:* {{ severity }}
*命名空間:* {{ namespace }}
*總警報數:* {{ total_alerts }}

{% if firing_count > 0 %}
*觸發中:* {{ firing_count }}
{% endif %}
{% if resolved_count > 0 %}
*已解決:* {{ resolved_count }}
{% endif %}
```

## 注意事項

1. **推薦使用 `.tmpl` 格式**: Go template 語法具有更好的性能和完整功能支援
2. **`.j2` 格式限制**: Jinja2 轉換是基於字串替換的簡化實現，僅支援基本語法
3. **轉換支援的語法**:
   - 基本變數引用
   - 簡單條件語句 (if/else/endif)
   - 基本迴圈語句 (for/endfor)
   - 常用的比較運算子
4. **不支援的 Jinja2 功能**:
   - 複雜的過濾器 (filters)
   - 巢狀迴圈
   - 複雜的條件表達式
   - 自定義函數
5. 如果模板載入失敗，系統會自動使用內建的模板邏輯
6. 修改模板後需要重新啟動服務
7. 支援的語言代碼：`eng` (英文)、`tw` (繁體中文)、`zh` (簡體中文)、`ja` (日文)、`ko` (韓文)

## 新增語言模板

1. 複製現有的模板檔案 (建議使用 `.tmpl` 格式)
2. 重新命名為 `tg_template_{language_code}.tmpl`
3. 翻譯模板內容為目標語言
4. 重新啟動服務以載入新模板
5. 測試新語言模板的功能

### 範例語言代碼

- `ja` - 日本語
- `ko` - 한국어
- `fr` - Français
- `de` - Deutsch
- `es` - Español
