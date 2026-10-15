package logger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"alert-webhooks/config"
)

func TestNewConfigPatchMapsLogConfiguration(t *testing.T) {
	logConfig := config.LogConf{
		Level:         "debug",
		Format:        "json",
		LogPath:       "/tmp/logs",
		LogFile:       "alerts.log",
		Outputs:       "console, file",
		AddCaller:     true,
		AddStacktrace: true,
	}

	resolved, err := newConfigPatch(logConfig, true).Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if resolved.Level != "debug" || resolved.Format != "json" {
		t.Fatalf("日誌等級或格式未正確映射：%+v", resolved)
	}
	if got, want := resolved.Outputs, []string{"console", "file"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("Outputs = %v，預期 %v", got, want)
	}
	if resolved.LogPath != logConfig.LogPath || resolved.FileName != logConfig.LogFile {
		t.Fatalf("檔案設定未正確映射：%+v", resolved)
	}
	if !resolved.AddCaller || !resolved.AddStacktrace || !resolved.Development {
		t.Fatalf("布林設定未正確映射：%+v", resolved)
	}
}

func TestConfigureRejectsInvalidOutput(t *testing.T) {
	cleanup, err := Configure(config.LogConf{
		LogFile: "../alerts.log",
		Outputs: "file",
	}, false)
	if err == nil {
		t.Fatal("Configure() 預期回傳設定錯誤")
	}
	if cleanup != nil {
		t.Fatal("Configure() 失敗時不應回傳 cleanup")
	}
}

func TestNewConfigPatchUsesDefaultsForUnsupportedLegacyValues(t *testing.T) {
	resolved, err := newConfigPatch(config.LogConf{
		Level:   "verbose",
		Format:  "console,json",
		Outputs: "unsupported",
	}, false).Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}
	if resolved.Level != "info" || resolved.Format != "console" {
		t.Fatalf("未使用相容預設值：%+v", resolved)
	}
	if got, want := resolved.Outputs, []string{"console"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("Outputs = %v，預期 %v", got, want)
	}
}

func TestConfigureWritesCategoryFieldAndCleansUp(t *testing.T) {
	logPath := t.TempDir()
	cleanup, err := Configure(config.LogConf{
		Level:   "info",
		Format:  "json",
		LogPath: logPath,
		LogFile: "alerts.log",
		Outputs: "file",
	}, false)
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("cleanup() error = %v", err)
		}
	})

	Info("測試日誌", "logger_test", String("request_id", "req-123"))
	InfoContext(WithRequestID(context.Background(), "req-456"), "context 日誌", String("operation", "notify"))

	if err := Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(logPath, "alerts.log"))
	if err != nil {
		t.Fatalf("讀取日誌檔失敗：%v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Fatalf("日誌行數 = %d，預期 3", len(lines))
	}
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &entry); err != nil {
		t.Fatalf("日誌不是有效 JSON：%v", err)
	}
	if got := entry["category"]; got != "logger_test" {
		t.Fatalf("category = %v，預期 logger_test", got)
	}
	if got := entry["request_id"]; got != "req-123" {
		t.Fatalf("request_id = %v，預期 req-123", got)
	}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("日誌不是有效 JSON：%v", err)
	}
	if got := entry["category"]; got != "default" {
		t.Fatalf("category = %v，預期 default", got)
	}
	if got := entry["request_id"]; got != "req-456" {
		t.Fatalf("request_id = %v，預期 req-456", got)
	}
	if got := entry["operation"]; got != "notify" {
		t.Fatalf("operation = %v，預期 notify", got)
	}
}
