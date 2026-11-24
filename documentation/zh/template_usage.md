# 模板配置使用指南

## Telegram 普通文字與 HTML 格式

Telegram 使用 HTML 模式傳送訊息。自訂模板中的外部普通文字應透過 `format_text` 輸出，例如：

```gotemplate
{{ format_text $.Platform (index $alert.Annotations "description") }}
```

在 Telegram 平台，`format_text` 會跳脫 `<`、`>`、`&` 與引號，避免 `(<5%)` 等告警內容被誤認為 HTML 標籤。輸入應為原始文字，不需預先跳脫；既有 `&lt;` 會顯示為字面文字 `&lt;`。

需要粗體、斜體、程式碼或連結時，使用既有的 `format_bold`、`format_italic`、`format_code`、`format_link`。傳入 `format_text` 的 HTML 標籤會顯示為普通文字；直接插值且未使用格式化函式的自訂欄位不會自動跳脫。不要對整則渲染結果再次跳脫，以免破壞模板產生的格式標籤。

其他平台的 `format_text` 維持原樣輸出。

## 1. 基本配置修改

### 隱藏超連結

編輯 `configs/telegram_config.yaml`：

```yaml
format_options:
  show_generator_url:
    enabled: false # 隱藏警報詳情連結
  show_external_url:
    enabled: false # 隱藏外部連結
```

### 隱藏表情符號

```yaml
format_options:
  show_emoji:
    enabled: false # 隱藏所有表情符號
```

### 隱藏時間戳

```yaml
format_options:
  show_timestamps:
    enabled: false # 隱藏時間信息
```

## 2. 使用不同配置檔案

### 創建極簡配置

複製 `configs/telegram_config.minimal.yaml`（已創建）並使用：

```go
// 在代碼中載入極簡配置
templateEngine := template.NewTemplateEngine()
err := templateEngine.LoadConfigWithProfile("minimal")
```

### 創建詳細配置

創建 `configs/telegram_config.detailed.yaml`：

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
    value: 500 # 更長的摘要
```

## 3. 動態切換配置

### 在運行時重新載入配置

```go
// 切換到極簡模式
err := templateEngine.ReloadConfigWithProfile("minimal")

// 切換回預設模式
err := templateEngine.ReloadConfigWithProfile("")
```

## 4. 配置效果對比

### 預設配置 (顯示所有元素)

```
🚨 *警報通知*
*狀態:* firing
*警報名稱:* TEST_Pod_CPU_Usage High
*環境:* uat
*嚴重程度:* critical
...
• 🔗 查看詳情: http://...
🔗 查看所有警報詳情: http://...
```

### 極簡配置 (隱藏連結和表情符號)

```
*警報通知*
*狀態:* firing
*警報名稱:* TEST_Pod_CPU_Usage High
*環境:* uat
*嚴重程度:* critical
...
```

## 5. 自定義配置選項

您可以在 `configs/telegram_config.yaml` 中調整以下選項：

- `show_links.enabled`: 是否顯示超連結
- `show_timestamps.enabled`: 是否顯示時間戳
- `show_external_url.enabled`: 是否顯示外部連結
- `show_generator_url.enabled`: 是否顯示生成器連結
- `show_emoji.enabled`: 是否顯示表情符號
- `compact_mode.enabled`: 是否使用緊湊模式
- `max_summary_length.value`: 摘要最大長度
