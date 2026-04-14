# Changelog

## [v2.0.6] - 2026-04-14

### Added
- TraceConf 新增 `tlsSkipVerify` 欄位，支援 HTTPS + 跳過自簽憑證驗證場景
- otelgin middleware 加入 `WithFilter`，過濾 `/`、`/healthz`、`/healthy` 不產生 trace span
- Telegram `SendMessage` 傳遞 request context，使 Telegram API 呼叫成為 HTTP request span 的 child span
- Helm Secret 加入 `trace-auth-user` / `trace-auth-passwd`，Deployment 注入 `TRACE_AUTH_USER` / `TRACE_AUTH_PASSWD` 環境變數
- Helm ConfigMap 加入 `tlsSkipVerify` 欄位

### Changed
- OTLP exporter 改用 `WithEndpointURL` 明確帶上 `http://` 或 `https://` scheme，取代 `WithEndpoint` + `WithInsecure()` 組合，解決 insecure 設定不生效的問題

### Fixed
- 修正 OTLP HTTP exporter 在 `insecure: true` 時仍走 HTTPS 導致 `tls: failed to verify certificate` 錯誤
- 修正 ELB health check（`GET /`）產生無用 trace span 的問題

---

## [v2.0.6] - 2026-04-14 (English)

### Added
- Added `tlsSkipVerify` field to TraceConf for HTTPS with self-signed certificate skip
- Added `WithFilter` to otelgin middleware to exclude `/`, `/healthz`, `/healthy` from tracing
- Telegram `SendMessage` now propagates request context for proper trace span chaining
- Helm Secret includes `trace-auth-user` / `trace-auth-passwd`, Deployment injects `TRACE_AUTH_USER` / `TRACE_AUTH_PASSWD` env vars
- Helm ConfigMap includes `tlsSkipVerify` field

### Changed
- OTLP exporter switched from `WithEndpoint` + `WithInsecure()` to `WithEndpointURL` with explicit `http://` or `https://` scheme to ensure correct protocol selection

### Fixed
- Fixed OTLP HTTP exporter using HTTPS despite `insecure: true`, causing `tls: failed to verify certificate` error
- Fixed ELB health check (`GET /`) generating unnecessary trace spans

---

## [v2.0.5] - 2026-04-14

### Added
- 整合 OpenTelemetry TracerProvider，支援 OTLP over HTTP 匯出 trace 資料至後端（Tempo、Jaeger 等）
- Telegram、Slack、Discord 訊息發送自動建立 child span，可追蹤各平台發送延遲與錯誤
- 發送失敗時記錄 `span.RecordError`，可在 Grafana 以 `error=true` 篩選
- span 帶有 `messaging.system`、`messaging.level`、`messaging.channel` attribute
- 新增 `pkg/trace/provider.go`，封裝 TracerProvider 初始化與優雅關閉
- TracerProvider 初始化時輸出完整設定 log
- 中英文 `config_guide.md` 加入 OpenTelemetry Tracing 設定說明

### Changed
- 新增全域變數 `var Trace TraceConf`，修正 `updateGlobalConfigs` 缺少的 Trace 同步
- 修正 `AppName`、`TrustedProxies` mapstructure tag 與 YAML 不一致
- Dockerfile 改用 `$BUILDPLATFORM` + Go cross-compile，解決 ARM Mac buildx QEMU segfault

### Fixed
- 修正 `serviceName` 為空（AppName tag 不匹配 YAML）
- 修正 OTEL resource schema URL 衝突（`resource.Merge` → `resource.New`）

---

## [v2.0.5] - 2026-04-14 (English)

### Added
- Integrated OpenTelemetry TracerProvider with OTLP over HTTP export to tracing backends (Tempo, Jaeger, etc.)
- Auto-created child spans for Telegram, Slack, and Discord message sending to track per-platform latency and errors
- Record errors via `span.RecordError` on send failure, filterable by `error=true` in Grafana
- Spans include `messaging.system`, `messaging.level`, `messaging.channel` attributes
- Added `pkg/trace/provider.go` module for TracerProvider initialization and graceful shutdown
- Log full trace configuration on TracerProvider initialization
- Added OpenTelemetry Tracing documentation to both EN and ZH `config_guide.md`

### Changed
- Added global `var Trace TraceConf` and fixed missing Trace sync in `updateGlobalConfigs`
- Fixed `AppName` and `TrustedProxies` mapstructure tags to match YAML keys
- Dockerfile uses `$BUILDPLATFORM` + Go cross-compile to avoid QEMU segfault on ARM Mac

### Fixed
- Fixed empty `serviceName` caused by mismatched AppName mapstructure tag
- Fixed OTEL resource schema URL conflict (`resource.Merge` → `resource.New`)
