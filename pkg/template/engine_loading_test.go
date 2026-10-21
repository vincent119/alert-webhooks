package template

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeTemplateFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("寫入測試模板失敗：%v", err)
	}
	return path
}

func TestTemplateEngineLoadTemplatesRejectsInvalidDirectories(t *testing.T) {
	tests := []struct {
		name    string
		path    func(*testing.T) string
		wantErr string
	}{
		{
			name: "missing directory",
			path: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "missing")
			},
			wantErr: "template directory does not exist",
		},
		{
			name: "empty directory",
			path: func(t *testing.T) string {
				return t.TempDir()
			},
			wantErr: "no template files found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewTemplateEngine()
			err := engine.LoadTemplates(tt.path(t))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateEngineScanTemplateFilesFiltersAndPrioritizes(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "alert_template_tw.j2", "Jinja2")
	preferred := writeTemplateFile(t, dir, "alert_template_tw.tmpl", "Go template")
	english := writeTemplateFile(t, dir, "alert_template_eng.tmpl", "English")
	writeTemplateFile(t, dir, "alert_template_.tmpl", "invalid language")
	writeTemplateFile(t, dir, "alert_template_ja.txt", "unsupported")
	writeTemplateFile(t, dir, "other_template_ko.tmpl", "wrong prefix")
	if err := os.Mkdir(filepath.Join(dir, "alert_template_subdir.tmpl"), 0o700); err != nil {
		t.Fatalf("建立測試子目錄失敗：%v", err)
	}

	engine := NewTemplateEngine()
	files, err := engine.scanTemplateFiles(dir)
	if err != nil {
		t.Fatalf("掃描模板失敗：%v", err)
	}
	if len(files) != 2 {
		t.Fatalf("模板數量不符預期：got=%v", files)
	}
	if files["tw"] != preferred || files["eng"] != english {
		t.Fatalf("模板選擇不符預期：%v", files)
	}
	if got := engine.getExtensionPriority(".tmpl", []string{".tmpl", ".j2"}); got != 0 {
		t.Fatalf("已知副檔名優先權不符預期：%d", got)
	}
	if got := engine.getExtensionPriority(".txt", []string{".tmpl", ".j2"}); got != 2 {
		t.Fatalf("未知副檔名應排在最後：%d", got)
	}
}

func TestTemplateEngineScanTemplateFilesReportsReadError(t *testing.T) {
	engine := NewTemplateEngine()
	_, err := engine.scanTemplateFiles(filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("掃描不存在目錄應回傳錯誤")
	}
}

func TestTemplateEngineLoadTemplatesKeepsValidTemplates(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "alert_template_eng.tmpl", "status={{ .Status }}")
	writeTemplateFile(t, dir, "alert_template_tw.tmpl", "{{ invalid syntax")

	engine := NewTemplateEngine()
	if err := engine.LoadTemplates(dir); err != nil {
		t.Fatalf("至少一個模板有效時應成功載入：%v", err)
	}
	if !engine.HasLanguage("eng") || engine.HasLanguage("tw") {
		t.Fatalf("已載入語言不符預期：%v", engine.GetAvailableLanguages())
	}

	got, err := engine.RenderTemplate("eng", TemplateData{Status: "firing"})
	if err != nil || got != "status=firing" {
		t.Fatalf("模板渲染不符預期：got=%q err=%v", got, err)
	}
}

func TestTemplateEngineLoadTemplatesFailsWhenEveryTemplateIsInvalid(t *testing.T) {
	dir := t.TempDir()
	writeTemplateFile(t, dir, "alert_template_eng.tmpl", "{{ invalid syntax")

	engine := NewTemplateEngine()
	err := engine.LoadTemplates(dir)
	if err == nil || !strings.Contains(err.Error(), "failed to load any templates") {
		t.Fatalf("錯誤不符預期：%v", err)
	}
}

func TestTemplateEngineLoadsConvertedJinjaTemplate(t *testing.T) {
	dir := t.TempDir()
	path := writeTemplateFile(t, dir, "alert_template_eng.j2", strings.Join([]string{
		"{{ status }} {{ alert_name }} {{ env }} {{ severity }} {{ namespace }}",
		"{% if firing_count > 0 %}firing={{ firing_count }}{% elif resolved_count > 0 %}resolved={{ resolved_count }}{% else %}empty{% endif %}",
		"{% for alert in alerts %}{{ loop.index }}:{{ alert.status }}:{{ alert.annotations.summary }}:{{ alert.labels.pod }}{% endfor %}",
		"{% if externalURL %}{{ externalURL }}{% endif %}",
	}, "\n"))

	engine := NewTemplateEngine()
	tmpl, err := engine.loadTemplate(path)
	if err != nil {
		t.Fatalf("載入 Jinja2 模板失敗：%v", err)
	}
	engine.templates["eng"] = tmpl
	got, err := engine.RenderTemplate("eng", TemplateData{
		Status:      "firing",
		AlertName:   "HighCPU",
		Env:         "production",
		Severity:    "critical",
		Namespace:   "monitoring",
		FiringCount: 1,
		Alerts: []AlertData{{
			Status:      "firing",
			Labels:      map[string]string{"pod": "api-1"},
			Annotations: map[string]string{"summary": "CPU high"},
		}},
		ExternalURL: "https://example.test",
	})
	if err != nil {
		t.Fatalf("渲染 Jinja2 轉換模板失敗：%v", err)
	}
	for _, fragment := range []string{"firing HighCPU production critical monitoring", "firing=1", "1:firing:CPU high:api-1", "https://example.test"} {
		if !strings.Contains(got, fragment) {
			t.Fatalf("渲染結果缺少 %q：%s", fragment, got)
		}
	}
}

func TestTemplateEngineValidateAndReloadTemplates(t *testing.T) {
	firstDir := t.TempDir()
	validPath := writeTemplateFile(t, firstDir, "alert_template_eng.tmpl", "English")
	invalidPath := writeTemplateFile(t, firstDir, "alert_template_tw.tmpl", "{{ invalid")

	engine := NewTemplateEngine()
	if err := engine.ValidateTemplate(validPath); err != nil {
		t.Fatalf("有效模板驗證失敗：%v", err)
	}
	if err := engine.ValidateTemplate(invalidPath); err == nil {
		t.Fatal("無效模板驗證應失敗")
	}
	if err := engine.ValidateTemplate(filepath.Join(firstDir, "missing.tmpl")); err == nil {
		t.Fatal("不存在模板驗證應失敗")
	}
	if err := engine.LoadTemplates(firstDir); err != nil {
		t.Fatalf("載入第一批模板失敗：%v", err)
	}

	secondDir := t.TempDir()
	writeTemplateFile(t, secondDir, "alert_template_tw.tmpl", "繁體中文")
	if err := engine.ReloadTemplates(secondDir); err != nil {
		t.Fatalf("重新載入模板失敗：%v", err)
	}
	if engine.HasLanguage("eng") || !engine.HasLanguage("tw") {
		t.Fatalf("重新載入後語言不符預期：%v", engine.GetAvailableLanguages())
	}
}
