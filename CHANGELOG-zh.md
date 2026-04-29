# 變更日誌

本檔案記錄此專案所有重要變更。

格式基於 [Keep a Changelog](https://keepachangelog.com/zh-TW/1.1.0/)，
版本號遵循 [語意化版本](https://semver.org/lang/zh-TW/spec/v2.0.0.html)。

## [v2.0.6] - 2026-04-14

### 新增
- 整合 OpenTelemetry TracerProvider，支援 OTLP over HTTP 匯出 trace 資料至後端（Tempo、Jaeger 等）
- TraceConf 新增 `tlsSkipVerify` 欄位，支援 HTTPS + 跳過自簽憑證驗證場景
- otelgin middleware 加入 `WithFilter`，過濾 `/`、`/healthz`、`/healthy` 不產生 trace span
- Telegram、Slack、Discord 訊息發送自動建立 child span，可追蹤各平台發送延遲與錯誤
- 發送失敗時記錄 `span.RecordError`，可在 Grafana 以 `error=true` 篩選
- span 帶有 `messaging.system`、`messaging.level`、`messaging.channel` attribute
- 新增 `pkg/trace/provider.go`，封裝 TracerProvider 初始化與優雅關閉
- Helm Secret 加入 `trace-auth-user` / `trace-auth-passwd`，Deployment 注入 `TRACE_AUTH_USER` / `TRACE_AUTH_PASSWD` 環境變數
- Helm ConfigMap 加入 `tlsSkipVerify` 欄位
- Kustomize deployment 新增 `TRACE_AUTH_USER` / `TRACE_AUTH_PASSWD` 環境變數，對齊 Helm Chart
- Kustomize `config.production.yaml` 新增完整 `trace` 設定區段
- README/README-zh 新增可觀測性區段（Prometheus Metrics、OpenTelemetry Tracing）
- `config.expamle.yaml` 新增完整 `trace` 設定區段（含全部 9 個欄位）
- `kubernetes-env-vars.md`（中英文）補上 `TRACE_AUTH_USER`、`TRACE_AUTH_PASSWD`、`DISCORD_TOKEN` 環境變數
- 中英文 `config_guide.md` 加入 OpenTelemetry Tracing 設定說明

### 變更
- OTLP exporter 改用 `WithEndpointURL` 明確帶上 `http://` 或 `https://` scheme，取代 `WithEndpoint` + `WithInsecure()` 組合
- 新增全域變數 `var Trace TraceConf`，修正 `updateGlobalConfigs` 缺少的 Trace 同步
- 修正 `AppName`、`TrustedProxies` mapstructure tag 與 YAML 不一致
- Dockerfile 改用 `$BUILDPLATFORM` + Go cross-compile，解決 ARM Mac buildx QEMU segfault
- 路由註冊邏輯更新，註冊前檢查服務可用性

### 修正
- 修正 OTLP HTTP exporter 在 `insecure: true` 時仍走 HTTPS 導致 `tls: failed to verify certificate` 錯誤
- 修正 ELB health check（`GET /`）產生無用 trace span 的問題
- 修正 `serviceName` 為空（AppName tag 不匹配 YAML）
- 修正 OTEL resource schema URL 衝突（`resource.Merge` → `resource.New`）
- 修正 `config_guide.md`（中英文）`insecure` 欄位描述與程式碼行為相反的問題
- 補上 `config_guide.md`（中英文）缺少的 `tlsSkipVerify` 欄位

## [v2.0.4] - 2025-10-31

### 修正
- 修正警報等級路由問題

## [v2.0.3] - 2025-10-31

### 新增
- 更新 GitLab CI 支援 tag 觸發 pipeline

## [v2.0.2] - 2025-10-31

### 修正
- 修改 Telegram Level 4（測試群組）路由問題

## [v2.0.1] - 2025-10-31

### 修正
- 修改 Telegram Level 4（測試群組）路由問題（同 v2.0.2）

## [v2.0] - 2025-10-31

### 修正
- 修改 Telegram Level 4（測試群組）路由問題

## [v1.9] - 2025-09-17

### 修正
- 修正 template 多餘空白問題

## [v1.8] - 2025-09-16

### 新增
- 新增 Debug Telegram server 回應訊息

### 修正
- 修正 template 渲染問題

## [v1.7] - 2025-09-12

### 新增
- 新增 debug 顯示接收到的訊息資料
- 新增 Kustomize 部署配置
- 新增 Helm chart 與 Kustomization ConfigMap 支援

### 修正
- 修正 template 顯示問題，新增接收訊息資料記錄
- 修正外部連結顯示問題
- 修正 namespace 值配置
- 更新容器映像倉庫參照

## [v1.6] - 2025-09-12

### 新增
- 新增模板描述
- 新增 Helm chart 部署支援
- 合併 `feature/alert-webhooks-fix` 功能分支

### 修正
- 修正 CORS 配置
- 修正 TOC（目錄）恢復問題
- 修正外部連結顯示問題

## [v1.5] - 2025-09-08

### 新增
- 新增專案描述與文件

## [v1.4] - 2025-09-03

### 修正
- 修正 TOC（目錄）恢復問題
- 配置更新與改進

## [v1.3] - 2025-09-02

### 新增
- 新增 Helm chart 部署配置
- 新增 `.gitignore` 規則

### 修正
- 修正外部連結顯示問題

## [v1.2] - 2025-08-26

### 修正
- 修正 CORS（跨來源資源共享）配置

## [v1.1] - 2025-08-26

### 新增
- 新增 ECR（Elastic Container Registry）倉庫配置

### 修正
- 修正 CORS 配置

## [v1.0] - 2025-08-25

### 新增
- 首次發布
- AlertManager webhook 處理服務
- Telegram 通知支援，含多等級聊天群組路由
- HTTP Basic Auth 認證保護
- 多語言模板支援（英語、繁體中文、簡體中文、日語、韓語）
- Full 和 Minimal 兩種模板顯示模式
- Firing 和 Resolved 警報分離通知
- Swagger API 文件
- GitHub Actions CI/CD 流程
- Docker 和 Docker Compose 支援
