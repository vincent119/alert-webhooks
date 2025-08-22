# Kubernetes 環境變數配置指南

本指南說明如何在 Kubernetes 中使用環境變數來覆蓋 `config.yaml` 中的敏感配置。

## 支援的環境變數

系統會按照以下優先級讀取配置：

1. **環境變數** (最高優先級)
2. `config.yaml` 檔案 (備用)

## 認證系統說明

系統使用兩套獨立的認證系統：

1. **Metrics 認證** (`metric.user/password`)：
   - 用於 `/api/v1/metrics` 端點（Prometheus 指標）
   - 只影響監控指標的存取權限

2. **Webhooks 認證** (`webhooks.base_auth_user/password`)：
   - 用於所有 Telegram/Slack API 端點
   - 控制警報通知 API 的存取權限
   - 只有在 `webhooks.enable: true` 時才啟用認證

### 支援的環境變數列表

| 環境變數名稱        | 對應配置項                    | 說明                                    |
| ------------------- | ----------------------------- | --------------------------------------- |
| `METRIC_USER`       | `metric.user`                 | Prometheus metrics 端點的認證用戶名     |
| `METRIC_PASSWORD`   | `metric.password`             | Prometheus metrics 端點的認證密碼       |
| `WEBHOOKS_USER`     | `webhooks.base_auth_user`     | Telegram/Slack API 端點的認證用戶名     |
| `WEBHOOKS_PASSWORD` | `webhooks.base_auth_password` | Telegram/Slack API 端點的認證密碼       |
| `TELEGRAM_TOKEN`    | `telegram.token`              | Telegram Bot 的 API Token               |
| `SLACK_TOKEN`       | `slack.token`                 | Slack Bot 的 API Token                  |

## Kubernetes 部署示例

### 1. 創建 Secret

```bash
# 方法 1：使用 kubectl 創建 secret
kubectl create secret generic alert-webhooks-secrets \
  --from-literal=metric-user="your-metric-user" \
  --from-literal=metric-password="your-metric-password" \
  --from-literal=webhooks-user="your-webhook-user" \
  --from-literal=webhooks-password="your-webhook-password" \
  --from-literal=telegram-token="your-telegram-bot-token" \
  --from-literal=slack-token="your-slack-bot-token" \
  --namespace monitoring

# 方法 2：手動編碼並應用 YAML
echo -n "your-value" | base64
```

### 2. 在 Deployment 中引用 Secret

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: alert-webhooks
spec:
  template:
    spec:
      containers:
        - name: alert-webhooks
          env:
            - name: METRIC_USER
              valueFrom:
                secretKeyRef:
                  name: alert-webhooks-secrets
                  key: metric-user
            - name: METRIC_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: alert-webhooks-secrets
                  key: metric-password
          # ... 其他環境變數
```

## 配置檔案範例

### config.production.yaml

```yaml
metric:
  # 如果沒有設定環境變數，會使用以下預設值
  user: "default-user"
  password: "default-password"

webhooks:
  # enable 欄位預設為 false，如果要啟用請明確設定為 true
  enable: true
  # 如果沒有設定環境變數，會使用以下預設值
  base_auth_user: "default-user"
  base_auth_password: "default-password"

telegram:
  # enable 欄位預設為 false，如果要啟用請明確設定為 true
  enable: true
  # 如果沒有設定 TELEGRAM_TOKEN 環境變數，會使用以下預設值
  token: "default-token"

slack:
  # enable 欄位預設為 false，如果要啟用請明確設定為 true
  enable: true
  # 如果沒有設定 SLACK_TOKEN 環境變數，會使用以下預設值
  token: "default-token"
```

## 服務啟用邏輯

**重要**: 所有服務（webhooks、telegram、slack）的 `enable` 欄位預設值為 `false`。

- **如果配置檔案中沒有 `enable` 欄位**：服務將被停用 (disable)
- **如果配置檔案中 `enable` 欄位為空值**：服務將被停用 (disable)
- **如果配置檔案中 `enable: false`**：服務將被停用 (disable)
- **如果配置檔案中 `enable: true`**：服務將被啟用 (enable)

這樣的設計確保了安全性 - 只有明確配置啟用的服務才會運行。

## 安全最佳實踐

1. **使用 Kubernetes Secrets**：所有敏感資訊都應該儲存在 Kubernetes Secrets 中
2. **限制 Secret 存取權限**：使用 RBAC 控制誰可以存取這些 Secrets
3. **加密靜態資料**：確保 etcd 中的 Secrets 資料已加密
4. **定期輪換 Token**：定期更新 Bot Token 和密碼

## 驗證配置

部署後，您可以通過以下方式驗證環境變數是否正確載入：

1. 檢查應用程式日誌：

```bash
kubectl logs deployment/alert-webhooks -n monitoring
```

2. 查找以下日誌訊息：

```
Override metric user from env var: your-user
Override metric password from env var: [REDACTED]
Override telegram token from env var: [REDACTED]
Override slack token from env var: [REDACTED]
Override webhooks user from env var: your-user
Override webhooks password from env var: [REDACTED]
```

## 故障排除

### 環境變數未生效

1. 檢查 Secret 是否存在：`kubectl get secrets -n monitoring`
2. 檢查 Secret 的內容：`kubectl describe secret alert-webhooks-secrets -n monitoring`
3. 檢查 Pod 的環境變數：`kubectl exec -it <pod-name> -- env | grep -E "(METRIC|WEBHOOK|TELEGRAM|SLACK)"`

### Pod 無法啟動

1. 檢查 Secret 引用是否正確
2. 檢查 ConfigMap 是否存在
3. 查看 Pod 事件：`kubectl describe pod <pod-name> -n monitoring`

## 完整範例

查看 `kubernetes/deployment-example.yaml` 檔案以獲取完整的 Kubernetes 部署範例。

## 相關文件

- [Kubernetes 基本配置](./kubernetes-basic.md)
- [Telegram 設定指南](./telegram_setup.md)
- [Slack 設定指南](./slack_setup.md)
