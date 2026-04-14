# Changelog

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
