// Package trace 提供 OpenTelemetry TracerProvider 初始化功能
package trace

import (
	"alert-webhooks/config"
	"alert-webhooks/pkg/logger"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

var traceCategory = "trace"

// InitTracerProvider 根據 TraceConf 初始化 OTEL TracerProvider
// 回傳 shutdown function 供優雅關閉時呼叫
func InitTracerProvider(ctx context.Context, cfg config.TraceConf, serviceName string) (func(context.Context) error, error) {
	// 套用預設值：enable=false, insecure=true(HTTP), sampleRate=1.0, authUser/authPasswd 空值不啟用認證
	applyDefaults(&cfg)

	logger.Info("TracerProvider 初始化中", traceCategory,
		logger.Bool("enable", cfg.Enable),
		logger.String("url", cfg.Url),
		logger.String("port", cfg.Port),
		logger.Bool("insecure", cfg.Insecure),
		logger.Bool("tlsSkipVerify", cfg.TlsSkipVerify),
		logger.String("urlPath", cfg.UrlPath),
		logger.Float64("sampleRate", cfg.SampleRate),
	)

	if !cfg.Enable {
		logger.Info("OpenTelemetry tracing 已停用", traceCategory)
		// 未啟用時回傳 noop shutdown
		return func(ctx context.Context) error { return nil }, nil
	}

	// 組合 endpoint
	endpoint := cfg.Url
	if cfg.Port != "" {
		endpoint = fmt.Sprintf("%s:%s", cfg.Url, cfg.Port)
	}

	// 決定 scheme：insecure=true 走 http，否則走 https
	scheme := "https"
	if cfg.Insecure {
		scheme = "http"
	}

	// 組合完整 URL path
	urlPath := cfg.UrlPath
	if urlPath == "" {
		urlPath = "/v1/traces"
	}

	// 使用 WithEndpointURL 明確帶上 scheme，避免 WithInsecure() 不生效的問題
	endpointURL := fmt.Sprintf("%s://%s%s", scheme, endpoint, urlPath)

	logger.Info("OTLP exporter endpoint URL", traceCategory,
		logger.String("endpointURL", endpointURL),
	)

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(endpointURL),
	}

	// TLS 設定：insecure=false 且 tlsSkipVerify=true 時跳過憑證驗證（自簽憑證場景）
	if !cfg.Insecure && cfg.TlsSkipVerify {
		tlsCfg := &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec // 使用者明確設定跳過憑證驗證（自簽憑證場景）
		}
		opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsCfg))
	}

	// Basic auth
	if cfg.AuthUser != "" && cfg.AuthPasswd != "" {
		headerValue := "Basic " + basicAuth(cfg.AuthUser, cfg.AuthPasswd)
		opts = append(opts, otlptracehttp.WithHeaders(map[string]string{
			"Authorization": headerValue,
		}))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("建立 OTLP exporter 失敗: %w", err)
	}

	// 建立 resource，直接用 resource.New 避免 schema URL 衝突
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("建立 OTEL resource 失敗: %w", err)
	}

	// 設定 sampler
	var sampler sdktrace.Sampler
	switch {
	case cfg.SampleRate <= 0:
		sampler = sdktrace.NeverSample()
	case cfg.SampleRate >= 1.0:
		sampler = sdktrace.AlwaysSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(cfg.SampleRate)
	}

	// 建立 TracerProvider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sampler)),
	)

	// 註冊全域 TracerProvider 和 Propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("TracerProvider 初始化完成", traceCategory,
		logger.String("endpointURL", endpointURL),
		logger.String("serviceName", serviceName),
		logger.Float64("sampleRate", cfg.SampleRate),
	)

	return tp.Shutdown, nil
}

// applyDefaults 套用 TraceConf 預設值
// enable 預設 false（零值即可）、insecure 預設 true（HTTP）、sampleRate 預設 1.0
func applyDefaults(cfg *config.TraceConf) {
	// sampleRate 零值時套用預設 1.0（完整取樣）
	if cfg.SampleRate == 0 {
		cfg.SampleRate = 1.0
	}
}

// basicAuth 產生 Base64 編碼的 basic auth 字串
func basicAuth(user, password string) string {
	credentials := user + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(credentials))
}
