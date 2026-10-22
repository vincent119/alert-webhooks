# Docker 部署指南

本指南說明如何使用 Docker Hub 的公開 image 執行固定版本的
Alert Webhooks，涵蓋單機 Docker 部署、健康驗證、例行操作、回滾與公開
image 發布規範。

Kubernetes 部署請使用目前維護的
[Helm 指南](../../kubernetes/helm/alert-webhooks-chart/README.zh-TW.md)，或
[Kustomize 指南](../../kubernetes/kustomize/README.md)。

## 前置條件

- 可存取 Docker Hub 的 Docker Engine。
- UID `1500` 可讀取的 Alert Webhooks 本機設定檔。
- 主機的 `9999` port 尚未被占用。
- 若使用內附的 `examples/config.expamle.yaml`，需要先取得 repository。

不得將真實 token 或密碼存入 repository、image、shell history 或 CI log。
支援的設定與環境變數覆寫方式請參考[設定指南](./config_guide.md)。

## 公開 image 發布規範

- Repository：`docker.io/vincent119/alert-webhooks`。
- 目前 release tag：`v2.0.10`。
- 平台：`linux/amd64`、`linux/arm64`。
- Runtime user：非 root UID `1500`。
- 健康端點：port `9999` 的 `/healthy`。
- 固定版本 tag 不可覆寫，不發布 Docker `latest`。
- Docker Hub 是公開發布來源，private ECR production 流程維持不變。

`v2.0.10` manifest digest：

```text
sha256:ce8726a1007685adb18860b83b4bbb179553cc1771a0d0bbc5b52e98d433c3da
```

發布證據請參考
[v2.0.10 GitHub Release](https://github.com/vincent119/alert-webhooks/releases/tag/v2.0.10)。

## 啟動容器

使用內附範例設定時，請在 repository 根目錄執行：

```bash
image="docker.io/vincent119/alert-webhooks:v2.0.10"
config_file="$(pwd)/examples/config.expamle.yaml"

docker pull "${image}"
docker run --detach \
  --name alert-webhooks \
  --publish 127.0.0.1:9999:9999 \
  --volume "${config_file}:/app/configs/config.production.yaml:ro" \
  --env APP_ENV=production \
  "${image}"
```

若設定檔位於 repository 以外，請將 `config_file` 改為該檔案的絕對路徑。
啟用通知整合前，必須先檢查並保護設定內容。

## 驗證部署

檢查公開健康端點：

```bash
curl --fail http://127.0.0.1:9999/healthy
```

成功時 HTTP 狀態碼為 `200`、`status` 為 `ok`，並回傳設定檔載入的版本。

確認處理程序使用預期的非 root 使用者：

```bash
docker exec alert-webhooks id -u
```

預期輸出：

```text
1500
```

檢查 Docker 健康狀態與應用程式日誌：

```bash
docker inspect --format '{{.State.Health.Status}}' alert-webhooks
docker logs --follow alert-webhooks
```

如果健康端點未回傳成功或 UID 不是 `1500`，應停止後續操作並檢查日誌。

## 例行操作

### 查看日誌

```bash
docker logs --since 2h alert-webhooks
```

### 停止與啟動

```bash
docker stop alert-webhooks
docker start alert-webhooks
```

### 以另一個核准版本替換 container

1. 記錄目前 image tag 與設定檔路徑。
2. 拉取核准的固定版本 tag。
3. 停止並移除既有容器。
4. 使用相同明確設定啟動新版本。
5. 重新驗證健康端點與 UID。

移除容器不會刪除掛載的設定檔。替換前必須確認需要保留的日誌已存放在容器外。

## 回滾

由於 release tag 不可覆寫，回滾使用前一個已核准 tag：

1. 停止並移除未成功的容器。
2. 使用相同設定啟動前一版 image tag。
3. 驗證 `/healthy`、UID `1500` 與應用程式日誌。
4. 保留未成功版本供調查，不得覆寫或重用其 tag。

## 故障排除

### 容器無法啟動

```bash
docker logs alert-webhooks
docker inspect alert-webhooks
```

確認掛載檔案存在，且 UID `1500` 可讀取：

```bash
ls -l "${config_file}"
```

### 健康檢查失敗

```bash
docker exec alert-webhooks \
  curl --fail http://127.0.0.1:9999/healthy
```

確認服務監聽 port `9999`，且掛載檔案可由
`/app/configs/config.production.yaml` 讀取。

### Host port 已被占用

可改用其他 loopback port，container port 維持不變：

```bash
--publish 127.0.0.1:19999:9999
```

對應的主機健康檢查 URL 會改為 `http://127.0.0.1:19999/healthy`。

## 本機開發 image

本機建置應使用明確的本機 image tag，避免與公開 release 混淆：

```bash
docker build --tag alert-webhooks:dev .
```

公開 image 發布 workflow 定義於
[publish-container-image.yml](../../.github/workflows/publish-container-image.yml)。
只有有效的 GitHub `v<major>.<minor>.<patch>` tag 或授權的手動執行可以發布 image；
一般 branch push、pull request 與合併不會發布至 Docker Hub。

## 相關文件

- [設定指南](./config_guide.md)
- [Kubernetes 環境變數](./kubernetes-env-vars.md)
- [Helm 安裝](../../kubernetes/helm/alert-webhooks-chart/README.zh-TW.md)
- [Kustomize 安裝](../../kubernetes/kustomize/README.md)
