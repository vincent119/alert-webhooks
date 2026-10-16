# Kustomize 安裝指南

本指南使用 repository 內的 Kustomize overlay 部署 Alert Webhooks。除非段落另有說明，所有命令都從 repository root 執行。

## 前置條件

- 可連線的 Kubernetes cluster，以及具備目標 namespace 權限的 kubeconfig。
- kubectl。kubectl 自 Kubernetes 1.14 起內建 Kustomize，可直接使用 `kubectl kustomize` 與 `kubectl apply -k`。
- 選配：獨立的 Kustomize CLI。

macOS 可使用 Homebrew 安裝：

```bash
brew install kubectl
brew install kustomize
kubectl version --client
kubectl kustomize --help
```

其他作業系統請參考 [kubectl 安裝文件](https://kubernetes.io/docs/tasks/tools/) 與 [Kustomize 官方說明](https://kubectl.docs.kubernetes.io/installation/kustomize/)。

## 部署前檢查

### 1. 確認 Kubernetes context

```bash
kubectl config current-context
kubectl auth can-i create deployments --namespace alert-webhooks
```

如果 context 或權限不符合預期，請先停止，不要繼續部署。

### 2. 檢查 image 與 Ingress

部署前檢查 [`kustomization.yaml`](./kustomization.yaml)：

- `images[].newName` 與 `images[].newTag` 必須指向可由 cluster 拉取的 image。
- Ingress certificate ARN 與 host 必須替換為目標環境值。
- 若不使用 AWS Load Balancer Controller，必須依目標 ingress controller 調整或停用 Ingress。

不要把私有 registry credential、certificate 或內部 domain 寫入公開 repository。環境專屬值應由受控 overlay 或部署流程管理。

### 3. 建立必要 Secret

Deployment 需要 namespace `alert-webhooks` 內的 `alert-webhooks-secrets`。先建立 namespace，再以受控的 Secret 管理方式建立 Secret。以下指令只顯示 placeholder；執行前必須由安全來源替換：

```bash
kubectl create namespace alert-webhooks --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic alert-webhooks-secrets \
  --namespace alert-webhooks \
  --from-literal=metric-user="REPLACE_METRIC_USER" \
  --from-literal=metric-password="REPLACE_METRIC_PASSWORD" \
  --from-literal=webhooks-user="REPLACE_WEBHOOKS_USER" \
  --from-literal=webhooks-password="REPLACE_WEBHOOKS_PASSWORD" \
  --from-literal=telegram-token="REPLACE_TELEGRAM_TOKEN" \
  --from-literal=slack-token="REPLACE_SLACK_TOKEN" \
  --from-literal=discord-token="REPLACE_DISCORD_TOKEN" \
  --from-literal=trace-auth-user="REPLACE_TRACE_AUTH_USER" \
  --from-literal=trace-auth-passwd="REPLACE_TRACE_AUTH_PASSWORD" \
  --dry-run=client -o yaml | kubectl apply -f -
```

不要把實際 Secret 寫進 `kustomization.yaml`、shell history、CI log 或 repository。正式環境應優先使用既有 Secret 管理方案。

## Render 與差異預覽

先在本機 render，確認沒有 placeholder、錯誤 image 或非預期資源：

```bash
kubectl kustomize kubernetes/kustomize > /tmp/alert-webhooks-kustomize.yaml
```

有 cluster 權限時，可在套用前查看差異：

```bash
kubectl diff -k kubernetes/kustomize
```

`kubectl diff` 可能在存在差異時回傳非 0，應閱讀輸出後再決定是否部署。

## 部署

```bash
kubectl apply -k kubernetes/kustomize
```

## 驗證部署

```bash
kubectl get pods --namespace alert-webhooks
kubectl get service --namespace alert-webhooks
kubectl get deployment --namespace alert-webhooks
kubectl get ingress --namespace alert-webhooks
kubectl rollout status deployment/prod-alert-webhooks \
  --namespace alert-webhooks \
  --timeout 5m
kubectl logs --namespace alert-webhooks --selector app=alert-webhooks
```

## 更新部署

更新 [`kustomization.yaml`](./kustomization.yaml) 的 image tag 或受控 overlay 後，先 render、檢查差異，再重新套用：

```bash
kubectl kustomize kubernetes/kustomize > /tmp/alert-webhooks-kustomize.yaml
kubectl diff -k kubernetes/kustomize
kubectl apply -k kubernetes/kustomize
kubectl rollout status deployment/prod-alert-webhooks \
  --namespace alert-webhooks \
  --timeout 5m
```

## 疑難排解

```bash
kubectl get events --namespace alert-webhooks --sort-by='.lastTimestamp'
kubectl describe deployment prod-alert-webhooks --namespace alert-webhooks
kubectl describe pod --namespace alert-webhooks --selector app=alert-webhooks
kubectl get configmap prod-alert-webhooks-config --namespace alert-webhooks -o yaml
kubectl get secret alert-webhooks-secrets --namespace alert-webhooks
```

查看 Secret 時不要使用會輸出 Secret data 的參數，也不要把輸出貼到 issue、chat 或 CI log。

## 清理資源

先預覽將刪除的資源：

```bash
kubectl delete -k kubernetes/kustomize --dry-run=client -o name
```

確認後刪除 Kustomize 管理的資源：

```bash
kubectl delete -k kubernetes/kustomize
```

外部建立的 `alert-webhooks-secrets` 不屬於 Kustomize 管理範圍，需要時必須另行刪除。刪除 namespace 前，先確認其中沒有其他工作負載。

## 目錄結構

```text
kubernetes/kustomize/
├── README.md
├── kustomization.yaml
├── base/
│   ├── kustomization.yaml
│   ├── namespace.yaml
│   ├── deployment.yaml
│   ├── service.yaml
│   └── ingress.yaml
└── configs/
    ├── config.production.yaml
    ├── alert_config.yaml
    └── alert_config.minimal.yaml
```
