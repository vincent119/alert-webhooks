package template

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	texttemplate "text/template"
)

func changeWorkingDirectory(t *testing.T, dir string) {
	t.Helper()

	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("讀取目前目錄失敗：%v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("切換測試目錄失敗：%v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(original); err != nil {
			t.Errorf("還原測試目錄失敗：%v", err)
		}
	})
}

func TestTemplateEngineLoadConfigUsesFileOrDefaults(t *testing.T) {
	tests := []struct {
		name        string
		content     *string
		wantVersion string
		wantLinks   bool
	}{
		{
			name: "valid config",
			content: stringPointer(`
template_config:
  version: "2.0.0"
  supported_languages:
    - code: tw
      name: 繁體中文
  fallback_order: [tw]
  naming_convention:
    prefix: custom_
    supported_extensions: [.tmpl]
    priority_order: [.tmpl]
  format_options:
    show_links:
      enabled: false
`),
			wantVersion: "2.0.0",
			wantLinks:   false,
		},
		{
			name:        "invalid yaml falls back",
			content:     stringPointer("template_config: [invalid"),
			wantVersion: "1.0.0",
			wantLinks:   true,
		},
		{
			name:        "missing file falls back",
			content:     nil,
			wantVersion: "1.0.0",
			wantLinks:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.yaml")
			if tt.content != nil {
				if err := os.WriteFile(path, []byte(*tt.content), 0o600); err != nil {
					t.Fatalf("寫入測試設定失敗：%v", err)
				}
			}

			engine := NewTemplateEngine()
			if err := engine.LoadConfig(path); err != nil {
				t.Fatalf("LoadConfig 不應回傳錯誤：%v", err)
			}
			if engine.config.Version != tt.wantVersion || engine.config.FormatOptions.ShowLinks.Enabled != tt.wantLinks {
				t.Fatalf("載入設定不符預期：%+v", engine.config)
			}
		})
	}
}

func stringPointer(value string) *string {
	return &value
}

func TestTemplateEngineProfileConfigFallbacks(t *testing.T) {
	changeWorkingDirectory(t, t.TempDir())
	engine := NewTemplateEngine()

	if err := engine.LoadConfigFromConfigs(); err != nil {
		t.Fatalf("缺少 configs 時應使用預設值：%v", err)
	}
	if err := engine.LoadConfigWithProfile("missing"); err != nil {
		t.Fatalf("缺少 profile 時應 fallback：%v", err)
	}
	if err := engine.ReloadConfigWithProfile(""); err != nil {
		t.Fatalf("重新載入缺少的預設 profile 應 fallback：%v", err)
	}
	if engine.config.Version != "1.0.0" {
		t.Fatalf("fallback config version 不符預期：%q", engine.config.Version)
	}
}

func TestTemplateEngineLoadsProfileFromWorkingDirectory(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "configs")
	if err := os.Mkdir(configDir, 0o700); err != nil {
		t.Fatalf("建立 configs 目錄失敗：%v", err)
	}
	writeTemplateFile(t, configDir, "alert_config.custom.yaml", "template_config:\n  version: custom\n")
	changeWorkingDirectory(t, root)

	engine := NewTemplateEngine()
	if err := engine.LoadConfigWithProfile("custom"); err != nil {
		t.Fatalf("載入指定 profile 失敗：%v", err)
	}
	if engine.config.Version != "custom" {
		t.Fatalf("profile version 不符預期：%q", engine.config.Version)
	}
	if err := engine.ReloadConfigWithProfile("custom"); err != nil {
		t.Fatalf("重新載入指定 profile 失敗：%v", err)
	}
}

func TestTemplateEngineLanguageQueriesAndFallbacks(t *testing.T) {
	engine := NewTemplateEngine()
	engine.config = &TemplateConfig{
		SupportedLanguages: []LanguageConfig{
			{Code: "eng", Name: "English"},
			{Code: "tw", Name: "繁體中文"},
		},
		FallbackOrder: []string{"eng", "tw"},
	}
	engine.templates["eng"] = texttemplate.Must(texttemplate.New("eng").Parse("English"))
	engine.templates["tw"] = texttemplate.Must(texttemplate.New("tw").Parse("繁體中文"))

	languages := engine.GetAvailableLanguages()
	sort.Strings(languages)
	if !reflect.DeepEqual(languages, []string{"eng", "tw"}) {
		t.Fatalf("可用語言不符預期：%v", languages)
	}
	if !reflect.DeepEqual(engine.GetSupportedLanguages(), []string{"eng", "tw"}) {
		t.Fatalf("支援語言不符預期：%v", engine.GetSupportedLanguages())
	}
	if !reflect.DeepEqual(engine.GetSupportedLanguageDetails(), engine.config.SupportedLanguages) {
		t.Fatalf("語言詳細資訊不符預期：%v", engine.GetSupportedLanguageDetails())
	}
	if !engine.HasLanguage("tw") || engine.HasLanguage("ja") {
		t.Fatal("HasLanguage 結果不符預期")
	}
	if got := engine.GetDefaultLanguage("tw"); got != "tw" {
		t.Fatalf("已載入 preferred language 應直接使用：%q", got)
	}
	if got := engine.GetDefaultLanguage("ja"); got != "eng" {
		t.Fatalf("fallback language 不符預期：%q", got)
	}

	engine.config.FallbackOrder = []string{"ja"}
	if got := engine.GetDefaultLanguage("ko"); got != "eng" && got != "tw" {
		t.Fatalf("應回退至任一已載入語言：%q", got)
	}
	engine.templates = map[string]*texttemplate.Template{}
	if got := engine.GetDefaultLanguage("ko"); got != "ko" {
		t.Fatalf("沒有模板時應保留 preferred language：%q", got)
	}

	engine.config = nil
	if len(engine.GetSupportedLanguages()) != 0 || len(engine.GetSupportedLanguageDetails()) != 0 {
		t.Fatal("nil config 應回傳空語言資訊")
	}
}

func TestTemplateEngineRenderTemplateAndPlatform(t *testing.T) {
	engine := NewTemplateEngine()
	engine.templates["eng"] = texttemplate.Must(texttemplate.New("eng").Parse("{{ .Platform }}|{{ .Status }}|{{ .FormatOptions.MaxSummaryLength.Value }}"))

	got, err := engine.RenderTemplateForPlatform("eng", "slack", TemplateData{Status: "firing"})
	if err != nil {
		t.Fatalf("平台模板渲染失敗：%v", err)
	}
	if got != "slack|firing|200" {
		t.Fatalf("平台或預設 FormatOptions 不符預期：%q", got)
	}

	custom := FormatOptions{}
	custom.MaxSummaryLength.Value = 55
	got, err = engine.RenderTemplate("eng", TemplateData{Status: "resolved", FormatOptions: custom})
	if err != nil || got != "|resolved|55" {
		t.Fatalf("自訂 FormatOptions 未保留：got=%q err=%v", got, err)
	}

	if _, err := engine.RenderTemplate("missing", TemplateData{}); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("缺少模板錯誤不符預期：%v", err)
	}
	if _, err := engine.RenderTemplateForPlatform("missing", "discord", TemplateData{}); err == nil {
		t.Fatal("平台渲染缺少模板應失敗")
	}

	engine.templates["broken"] = texttemplate.Must(texttemplate.New("broken").Parse("{{ index .Alerts 2 }}"))
	if _, err := engine.RenderTemplate("broken", TemplateData{}); err == nil || !strings.Contains(err.Error(), "failed to execute") {
		t.Fatalf("模板執行錯誤不符預期：%v", err)
	}
}

func TestTemplateEngineConfigFilesAndDefaults(t *testing.T) {
	engine := NewTemplateEngine()
	dir := t.TempDir()
	valid := writeTemplateFile(t, dir, "valid.yaml", `
template_config:
  version: "3.0.0"
  format_options:
    max_summary_length:
      value: 77
`)
	invalid := writeTemplateFile(t, dir, "invalid.yaml", "template_config: [invalid")

	loaded, err := engine.loadConfigFromFile(valid)
	if err != nil || loaded.Version != "3.0.0" || loaded.FormatOptions.MaxSummaryLength.Value != 77 {
		t.Fatalf("載入設定檔不符預期：config=%+v err=%v", loaded, err)
	}
	if _, err := engine.loadConfigFromFile(filepath.Join(dir, "missing.yaml")); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("不存在設定錯誤不符預期：%v", err)
	}
	if _, err := engine.loadConfigFromFile(invalid); err == nil || !strings.Contains(err.Error(), "failed to parse") {
		t.Fatalf("無效設定錯誤不符預期：%v", err)
	}

	minimal := engine.GetMinimalDefaultConfig()
	if minimal.FormatOptions.MaxSummaryLength.Value != 100 || len(minimal.SupportedLanguages) != 4 {
		t.Fatalf("minimal defaults 不符預期：%+v", minimal)
	}
	full := engine.GetFullDefaultConfig()
	if full.FormatOptions.MaxSummaryLength.Value != 200 || len(full.SupportedLanguages) != 5 {
		t.Fatalf("full defaults 不符預期：%+v", full)
	}
	if engine.GetCurrentFormatOptions().MaxSummaryLength.Value != 200 {
		t.Fatal("目前 format options 不符預期")
	}
	var nilEngine *TemplateEngine
	if nilEngine.GetCurrentFormatOptions().MaxSummaryLength.Value != 200 {
		t.Fatal("nil engine 應回傳 full defaults")
	}
	engine.config = nil
	if engine.GetCurrentFormatOptions().MaxSummaryLength.Value != 200 {
		t.Fatal("nil config 應回傳 full defaults")
	}
}

func TestTemplateEngineLoadsMinimalConfigFromExpectedPath(t *testing.T) {
	root := t.TempDir()
	configDir := filepath.Join(root, "configs")
	if err := os.Mkdir(configDir, 0o700); err != nil {
		t.Fatalf("建立 configs 目錄失敗：%v", err)
	}
	writeTemplateFile(t, configDir, "alert_config.minimal.yaml", `
template_config:
  version: custom-minimal
  format_options:
    max_summary_length:
      value: 33
`)
	changeWorkingDirectory(t, root)

	engine := NewTemplateEngine()
	config, err := engine.loadMinimalConfig()
	if err != nil || config.Version != "custom-minimal" {
		t.Fatalf("載入 minimal config 失敗：config=%+v err=%v", config, err)
	}
	if got := engine.GetMinimalDefaultConfig(); got.FormatOptions.MaxSummaryLength.Value != 33 {
		t.Fatalf("GetMinimalDefaultConfig 未使用檔案：%+v", got)
	}
}

func TestTemplateEngineTimeAndTextFormatting(t *testing.T) {
	engine := NewTemplateEngine()
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: "未設定"},
		{name: "zero time", input: "0001-01-01T00:00:00Z", want: "未設定"},
		{name: "invalid", input: "not-time", want: "not-time"},
		{name: "RFC3339", input: "2026-10-21T07:08:09Z", want: "2026-10-21 07:08:09"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := engine.formatTime(tt.input); got != tt.want {
				t.Fatalf("時間格式不符預期：got=%q want=%q", got, tt.want)
			}
			if got := engine.formatTimeForPlatform("telegram", tt.input); got != tt.want {
				t.Fatalf("平台時間格式不符預期：got=%q want=%q", got, tt.want)
			}
		})
	}
	if got := engine.formatTextForPlatform("slack", "原始文字"); got != "原始文字" {
		t.Fatalf("一般文字不應變更：%q", got)
	}
	if got := engine.formatBoldForPlatform("slack", "text"); got != "*text*" {
		t.Fatalf("Markdown 粗體不符預期：%q", got)
	}
	if got := engine.formatItalicForPlatform("discord", "text"); got != "_text_" {
		t.Fatalf("Markdown 斜體不符預期：%q", got)
	}
	if got := engine.formatCodeForPlatform("unknown", "text"); got != "`text`" {
		t.Fatalf("Markdown code 不符預期：%q", got)
	}
	if got := engine.formatLinkForPlatform("unknown", "https://example.test", "example"); got != "[example](https://example.test)" {
		t.Fatalf("預設 link 格式不符預期：%q", got)
	}
}
