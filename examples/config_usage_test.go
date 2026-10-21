package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"alert-webhooks/config"
)

func TestValidateConfigReportsInvalidAndValidConfiguration(t *testing.T) {
	originalApp := config.App
	originalLog := config.Log
	t.Cleanup(func() {
		config.App = originalApp
		config.Log = originalLog
	})

	config.App = config.AppConf{}
	config.Log = config.LogConf{}
	invalidOutput := captureExampleOutput(t, validateConfig)
	for _, expected := range []string{
		"應用端口未設定",
		"應用密鑰未設定",
		"日誌級別未設定",
		"日誌文件大小設定不合理",
		"日誌保留天數設定不合理",
		"配置驗證完成",
	} {
		if !strings.Contains(invalidOutput, expected) {
			t.Fatalf("invalid output 未包含 %q：%s", expected, invalidOutput)
		}
	}

	config.App = config.AppConf{Port: "9999", Key: "test-key"}
	config.Log = config.LogConf{Level: "info", MaxSize: 10, MaxAge: 7}
	validOutput := captureExampleOutput(t, validateConfig)
	if got, want := strings.TrimSpace(validOutput), "配置驗證完成"; got != want {
		t.Fatalf("valid output = %q，預期 %q", got, want)
	}
}

func TestExportConfigWritesCurrentSettings(t *testing.T) {
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("取得工作目錄失敗：%v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("切換 temporary directory 失敗：%v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Errorf("還原工作目錄失敗：%v", err)
		}
	})

	manager := config.NewConfigManager()
	manager.Set("app.mode", "test")
	output := captureExampleOutput(t, func() { exportConfig(manager) })
	if !strings.Contains(output, "當前所有配置:") || !strings.Contains(output, "配置已導出到:") {
		t.Fatalf("export output 不完整：%s", output)
	}

	matches, err := filepath.Glob(filepath.Join(workingDir, "config_export_*.yaml"))
	if err != nil {
		t.Fatalf("搜尋匯出設定失敗：%v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("匯出檔案數量 = %d，預期 1", len(matches))
	}
}

func captureExampleOutput(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("建立 stdout pipe 失敗：%v", err)
	}
	previousStdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = previousStdout })

	run()
	if err := writer.Close(); err != nil {
		t.Fatalf("關閉 stdout writer 失敗：%v", err)
	}
	os.Stdout = previousStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("讀取 stdout 失敗：%v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("關閉 stdout reader 失敗：%v", err)
	}
	return string(output)
}
