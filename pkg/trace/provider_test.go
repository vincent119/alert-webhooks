package trace

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"alert-webhooks/config"
)

func TestApplyDefaultsAndBasicAuth(t *testing.T) {
	cfg := config.TraceConf{}
	applyDefaults(&cfg)
	if cfg.SampleRate != 1.0 {
		t.Fatalf("default sample rate = %v，預期 1.0", cfg.SampleRate)
	}

	cfg.SampleRate = 0.25
	applyDefaults(&cfg)
	if cfg.SampleRate != 0.25 {
		t.Fatalf("explicit sample rate = %v，預期 0.25", cfg.SampleRate)
	}

	want := base64.StdEncoding.EncodeToString([]byte("user:password"))
	if got := basicAuth("user", "password"); got != want {
		t.Fatalf("basic auth = %q，預期 %q", got, want)
	}
}

func TestInitTracerProviderDisabledReturnsNoopShutdown(t *testing.T) {
	shutdown, err := InitTracerProvider(context.Background(), config.TraceConf{}, "disabled-test")
	if err != nil {
		t.Fatalf("初始化 disabled tracer 失敗：%v", err)
	}
	if err := shutdown(context.Background()); err != nil {
		t.Fatalf("disabled shutdown 失敗：%v", err)
	}
}

func TestInitTracerProviderEnabledConfigurations(t *testing.T) {
	testCases := []struct {
		name string
		cfg  config.TraceConf
	}{
		{
			name: "insecure default path never sample",
			cfg: config.TraceConf{
				Enable: true, Url: "127.0.0.1", Port: "4318", Insecure: true, SampleRate: -1,
			},
		},
		{
			name: "tls custom path always sample with auth",
			cfg: config.TraceConf{
				Enable: true, Url: "collector.example.test", UrlPath: "/custom/traces",
				TlsSkipVerify: true, AuthUser: "user", AuthPasswd: "password", SampleRate: 1,
			},
		},
		{
			name: "ratio sampler",
			cfg: config.TraceConf{
				Enable: true, Url: "127.0.0.1", Insecure: true, SampleRate: 0.25,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			shutdown, err := InitTracerProvider(context.Background(), testCase.cfg, "trace-test")
			if err != nil {
				t.Fatalf("初始化 tracer 失敗：%v", err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				t.Fatalf("關閉 tracer 失敗：%v", err)
			}
		})
	}
}
