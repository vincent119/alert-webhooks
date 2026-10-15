package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"alert-webhooks/config"

	"github.com/vincent119/zlogger"
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
		t.Fatalf("log level or format was not mapped correctly: %+v", resolved)
	}
	if got, want := resolved.Outputs, []string{"console", "file"}; len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("Outputs = %v, want %v", got, want)
	}
	if resolved.LogPath != logConfig.LogPath || resolved.FileName != logConfig.LogFile {
		t.Fatalf("file configuration was not mapped correctly: %+v", resolved)
	}
	if !resolved.AddCaller || !resolved.AddStacktrace || !resolved.Development {
		t.Fatalf("boolean configuration was not mapped correctly: %+v", resolved)
	}
}

func TestConfigureRejectsInvalidOutput(t *testing.T) {
	cleanup, err := zlogger.Configure(newConfigPatch(config.LogConf{
		LogFile: "../alerts.log",
		Outputs: "file",
	}, false))
	if err == nil {
		t.Fatal("Configure() should return a configuration error")
	}
	if cleanup != nil {
		t.Fatal("Configure() must not return cleanup on failure")
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
		t.Fatalf("compatible defaults were not used: %+v", resolved)
	}
	if got, want := resolved.Outputs, []string{"console"}; len(got) != len(want) || got[0] != want[0] {
		t.Fatalf("Outputs = %v, want %v", got, want)
	}
}

func TestConfigureWritesCategoryAndContextFields(t *testing.T) {
	logPath := t.TempDir()
	cleanup, err := zlogger.Configure(newConfigPatch(config.LogConf{
		Level:   "info",
		Format:  "json",
		LogPath: logPath,
		LogFile: "alerts.log",
		Outputs: "file",
	}, false))
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("cleanup() error = %v", err)
		}
	})

	zlogger.Info("test log", zlogger.String("category", "logger_test"), zlogger.String("request_id", "req-123"))
	ctx := zlogger.WithRequestID(context.Background(), "req-456")
	zlogger.InfoContext(ctx, "context log", zlogger.String("category", "default"), zlogger.String("operation", "notify"))

	if err := zlogger.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(logPath, "alerts.log"))
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Fatalf("log line count = %d, want 3", len(lines))
	}

	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[1]), &entry); err != nil {
		t.Fatalf("log is not valid JSON: %v", err)
	}
	if got := entry["category"]; got != "logger_test" {
		t.Fatalf("category = %v, want logger_test", got)
	}
	if got := entry["request_id"]; got != "req-123" {
		t.Fatalf("request_id = %v, want req-123", got)
	}

	if err := json.Unmarshal([]byte(lines[2]), &entry); err != nil {
		t.Fatalf("log is not valid JSON: %v", err)
	}
	if got := entry["category"]; got != "default" {
		t.Fatalf("category = %v, want default", got)
	}
	if got := entry["request_id"]; got != "req-456" {
		t.Fatalf("request_id = %v, want req-456", got)
	}
	if got := entry["operation"]; got != "notify" {
		t.Fatalf("operation = %v, want notify", got)
	}
}
