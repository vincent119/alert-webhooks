# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v2.0.6] - 2026-04-14

### Added
- Integrated OpenTelemetry TracerProvider with OTLP over HTTP export to tracing backends (Tempo, Jaeger, etc.)
- Added `tlsSkipVerify` field to TraceConf for HTTPS with self-signed certificate skip
- Added `WithFilter` to otelgin middleware to exclude `/`, `/healthz`, `/healthy` from tracing
- Auto-created child spans for Telegram, Slack, and Discord message sending to track per-platform latency and errors
- Record errors via `span.RecordError` on send failure, filterable by `error=true` in Grafana
- Spans include `messaging.system`, `messaging.level`, `messaging.channel` attributes
- Added `pkg/trace/provider.go` module for TracerProvider initialization and graceful shutdown
- Helm Secret includes `trace-auth-user` / `trace-auth-passwd`, Deployment injects `TRACE_AUTH_USER` / `TRACE_AUTH_PASSWD` env vars
- Helm ConfigMap includes `tlsSkipVerify` field
- Kustomize deployment adds `TRACE_AUTH_USER` / `TRACE_AUTH_PASSWD` env vars, aligned with Helm Chart
- Kustomize `config.production.yaml` adds complete `trace` configuration section
- README/README-zh adds Observability section (Prometheus Metrics, OpenTelemetry Tracing)
- `config.expamle.yaml` adds complete `trace` configuration section with all 9 fields
- `kubernetes-env-vars.md` (EN/ZH) adds `TRACE_AUTH_USER`, `TRACE_AUTH_PASSWD`, `DISCORD_TOKEN` env vars
- Added OpenTelemetry Tracing documentation to both EN and ZH `config_guide.md`

### Changed
- OTLP exporter switched from `WithEndpoint` + `WithInsecure()` to `WithEndpointURL` with explicit `http://` or `https://` scheme
- Added global `var Trace TraceConf` and fixed missing Trace sync in `updateGlobalConfigs`
- Fixed `AppName` and `TrustedProxies` mapstructure tags to match YAML keys
- Dockerfile uses `$BUILDPLATFORM` + Go cross-compile to avoid QEMU segfault on ARM Mac
- Route registration now checks service availability before registering endpoints

### Fixed
- Fixed OTLP HTTP exporter using HTTPS despite `insecure: true`, causing `tls: failed to verify certificate` error
- Fixed ELB health check (`GET /`) generating unnecessary trace spans
- Fixed empty `serviceName` caused by mismatched AppName mapstructure tag
- Fixed OTEL resource schema URL conflict (`resource.Merge` → `resource.New`)
- Fixed `insecure` field description in `config_guide.md` (EN/ZH) being opposite to actual code behavior
- Added missing `tlsSkipVerify` field to `config_guide.md` (EN/ZH)

## [v2.0.4] - 2025-10-31

### Fixed
- Fixed alert level routing issue

## [v2.0.3] - 2025-10-31

### Added
- Updated GitLab CI to support tag-triggered pipeline

## [v2.0.2] - 2025-10-31

### Fixed
- Fixed Telegram Level 4 (testing group) routing issue

## [v2.0.1] - 2025-10-31

### Fixed
- Fixed Telegram Level 4 (testing group) routing issue (same as v2.0.2)

## [v2.0] - 2025-10-31

### Fixed
- Fixed Telegram Level 4 (testing group) routing issue

## [v1.9] - 2025-09-17

### Fixed
- Fixed extra whitespace in message templates

## [v1.8] - 2025-09-16

### Added
- Added debug logging for Telegram server response messages

### Fixed
- Fixed template rendering issues

## [v1.7] - 2025-09-12

### Added
- Added debug display for received message data
- Added Kustomize deployment configuration
- Added Helm chart and Kustomization ConfigMap support

### Fixed
- Fixed template display issues and added received message data logging
- Fixed external link display issues
- Fixed namespace value configuration
- Updated container image repository references

## [v1.6] - 2025-09-12

### Added
- Added template descriptions
- Added Helm chart deployment support
- Merged feature branch `feature/alert-webhooks-fix`

### Fixed
- Fixed CORS configuration
- Fixed TOC (Table of Contents) recovery
- Fixed external link display issues

## [v1.5] - 2025-09-08

### Added
- Added project descriptions and documentation

## [v1.4] - 2025-09-03

### Fixed
- Fixed TOC (Table of Contents) recovery issues
- Configuration updates and improvements

## [v1.3] - 2025-09-02

### Added
- Added Helm chart deployment configuration
- Added `.gitignore` rules

### Fixed
- Fixed external link display issues

## [v1.2] - 2025-08-26

### Fixed
- Fixed CORS (Cross-Origin Resource Sharing) configuration

## [v1.1] - 2025-08-26

### Added
- Added ECR (Elastic Container Registry) repository configuration

### Fixed
- Fixed CORS configuration

## [v1.0] - 2025-08-25

### Added
- Initial release
- AlertManager webhook processing service
- Telegram notification support with multi-level chat group routing
- HTTP Basic Auth protection
- Multilingual template support (English, Traditional Chinese, Simplified Chinese, Japanese, Korean)
- Full and Minimal template display modes
- Separate firing and resolved alert notifications
- Swagger API documentation
- GitHub Actions CI/CD pipeline
- Docker and Docker Compose support
