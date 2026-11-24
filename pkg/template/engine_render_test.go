package template

import (
	"html"
	"path/filepath"
	"strings"
	"testing"
)

func TestTelegramTemplatesEscapeAlertText(t *testing.T) {
	engine := NewTemplateEngine()
	if err := engine.LoadTemplates(filepath.Join("..", "..", "templates", "alerts")); err != nil {
		t.Fatalf("載入內建模板失敗：%v", err)
	}
	for _, language := range []string{"tw", "zh", "eng", "ja", "ko"} {
		for _, status := range []string{"firing", "resolved"} {
			t.Run(language+"/"+status, func(t *testing.T) {
				values := []string{
					"名稱 (<5%)", "環境 <production>", "嚴重程度 A > B & C",
					"命名空間 <monitoring>", "摘要 (<5%) &lt;", `描述 <b>文字</b> & "'`,
					status + " <狀態>",
				}
				pod := "pod<1>&"
				url := "https://example.com/alerts?name=cpu&severity=high"
				data := TemplateData{
					Status: values[6], AlertName: values[0], Env: values[1], Severity: values[2], Namespace: values[3], TotalAlerts: 1,
					Alerts:      []AlertData{{Status: status, Annotations: map[string]string{"summary": values[4], "description": values[5]}, Labels: map[string]string{"pod": pod}, GeneratorURL: url}},
					ExternalURL: url,
				}
				data.FormatOptions.ShowGeneratorURL.Enabled = true
				data.FormatOptions.ShowExternalURL.Enabled = true
				if status == "firing" {
					data.FiringCount = 1
				} else {
					data.ResolvedCount = 1
				}
				got, err := engine.RenderTemplateForPlatform(language, "telegram", data)
				if err != nil {
					t.Fatalf("渲染失敗：%v", err)
				}
				for _, value := range values {
					if !strings.Contains(got, html.EscapeString(value)) {
						t.Errorf("缺少跳脫後文字：%q", value)
					}
					if strings.Contains(got, value) {
						t.Errorf("外部文字未跳脫：%q", value)
					}
				}
				for _, fragment := range []string{"<b>", "</b>", "<i>", "</i>", "<code>" + html.EscapeString(pod) + "</code>", `<a href="` + html.EscapeString(url) + `">`} {
					if !strings.Contains(got, fragment) {
						t.Errorf("格式標記或連結不符預期：%q", fragment)
					}
				}
				if strings.Contains(got, "<5%)") {
					t.Error("比較符號被輸出成標籤")
				}
			})
		}
	}
}
