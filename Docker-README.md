# Docker 部署指南

本指南說明如何使用 Docker 部署 Alert-Webhooks 服務。

## 🐳 Docker 文件說明

### Dockerfile

- **多階段構建**：分為 builder 和 runtime 階段
- **安全性**：使用非 root 用戶運行
- **優化**：靜態編譯，最小化映像大小
- **健康檢查**：內建健康檢查端點

### docker-compose.yml

- **生產部署**：完整的生產環境配置
- **資源限制**：設定記憶體和 CPU 限制
- **安全設定**：只讀文件系統，用戶隔離
- **網路隔離**：專用網路

### docker-compose.dev.yml

- **開發環境**：包含開發工具和調試端點
- **熱重載**：掛載源代碼目錄
- **調試支援**：開放 Delve 調試端口

## 🚀 快速開始

### 1. 構建映像

```bash
# 構建生產映像
make docker-build

# 或手動構建
docker build -t alert-webhooks:latest .
```

### 2. 運行容器

```bash
# 使用 docker-compose（推薦）
make docker-run

# 或手動運行
docker run -d \
  --name alert-webhooks \
  -p 9999:9999 \
  -e TELEGRAM_TOKEN="your-token" \
  -e SLACK_TOKEN="your-token" \
  alert-webhooks:latest
```

### 3. 查看日誌

```bash
# 查看日誌
make docker-logs

# 或
docker-compose logs -f
```

## ⚙️ 環境變數配置

### 必需的環境變數

```bash
# 服務 Token
TELEGRAM_TOKEN=your-telegram-bot-token
SLACK_TOKEN=your-slack-bot-token

# 認證資訊
METRIC_USER=your-metric-user
METRIC_PASSWORD=your-metric-password
WEBHOOKS_USER=your-webhook-user
WEBHOOKS_PASSWORD=your-webhook-password
```

### 可選的環境變數

```bash
# 應用設定
APP_ENV=production
GIN_MODE=release
LOG_LEVEL=info

# 其他設定
TZ=Asia/Taipei
```

## 📁 數據卷掛載

### 配置文件掛載

```yaml
volumes:
  - ./configs:/app/configs:ro
  - ./templates:/app/templates:ro
```

### 日誌持久化

```yaml
volumes:
  - ./logs:/app/logs
```

## 🔍 健康檢查

容器內建健康檢查：

```bash
# 檢查健康狀態
curl -f http://localhost:9999/healthz

# Docker 健康檢查
docker ps  # 查看 healthy/unhealthy 狀態
```

## 🛠️ 開發環境

### 啟動開發環境

```bash
# 使用開發配置
make docker-dev

# 查看開發日誌
make docker-logs-dev
```

### 進入容器調試

```bash
# 進入容器 shell
make docker-shell

# 或
docker exec -it alert-webhooks /bin/sh
```

## 🎯 Kubernetes 部署

### 創建 Secret

```bash
kubectl create secret generic alert-webhooks-secrets \
  --from-literal=telegram-token="your-telegram-token" \
  --from-literal=slack-token="your-slack-token" \
  --from-literal=metric-user="admin" \
  --from-literal=metric-password="password" \
  --from-literal=webhooks-user="admin" \
  --from-literal=webhooks-password="password"
```

### 部署應用

```bash
# 使用提供的 Kubernetes 部署文件
kubectl apply -f kubernetes/deployment-example.yaml
```

### 查看部署狀態

```bash
# 查看 Pod 狀態
kubectl get pods -l app=alert-webhooks

# 查看日誌
kubectl logs deployment/alert-webhooks

# 查看服務
kubectl get svc alert-webhooks-service
```

## 📊 監控和日誌

### Prometheus 指標

```bash
# 訪問指標端點
curl -u admin:admin http://localhost:9999/api/v1/metrics
```

### 日誌管理

```bash
# 實時查看日誌
docker-compose logs -f alert-webhooks

# 查看特定時間範圍的日誌
docker logs --since="2h" alert-webhooks
```

## 🔧 故障排除

### 常見問題

#### 1. 容器無法啟動

```bash
# 查看容器日誌
docker logs alert-webhooks

# 檢查配置
docker exec -it alert-webhooks cat /app/configs/config.production.yaml
```

#### 2. 環境變數未生效

```bash
# 檢查環境變數
docker exec -it alert-webhooks env | grep -E "(TELEGRAM|SLACK|WEBHOOK)"
```

#### 3. 健康檢查失敗

```bash
# 手動測試健康檢查
docker exec -it alert-webhooks curl -f http://localhost:9999/healthz
```

#### 4. 權限問題

```bash
# 檢查文件權限
docker exec -it alert-webhooks ls -la /app/
```

### 調試模式

```bash
# 使用調試模式運行
docker run -it \
  -e LOG_LEVEL=debug \
  -e GIN_MODE=debug \
  alert-webhooks:latest
```

## 🚨 安全考量

### 生產環境建議

1. **使用 Secrets 管理敏感資訊**
2. **啟用只讀文件系統**
3. **使用非 root 用戶**
4. **設定資源限制**
5. **定期更新基礎映像**

### 網路安全

```yaml
# 限制網路訪問
networks:
  alert-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 📈 效能優化

### 映像大小優化

- 使用 Alpine Linux 基礎映像
- 多階段構建去除構建工具
- 靜態編譯減少依賴

### 運行時優化

```yaml
# 資源限制
deploy:
  resources:
    limits:
      memory: 256M
      cpus: "0.5"
    reservations:
      memory: 128M
      cpus: "0.1"
```

## 🔄 CI/CD 整合

### 構建流程

```bash
# 1. 構建映像
docker build -t alert-webhooks:${VERSION} .

# 2. 運行測試
docker run --rm alert-webhooks:${VERSION} /app/alert-webhooks --version

# 3. 推送到倉庫
docker push alert-webhooks:${VERSION}
```

### 自動部署

```yaml
# GitHub Actions 範例
- name: Build and push Docker image
  uses: docker/build-push-action@v2
  with:
    context: .
    push: true
    tags: your-registry/alert-webhooks:latest
```

## 📚 相關文檔

- [Kubernetes 環境變數配置](documentation/zh/kubernetes-env-vars.md)
- [服務啟用配置指南](documentation/zh/service-enable-config.md)
- [Slack 設定指南](documentation/zh/slack_setup.md)
- [Telegram 設定指南](documentation/zh/telegram_setup.md)
