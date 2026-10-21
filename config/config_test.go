package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/viper"
)

func TestCurrentReturnsIndependentSnapshot(t *testing.T) {
	publish(Config{
		App:   AppConf{Mode: "development"},
		Slack: SlackConf{Channels: map[string]string{"critical": "C1"}},
		Discord: DiscordConf{
			Channels:     map[string]string{"critical": "D1"},
			MentionRoles: []string{"role-1"},
		},
	}, false)

	first := Current()
	first.Slack.Channels["critical"] = "modified"
	first.Discord.Channels["critical"] = "modified"
	first.Discord.MentionRoles[0] = "modified"

	current := Current()
	if current.Slack.Channels["critical"] != "C1" {
		t.Fatalf("Slack channel = %q, want %q", current.Slack.Channels["critical"], "C1")
	}
	if current.Discord.Channels["critical"] != "D1" {
		t.Fatalf("Discord channel = %q, want %q", current.Discord.Channels["critical"], "D1")
	}
	if current.Discord.MentionRoles[0] != "role-1" {
		t.Fatalf("Discord mention role = %q, want %q", current.Discord.MentionRoles[0], "role-1")
	}
}

func TestLoadConfigFromFileKeepsCurrentSnapshotWhenParsingFails(t *testing.T) {
	publish(Config{App: AppConf{Version: "stable"}}, false)

	configPath := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(configPath, []byte("app: ["), 0o600); err != nil {
		t.Fatalf("write invalid config: %v", err)
	}

	if err := LoadConfigFromFile(configPath); err == nil {
		t.Fatal("LoadConfigFromFile() error = nil, want parse error")
	}
	if got := Current().App.Version; got != "stable" {
		t.Fatalf("current version = %q, want %q", got, "stable")
	}
}

func TestCurrentIsSafeDuringConcurrentPublish(t *testing.T) {
	const iterations = 1000
	var writers sync.WaitGroup
	writers.Add(1)
	go func() {
		defer writers.Done()
		for i := 0; i < iterations; i++ {
			publish(Config{App: AppConf{Version: fmt.Sprintf("v%d", i)}}, false)
		}
	}()

	var readers sync.WaitGroup
	for i := 0; i < 4; i++ {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for j := 0; j < iterations; j++ {
				_ = Current().App.Version
			}
		}()
	}

	writers.Wait()
	readers.Wait()
}

func TestOverrideWithEnvVarsCORSAllowedOrigins(t *testing.T) {
	t.Setenv("WEBHOOKS_CORS_ALLOWED_ORIGINS", " https://console.example.com, ,*,https://admin.example.com ")
	candidate := Config{
		Webhooks: WebhooksConf{
			CORSAllowedOrigins: []string{"https://from-config.example.com"},
		},
	}

	overrideWithEnvVars(&candidate)

	want := []string{"https://console.example.com", "https://admin.example.com"}
	if !reflect.DeepEqual(candidate.Webhooks.CORSAllowedOrigins, want) {
		t.Fatalf("CORS origins = %v, want %v", candidate.Webhooks.CORSAllowedOrigins, want)
	}
}

func TestCurrentCopiesCORSAllowedOrigins(t *testing.T) {
	original := Current()
	t.Cleanup(func() { publish(original, false) })
	publish(Config{Webhooks: WebhooksConf{CORSAllowedOrigins: []string{"https://console.example.com"}}}, false)

	current := Current()
	current.Webhooks.CORSAllowedOrigins[0] = "https://changed.example.com"
	if got := Current().Webhooks.CORSAllowedOrigins[0]; got != "https://console.example.com" {
		t.Fatalf("CORS origin returned by Current can modify the published snapshot: %q", got)
	}
}

func TestOverrideWithEnvVarsAppliesSupportedValues(t *testing.T) {
	tests := map[string]string{
		"METRIC_USER":                   "metrics-user",
		"METRIC_PASSWORD":               "metrics-password",
		"WEBHOOKS_USER":                 "webhooks-user",
		"WEBHOOKS_PASSWORD":             "webhooks-password",
		"WEBHOOKS_CORS_ALLOWED_ORIGINS": "https://console.example.com",
		"TRACE_AUTH_USER":               "trace-user",
		"TRACE_AUTH_PASSWD":             "trace-password",
		"TELEGRAM_TOKEN":                "telegram-token",
		"SLACK_TOKEN":                   "slack-token",
		"DISCORD_TOKEN":                 "discord-token",
	}
	for name, value := range tests {
		t.Setenv(name, value)
	}

	candidate := Config{}
	overrideWithEnvVars(&candidate)

	if candidate.Metric.User != "metrics-user" || candidate.Metric.Password != "metrics-password" {
		t.Fatalf("Metric 環境變數未套用：%+v", candidate.Metric)
	}
	if candidate.Webhooks.BaseAuthUser != "webhooks-user" || candidate.Webhooks.BaseAuthPassword != "webhooks-password" {
		t.Fatalf("Webhooks 環境變數未套用：%+v", candidate.Webhooks)
	}
	if !reflect.DeepEqual(candidate.Webhooks.CORSAllowedOrigins, []string{"https://console.example.com"}) {
		t.Fatalf("CORS 環境變數未套用：%v", candidate.Webhooks.CORSAllowedOrigins)
	}
	if candidate.Trace.AuthUser != "trace-user" || candidate.Trace.AuthPasswd != "trace-password" {
		t.Fatalf("Trace 環境變數未套用：%+v", candidate.Trace)
	}
	if candidate.Telegram.Token != "telegram-token" || candidate.Slack.Token != "slack-token" || candidate.Discord.Token != "discord-token" {
		t.Fatalf("Provider token 環境變數未套用：telegram=%q slack=%q discord=%q", candidate.Telegram.Token, candidate.Slack.Token, candidate.Discord.Token)
	}
}

func TestPublishUpdatesLegacyGlobals(t *testing.T) {
	original := Current()
	t.Cleanup(func() { publish(original, true) })

	candidate := Config{
		App:      AppConf{Mode: "test"},
		Metric:   MetricConf{User: "metrics-user"},
		Trace:    TraceConf{Enable: true},
		Log:      LogConf{Level: "debug"},
		Telegram: TelegramConf{Enable: true},
		Webhooks: WebhooksConf{Enable: true},
		Slack:    SlackConf{Enable: true},
		Discord:  DiscordConf{Enable: true},
	}
	publish(candidate, true)

	if App.Mode != "test" || Metric.User != "metrics-user" || !Trace.Enable || Log.Level != "debug" {
		t.Fatalf("legacy globals 未更新：App=%+v Metric=%+v Trace=%+v Log=%+v", App, Metric, Trace, Log)
	}
	if !Telegram.Enable || !Webhooks.Enable || !Slack.Enable || !Conf.Discord.Enable {
		t.Fatalf("provider legacy globals 未更新：Telegram=%+v Webhooks=%+v Slack=%+v Discord=%+v", Telegram, Webhooks, Slack, Conf.Discord)
	}
	if Conf.App.Mode != "test" || Conf.Metric.User != "metrics-user" || !Conf.Trace.Enable || Conf.Log.Level != "debug" {
		t.Fatalf("Conf snapshot 未更新：%+v", Conf)
	}
}

func TestCurrentFallsBackToLegacyGlobalsBeforePublish(t *testing.T) {
	originalPointer := configStore.Load()
	originalApp := App
	originalDiscord := Conf.Discord
	t.Cleanup(func() {
		configStore.Store(originalPointer)
		App = originalApp
		Conf.Discord = originalDiscord
	})

	configStore.Store(nil)
	App = AppConf{Version: "legacy"}
	Conf.Discord = DiscordConf{Channels: map[string]string{"l0": "legacy-channel"}}

	got := Current()
	if got.App.Version != "legacy" || got.Discord.Channels["l0"] != "legacy-channel" {
		t.Fatalf("legacy fallback 不符預期：%+v", got)
	}
}

func TestEnvironmentHelpers(t *testing.T) {
	originalFlag := *env
	t.Cleanup(func() { *env = originalFlag })
	t.Setenv("APP_ENV", "")
	t.Setenv("GO_ENV", "")

	*env = "flag-env"
	if got := getEnvironment(); got != "flag-env" {
		t.Fatalf("flag 環境不符預期：%q", got)
	}

	*env = ""
	if err := os.Setenv("APP_ENV", "app-env"); err != nil {
		t.Fatalf("設定 APP_ENV 失敗：%v", err)
	}
	if got := GetEnvironment(); got != "app-env" {
		t.Fatalf("APP_ENV 不符預期：%q", got)
	}

	if err := os.Setenv("APP_ENV", ""); err != nil {
		t.Fatalf("清除 APP_ENV 失敗：%v", err)
	}
	if err := os.Setenv("GO_ENV", "go-env"); err != nil {
		t.Fatalf("設定 GO_ENV 失敗：%v", err)
	}
	if got := getEnvironment(); got != "go-env" {
		t.Fatalf("GO_ENV 不符預期：%q", got)
	}

	if err := os.Setenv("GO_ENV", ""); err != nil {
		t.Fatalf("清除 GO_ENV 失敗：%v", err)
	}
	if got := getEnvironment(); got != "development" {
		t.Fatalf("預設環境不符預期：%q", got)
	}

	for mode, checks := range map[string][3]bool{
		"development": {true, false, false},
		"production":  {false, true, false},
		"test":        {false, false, true},
	} {
		publish(Config{App: AppConf{Mode: mode}}, false)
		got := [3]bool{IsDevelopment(), IsProduction(), IsTest()}
		if got != checks {
			t.Fatalf("mode %q 判斷不符預期：got=%v want=%v", mode, got, checks)
		}
		if GetFullConfig().App.Mode != mode {
			t.Fatalf("GetFullConfig mode 不符預期：%q", GetFullConfig().App.Mode)
		}
	}
}

func TestLoadConfigFromFileSuccessAndErrors(t *testing.T) {
	clearConfigTokenEnvironment(t)
	publish(Config{App: AppConf{Version: "stable"}}, false)
	dir := t.TempDir()
	valid := writeConfigFile(t, dir, "valid.yaml", "app:\n  version: v2\n  mode: test\n")
	if err := LoadConfigFromFile(valid); err != nil {
		t.Fatalf("載入有效設定失敗：%v", err)
	}
	if got := Current().App; got.Version != "v2" || got.Mode != "test" {
		t.Fatalf("載入設定不符預期：%+v", got)
	}

	missing := filepath.Join(dir, "missing.yaml")
	if err := LoadConfigFromFile(missing); err == nil || !strings.Contains(err.Error(), "讀取設定檔") {
		t.Fatalf("不存在設定錯誤不符預期：%v", err)
	}
	invalidType := writeConfigFile(t, dir, "invalid-type.yaml", "app: invalid\n")
	if err := LoadConfigFromFile(invalidType); err == nil || !strings.Contains(err.Error(), "解析設定檔") {
		t.Fatalf("typed parse 錯誤不符預期：%v", err)
	}
}

func TestMergeConfigUpdatesOnlySupportedAppFields(t *testing.T) {
	publish(Config{
		App:   AppConf{Version: "v1", Mode: "development", Port: "8080"},
		Slack: SlackConf{Channel: "#alerts"},
	}, false)

	if err := MergeConfig(map[string]interface{}{
		"app.version": "v2",
		"app.mode":    "production",
		"app.port":    "9090",
	}); err != nil {
		t.Fatalf("合併設定失敗：%v", err)
	}
	got := Current()
	if got.App.Version != "v2" || got.App.Mode != "production" || got.App.Port != "9090" {
		t.Fatalf("App merge 不符預期：%+v", got.App)
	}
	if got.Slack.Channel != "#alerts" {
		t.Fatalf("未指定設定不應被清除：%+v", got.Slack)
	}

	if err := MergeConfig(map[string]interface{}{"app": "invalid"}); err == nil || !strings.Contains(err.Error(), "解析部分設定失敗") {
		t.Fatalf("無效 partial config 錯誤不符預期：%v", err)
	}
}

func TestGetConfigPathsReturnsConfiguredSearchPaths(t *testing.T) {
	originalConfigFile := *configFile
	originalEnv := *env
	originalPaths := append([]string(nil), configPaths...)
	t.Cleanup(func() {
		*configFile = originalConfigFile
		*env = originalEnv
		configPaths = originalPaths
	})

	*configFile = ""
	*env = "test"
	setupConfigPaths(viper.New())
	want := []string{".", "./config", "./configs", "./configs/test"}
	if !reflect.DeepEqual(GetConfigPaths(), want) {
		t.Fatalf("設定搜尋路徑不符預期：got=%v want=%v", GetConfigPaths(), want)
	}
}

func TestSetupConfigPathsUsesExplicitConfigFile(t *testing.T) {
	originalConfigFile := *configFile
	t.Cleanup(func() { *configFile = originalConfigFile })

	path := writeConfigFile(t, t.TempDir(), "explicit.yaml", "app:\n  mode: test\n")
	*configFile = path
	v := viper.New()
	setupConfigPaths(v)
	if err := v.ReadInConfig(); err != nil {
		t.Fatalf("讀取明確指定設定檔失敗：%v", err)
	}
	if got := v.ConfigFileUsed(); got != path {
		t.Fatalf("ConfigFileUsed = %q，預期 %q", got, path)
	}
}
