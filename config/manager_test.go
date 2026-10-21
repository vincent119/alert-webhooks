package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入測試設定失敗：%v", err)
	}
	return path
}

func clearConfigTokenEnvironment(t *testing.T) {
	t.Helper()

	for _, name := range []string{
		"METRIC_USER",
		"METRIC_PASSWORD",
		"WEBHOOKS_USER",
		"WEBHOOKS_PASSWORD",
		"WEBHOOKS_CORS_ALLOWED_ORIGINS",
		"TRACE_AUTH_USER",
		"TRACE_AUTH_PASSWD",
		"TELEGRAM_TOKEN",
		"SLACK_TOKEN",
		"DISCORD_TOKEN",
	} {
		t.Setenv(name, "")
	}
}

func TestConfigManagerLoadConfigPublishesTypedSnapshot(t *testing.T) {
	clearConfigTokenEnvironment(t)
	path := writeConfigFile(t, t.TempDir(), "config.yaml", `
app:
  mode: test
slack:
  enable: true
  token: file-token
  channel: '#alerts'
  channels:
    l0: '#critical'
`)
	manager := NewConfigManager()

	if err := manager.LoadConfig(path); err != nil {
		t.Fatalf("載入設定失敗：%v", err)
	}
	got := Current()
	if got.App.Mode != "test" || !got.Slack.Enable || got.Slack.Token != "file-token" {
		t.Fatalf("typed snapshot 不符預期：%+v", got)
	}
	if got.Slack.Channels["l0"] != "#critical" {
		t.Fatalf("Slack channels 不符預期：%v", got.Slack.Channels)
	}
	if manager.ConfigFileUsed() != path {
		t.Fatalf("ConfigFileUsed 不符預期：%q", manager.ConfigFileUsed())
	}
	if manager.GetViper() == nil {
		t.Fatal("GetViper 不應回傳 nil")
	}
}

func TestConfigManagerLoadConfigRejectsInvalidFiles(t *testing.T) {
	clearConfigTokenEnvironment(t)
	tests := []struct {
		name    string
		path    func(*testing.T) string
		wantErr string
	}{
		{
			name: "missing file",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing.yaml")
			},
			wantErr: "讀取設定檔失敗",
		},
		{
			name: "invalid yaml",
			path: func(t *testing.T) string {
				return writeConfigFile(t, t.TempDir(), "invalid.yaml", "app: [invalid")
			},
			wantErr: "讀取設定檔失敗",
		},
		{
			name: "invalid typed value",
			path: func(t *testing.T) string {
				return writeConfigFile(t, t.TempDir(), "invalid-type.yaml", "app: invalid\n")
			},
			wantErr: "解析設定失敗",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewConfigManager()
			err := manager.LoadConfig(tt.path(t))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
		})
	}
}

func TestConfigManagerAccessorsAndSettings(t *testing.T) {
	manager := NewConfigManager()
	manager.Set("service.name", "alert-webhooks")
	manager.Set("service.retries", 3)
	manager.Set("service.enabled", true)
	manager.Set("service.timeout", "2s")

	if got := manager.GetString("service.name"); got != "alert-webhooks" {
		t.Fatalf("GetString 不符預期：%q", got)
	}
	if got := manager.GetInt("service.retries"); got != 3 {
		t.Fatalf("GetInt 不符預期：%d", got)
	}
	if got := manager.GetBool("service.enabled"); !got {
		t.Fatal("GetBool 不符預期")
	}
	if got := manager.GetDuration("service.timeout"); got != 2*time.Second {
		t.Fatalf("GetDuration 不符預期：%s", got)
	}
	if !manager.IsSet("service.name") || manager.IsSet("service.missing") {
		t.Fatal("IsSet 結果不符預期")
	}
	settings := manager.AllSettings()
	if service, ok := settings["service"].(map[string]interface{}); !ok || service["name"] != "alert-webhooks" {
		t.Fatalf("AllSettings 不符預期：%v", settings)
	}
}

func TestConfigManagerWriteMethods(t *testing.T) {
	clearConfigTokenEnvironment(t)
	dir := t.TempDir()
	source := writeConfigFile(t, dir, "source.yaml", "app:\n  mode: test\n")
	manager := NewConfigManager()
	if err := manager.LoadConfig(source); err != nil {
		t.Fatalf("載入設定失敗：%v", err)
	}

	manager.Set("app.mode", "production")
	if err := manager.WriteConfig(); err != nil {
		t.Fatalf("WriteConfig 失敗：%v", err)
	}
	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("讀取寫入結果失敗：%v", err)
	}
	if !strings.Contains(string(content), "production") {
		t.Fatalf("WriteConfig 未寫入更新值：%s", content)
	}

	copyPath := filepath.Join(dir, "copy.yaml")
	if err := manager.WriteConfigAs(copyPath); err != nil {
		t.Fatalf("WriteConfigAs 失敗：%v", err)
	}
	if _, err := os.Stat(copyPath); err != nil {
		t.Fatalf("WriteConfigAs 未建立檔案：%v", err)
	}

	safePath := filepath.Join(dir, "safe.yaml")
	if err := manager.SafeWriteConfigAs(safePath); err != nil {
		t.Fatalf("SafeWriteConfigAs 第一次寫入失敗：%v", err)
	}
	if err := manager.SafeWriteConfigAs(safePath); err == nil {
		t.Fatal("SafeWriteConfigAs 不應覆寫既有檔案")
	}
	if err := manager.SafeWriteConfig(); err == nil {
		t.Fatal("SafeWriteConfig 不應覆寫目前既有設定檔")
	}
}

func TestConfigManagerWatchAndCloseWithoutConfiguration(t *testing.T) {
	manager := NewConfigManager()
	if err := manager.WatchConfig(); err == nil || !strings.Contains(err.Error(), "未設定設定檔") {
		t.Fatalf("缺少設定檔的 watcher 錯誤不符預期：%v", err)
	}
	called := false
	manager.OnConfigChange(func() { called = true })
	if len(manager.callbacks) != 1 || called {
		t.Fatal("OnConfigChange 應註冊但不立即執行 callback")
	}
	if err := manager.Close(); err != nil {
		t.Fatalf("沒有 watcher 時 Close 不應失敗：%v", err)
	}
}

func TestConfigManagerWatchReloadsChangedFileAndInvokesCallback(t *testing.T) {
	clearConfigTokenEnvironment(t)
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.yaml", "app:\n  version: before\n")
	manager := NewConfigManager()
	if err := manager.LoadConfig(path); err != nil {
		t.Fatalf("載入初始設定失敗：%v", err)
	}

	callback := make(chan struct{}, 1)
	manager.OnConfigChange(func() {
		select {
		case callback <- struct{}{}:
		default:
		}
	})
	if err := manager.WatchConfig(); err != nil {
		t.Fatalf("啟動設定監聽失敗：%v", err)
	}
	t.Cleanup(func() {
		if err := manager.Close(); err != nil {
			t.Errorf("關閉設定監聽失敗：%v", err)
		}
	})

	if err := os.WriteFile(path, []byte("app:\n  version: after\n"), 0o600); err != nil {
		t.Fatalf("更新設定檔失敗：%v", err)
	}

	select {
	case <-callback:
	case <-time.After(3 * time.Second):
		t.Fatal("等待設定變更 callback 逾時")
	}
	if got := Current().App.Version; got != "after" {
		t.Fatalf("重新載入版本 = %q，預期 %q", got, "after")
	}
}
