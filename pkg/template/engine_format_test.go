package template

import (
	"html"
	"testing"
)

func TestTelegramFormattersEscapeExternalText(t *testing.T) {
	engine := NewTemplateEngine()
	input := `<b>告警</b> & "'`
	escaped := html.EscapeString(input)

	tests := []struct {
		name string
		got  string
		want string
	}{
		{
			name: "粗體",
			got:  engine.formatBoldForPlatform("telegram", input),
			want: "<b>" + escaped + "</b>",
		},
		{
			name: "斜體",
			got:  engine.formatItalicForPlatform("telegram", input),
			want: "<i>" + escaped + "</i>",
		},
		{
			name: "程式碼",
			got:  engine.formatCodeForPlatform("telegram", input),
			want: "<code>" + escaped + "</code>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("format result is incorrect: got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestTelegramFormatLinkAllowsOnlyAbsoluteHTTPURLs(t *testing.T) {
	engine := NewTemplateEngine()

	validURL := "https://example.com/alerts?name=cpu&severity=high"
	validText := `<b>CPU</b> & "高"`
	if got, want := engine.formatLinkForPlatform("telegram", validURL, validText), `<a href="`+html.EscapeString(validURL)+`">`+html.EscapeString(validText)+`</a>`; got != want {
		t.Fatalf("valid HTTPS link format is incorrect: got %q, want %q", got, want)
	}

	if got, want := engine.formatLinkForPlatform("telegram", "http://example.com", ""), `<a href="http://example.com">http://example.com</a>`; got != want {
		t.Fatalf("empty text should use the valid URL as link text: got %q, want %q", got, want)
	}
}

func TestTelegramFormatLinkRejectsDisallowedURLs(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name string
		url  string
		text string
		want string
	}{
		{
			name: "javascript scheme",
			url:  "javascript:alert(1)",
			text: `<b>不可點擊</b>`,
			want: html.EscapeString(`<b>不可點擊</b>`),
		},
		{
			name: "Telegram deep link",
			url:  "tg://user?id=1",
			text: "Telegram 使用者",
			want: "Telegram 使用者",
		},
		{
			name: "相對 URL",
			url:  "/alerts/123",
			text: "告警",
			want: "告警",
		},
		{
			name: "HTTP URL 缺少 host",
			url:  "https:/alerts/123",
			text: "告警",
			want: "告警",
		},
		{
			name: "無文字時回退為 URL",
			url:  `javascript:alert("<b>告警</b>")`,
			want: html.EscapeString(`javascript:alert("<b>告警</b>")`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.formatLinkForPlatform("telegram", tt.url, tt.text)
			if got != tt.want {
				t.Fatalf("disallowed URL fallback result is incorrect: got %q, want %q", got, tt.want)
			}
			if containsAnchor(got) {
				t.Fatalf("disallowed URL must not produce a link: %q", got)
			}
		})
	}
}

func TestNonTelegramFormattersPreserveExistingOutput(t *testing.T) {
	engine := NewTemplateEngine()
	url := "javascript:alert(1)"
	text := `<b>告警</b>`

	if got, want := engine.formatLinkForPlatform("slack", url, text), "<"+url+"|"+text+">"; got != want {
		t.Fatalf("Slack format must not change: got %q, want %q", got, want)
	}
	if got, want := engine.formatLinkForPlatform("discord", url, text), "["+text+"]("+url+")"; got != want {
		t.Fatalf("Discord format must not change: got %q, want %q", got, want)
	}
}

func containsAnchor(value string) bool {
	return len(value) >= 2 && value[:2] == "<a"
}

func TestTelegramFormatTextEscapesPlainText(t *testing.T) {
	engine := NewTemplateEngine()
	tests := []struct{ name, input, want string }{
		{"比較符號", "剩餘空間 (<5%)", "剩餘空間 (&lt;5%)"},
		{"大於與連接符號", "A > B & C", "A &gt; B &amp; C"},
		{"引號", "\"'", "&#34;&#39;"},
		{"標籤外觀", "<b>文字</b>", "&lt;b&gt;文字&lt;/b&gt;"},
		{"中文", "磁碟空間不足", "磁碟空間不足"},
		{"空字串", "", ""},
		{"既有實體文字", "&lt;", "&amp;lt;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.formatTextForPlatform("telegram", tt.input)
			if got != tt.want {
				t.Fatalf("普通文字跳脫錯誤：got=%q want=%q", got, tt.want)
			}
			if html.UnescapeString(got) != tt.input {
				t.Fatalf("解碼後無法還原原文：%q", got)
			}
		})
	}
}

func TestNonTelegramFormatTextPreservesInput(t *testing.T) {
	engine := NewTemplateEngine()
	input := `<b>告警</b> (<5%) &lt; & "'`
	for _, platform := range []string{"slack", "discord", "", "unknown"} {
		t.Run(platform, func(t *testing.T) {
			if got := engine.formatTextForPlatform(platform, input); got != input {
				t.Fatalf("平台 %q 的原始文字被改變：%q", platform, got)
			}
		})
	}
}
