# Helm 安裝指南

本指南說明如何從公開 Helm repository 安裝、升級、驗證與回滾 Alert Webhooks。除非段落另有說明，從原始碼驗證的命令都應在 repository root 執行。

## 前置條件

- 可連線的 Kubernetes cluster，以及具備目標 namespace 權限的 kubeconfig。
- Helm 與 kubectl。macOS 可使用 Homebrew 安裝：

```bash
brew install helm kubectl
helm version
kubectl version --client
kubectl config current-context
```

其他作業系統請參考 [Helm 安裝文件](https://helm.sh/docs/intro/install/) 與 [kubectl 安裝文件](https://kubernetes.io/docs/tasks/tools/)。執行部署前，必須確認 `kubectl config current-context` 指向正確的 cluster。

## 安全設定

Chart 不提供 webhook、metric 或通知服務 credential 的預設值。安裝時必須以私有 values 檔提供 `secrets.metric.user`、`secrets.metric.password`、`secrets.webhooks.user` 與 `secrets.webhooks.password`。

當 `config.telegram.enable`、`config.slack.enable` 或 `config.discord.enable` 為 `true` 時，必須分別提供對應的 `secrets.<service>.token`。Trace 認證維持選填。

請勿將實際 credential 提交至 repository，也不要透過 `--set` 把 Secret 留在 shell history 或 CI log。私有 values 檔應放在 repository 外，並限制檔案權限。

以下最小範例停用所有通知服務；需要啟用時再於私有 values 檔加入對應 Token：

```yaml
ingress:
  enabled: false

config:
  telegram:
    enable: false
  slack:
    enable: false
  discord:
    enable: false

secrets:
  metric:
    user: "REPLACE_METRIC_USER"
    password: "REPLACE_METRIC_PASSWORD"
  webhooks:
    user: "REPLACE_WEBHOOKS_USER"
    password: "REPLACE_WEBHOOKS_PASSWORD"
```

## 版本契約

- Chart version 與 application version 分開管理。
- `chart-v<Chart version>` tag 只發布對應的不可變 Chart package。
- 本版本的 Chart version 為 `1.0.1`，`appVersion` 與預設 container image tag 為 `v2.0.8`。
- 已發布的 Chart version 不會被覆寫，公開 repository 也不提供可變的 `latest` Chart。

## 從公開 repository 安裝

先加入公開 Helm repository，並固定使用所需 Chart version：

```bash
helm repo add alert-webhooks https://vincent119.github.io/alert-webhooks
helm repo update
helm show chart alert-webhooks/alert-webhooks-chart --version 1.0.1
```

指定 repository 外的私有 values 檔路徑：

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
```

先執行 render 驗證，不連線修改 cluster：

```bash
helm template alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.1 \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

檢查 render 結果是否包含 placeholder、非預期 image reference 或 Secret，再繼續安裝或升級：

```bash
helm upgrade --install alert-webhooks alert-webhooks/alert-webhooks-chart \
  --version 1.0.1 \
  --namespace alert-webhooks \
  --create-namespace \
  --values "${PRIVATE_VALUES_FILE}" \
  --atomic \
  --wait \
  --timeout 5m
```

## 從原始碼驗證

```bash
PRIVATE_VALUES_FILE=/secure/path/alert-webhooks-values.yaml
CHART_DIR=kubernetes/helm/alert-webhooks-chart

helm lint "${CHART_DIR}" --values "${PRIVATE_VALUES_FILE}"
helm template alert-webhooks "${CHART_DIR}" \
  --namespace alert-webhooks \
  --values "${PRIVATE_VALUES_FILE}" \
  > /tmp/alert-webhooks-rendered.yaml
```

## 驗證部署

```bash
helm status alert-webhooks --namespace alert-webhooks
kubectl get pods,service,ingress --namespace alert-webhooks
kubectl rollout status deployment \
  --namespace alert-webhooks \
  --selector app.kubernetes.io/instance=alert-webhooks \
  --timeout 5m
```

必須等 rollout 完成且 Pod Ready 後，才能繼續測試通知功能。

## 回滾與移除

```bash
helm history alert-webhooks --namespace alert-webhooks
REVISION_NUMBER=1
helm rollback alert-webhooks "${REVISION_NUMBER}" \
  --namespace alert-webhooks \
  --wait \
  --timeout 5m
```

移除 release：

```bash
helm uninstall alert-webhooks --namespace alert-webhooks
```

`helm uninstall` 不會自動刪除 namespace。刪除 namespace 前，先確認其中沒有其他 workload 或必須保留的 resource。

## Chart values 與文件維護

完整 values contract 請參考 [`values.yaml`](./values.yaml) 與主要英文 [`README.md`](./README.md) 產生的 values table。人工內容維護於 `README.md.gotmpl`，values table 由固定版本的 `helm-docs` 產生：

```bash
make helm-docs
```

請勿手動修改產生的 values table。

## 相關文件

- [Docker 部署](../../../documentation/zh/docker-deployment.md)
- [Kubernetes 環境變數](../../../documentation/zh/kubernetes-env-vars.md)
- [Kustomize 安裝](../../kustomize/README.md)
