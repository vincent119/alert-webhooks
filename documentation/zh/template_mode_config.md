# Template Mode 配置說明

## 📋 配置方式

在 `configs/config.development.yaml` 中設定：

```yaml
telegram:
  template_mode: "minimal" # 可選值: "minimal", "full"
```

## 🎯 支援的模式

### 1. **"minimal" 模式**

- 簡潔的訊息格式
- 隱藏超連結
- 隱藏時間戳
- 隱藏表情符號
- 使用緊湊模式

### 2. **"full" 模式** (預設)

- 完整的訊息格式
- 顯示所有超連結
- 顯示時間戳
- 顯示表情符號
- 詳細資訊模式

## 🔄 動態切換

修改配置檔案後，重新啟動服務即可生效：

```bash
# 修改 configs/config.development.yaml
template_mode: "minimal"

# 重新啟動服務
go run cmd/main.go -e development
```

## 📱 訊息效果對比

### **"full" 模式效果:**

```
🚨 *警報通知*
*狀態:* firing
*警報名稱:* TEST_Pod_CPU_Usage High
*環境:* uat
*嚴重程度:* critical
*命名空間:* hcgateconsole
*總警報數:* 1
*觸發中:* 1
*🚨 觸發中的警報:*
*警報 1:*
• 摘要: uat Pod CPU_Usage: 6.12% > 80%
• Pod: hcgateconsole-deploy-6fdfbc5565-4d7vr
• 開始時間: 2022-11-18 08:17:31
• 🔗 查看詳情: http://prometheus-986dbd5cd-c5dlj:9090/graph?...
🔗 查看所有警報詳情: http://prometheus-alertmanager-7dd5f4895f-7wpml:9093
```

### **"minimal" 模式效果:**

```
*警報通知*
*狀態:* firing
*警報名稱:* TEST_Pod_CPU_Usage High
*環境:* uat
*嚴重程度:* critical
*命名空間:* hcgateconsole
*總警報數:* 1
*觸發中:* 1
*觸發中的警報:*
*警報 1:*
• 摘要: uat Pod CPU_Usage: 6.12% > 80%
• Pod: hcgateconsole-deploy-6fdfbc5565-4d7vr
```

## 🛠️ 自定義擴展

如果您需要更多模式，可以：

1. 創建新的配置檔案，如 `telegram_config.custom.yaml`
2. 在代碼中添加新的條件判斷
3. 支援更多的 template_mode 值

### 擴展範例：

```go
// 在 handler.go 中添加更多模式
switch templateMode {
case "minimal":
    err = templateEngine.LoadConfigWithProfile("minimal")
case "custom":
    err = templateEngine.LoadConfigWithProfile("custom")
case "debug":
    err = templateEngine.LoadConfigWithProfile("debug")
default:
    err = templateEngine.LoadConfigFromConfigs()
}
```
