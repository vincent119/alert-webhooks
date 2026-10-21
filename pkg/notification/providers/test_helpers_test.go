package providers

import (
	"os"
	"path/filepath"
	"testing"

	"alert-webhooks/config"
)

type fakeTemplateEngine struct {
	languages []string
}

func (f *fakeTemplateEngine) GetSupportedLanguages() []string {
	return f.languages
}

func loadProviderConfig(t *testing.T, content string) {
	t.Helper()

	for _, name := range []string{"TELEGRAM_TOKEN", "SLACK_TOKEN", "DISCORD_TOKEN"} {
		t.Setenv(name, "")
	}

	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入測試設定失敗：%v", err)
	}
	if err := config.LoadConfigFromFile(path); err != nil {
		t.Fatalf("載入測試設定失敗：%v", err)
	}
}
