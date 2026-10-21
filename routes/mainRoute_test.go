package routes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"alert-webhooks/config"
	"alert-webhooks/docs"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

func TestSetupSwaggerHostUsesEnvironmentAndModeDefaults(t *testing.T) {
	loadRouteConfig(t, "development", "9999", "", "", "")
	t.Setenv("SWAGGER_HOST", "docs.example.test")
	setupSwaggerHost()
	if got, want := docs.SwaggerInfo.Host, "docs.example.test"; got != want {
		t.Fatalf("Swagger host = %q，預期 %q", got, want)
	}

	t.Setenv("SWAGGER_HOST", "")
	setupSwaggerHost()
	if got, want := docs.SwaggerInfo.Host, "localhost:9999"; got != want {
		t.Fatalf("development Swagger host = %q，預期 %q", got, want)
	}

	loadRouteConfig(t, "production", "9999", "", "", "")
	setupSwaggerHost()
	if docs.SwaggerInfo.Host != "" {
		t.Fatalf("production Swagger host = %q，預期自動判斷", docs.SwaggerInfo.Host)
	}
}

func TestSetupGinModeMapsConfiguredModes(t *testing.T) {
	testCases := []struct {
		name string
		mode string
		want string
	}{
		{name: "production", mode: "production", want: gin.ReleaseMode},
		{name: "release", mode: "release", want: gin.ReleaseMode},
		{name: "development", mode: "development", want: gin.DebugMode},
		{name: "debug", mode: "debug", want: gin.DebugMode},
		{name: "test", mode: "test", want: gin.TestMode},
		{name: "unknown", mode: "unknown", want: gin.ReleaseMode},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			loadRouteConfig(t, testCase.mode, "9999", "", "", "")
			setupGinMode()
			if got := gin.Mode(); got != testCase.want {
				t.Fatalf("Gin mode = %q，預期 %q", got, testCase.want)
			}
		})
	}
}

func TestSetupTrustedProxiesAcceptsConfiguredAndDefaultRanges(t *testing.T) {
	testCases := []struct {
		name    string
		mode    string
		proxies string
	}{
		{name: "configured", mode: "test", proxies: "127.0.0.1,10.0.0.0/8"},
		{name: "production defaults", mode: "production"},
		{name: "non-production defaults", mode: "development"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			loadRouteConfig(t, testCase.mode, "9999", testCase.proxies, "", "")
			if err := setupTrustedProxies(gin.New()); err != nil {
				t.Fatalf("設定 trusted proxies 失敗：%v", err)
			}
		})
	}
}

func TestSetupTrustedProxiesRejectsInvalidRange(t *testing.T) {
	loadRouteConfig(t, "test", "9999", "not-a-proxy", "", "")
	if err := setupTrustedProxies(gin.New()); err == nil {
		t.Fatal("無效 trusted proxy 應回傳錯誤")
	}
}

func TestSetupPrometheusRegistersAuthenticatedAndOpenEndpoints(t *testing.T) {
	for _, testCase := range []struct {
		name     string
		user     string
		password string
		wantCode int
	}{
		{name: "open", wantCode: http.StatusOK},
		{name: "authenticated", user: "metrics", password: "secret", wantCode: http.StatusUnauthorized},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			resetPrometheusRegistry(t)
			loadRouteConfig(t, "test", "9999", "", testCase.user, testCase.password)
			router := gin.New()
			setupPrometheus(router)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/metrics", nil))
			if response.Code != testCase.wantCode {
				t.Fatalf("metrics status = %d，預期 %d", response.Code, testCase.wantCode)
			}
		})
	}
}

func TestDefaultRouteRegistersHealthEndpoints(t *testing.T) {
	resetPrometheusRegistry(t)
	loadRouteConfig(t, "test", "9999", "127.0.0.1", "", "")
	t.Setenv("SWAGGER_HOST", "")
	router := DefaultRoute()

	for _, testCase := range []struct {
		path string
		body string
	}{
		{path: "/healthy", body: `{"status":"ok","version":"v-test"}`},
		{path: "/api/v1/healthz", body: `{"status":"OK"}`},
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, testCase.path, nil))
		if response.Code != http.StatusOK {
			t.Fatalf("%s status = %d，預期 %d", testCase.path, response.Code, http.StatusOK)
		}
		if got := response.Body.String(); got != testCase.body {
			t.Fatalf("%s body = %q，預期 %q", testCase.path, got, testCase.body)
		}
	}
}

func loadRouteConfig(t *testing.T, mode, port, proxies, metricUser, metricPassword string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	content := []byte("app:\n" +
		"  version: v-test\n" +
		"  app_name: alert-webhooks-test\n" +
		"  mode: " + mode + "\n" +
		"  port: \"" + port + "\"\n" +
		"  trusted_proxies: \"" + proxies + "\"\n" +
		"metric:\n" +
		"  user: \"" + metricUser + "\"\n" +
		"  password: \"" + metricPassword + "\"\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("建立 route 測試設定失敗：%v", err)
	}
	if err := config.LoadConfigFromFile(path); err != nil {
		t.Fatalf("載入 route 測試設定失敗：%v", err)
	}
}

func resetPrometheusRegistry(t *testing.T) {
	t.Helper()
	registry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = registry
	prometheus.DefaultGatherer = registry
}
