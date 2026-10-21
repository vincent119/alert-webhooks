package alertmodel

import (
	"reflect"
	"testing"

	"alert-webhooks/pkg/template"
)

func TestBuildTemplateDataUsesCommonLabelsAndConvertsAlerts(t *testing.T) {
	formatOptions := template.FormatOptions{}
	formatOptions.ShowLinks.Enabled = true
	formatOptions.MaxSummaryLength.Value = 120

	alerts := []map[string]interface{}{
		{
			"status": "firing",
			"labels": map[string]interface{}{
				"alertname": "FirstAlert",
				"instance":  "server-1",
				"ignored":   42,
			},
			"annotations": map[string]interface{}{
				"summary": "CPU 過高",
				"ignored": true,
			},
			"startsAt":     "2026-10-21T07:00:00Z",
			"endsAt":       "2026-10-21T08:00:00Z",
			"generatorURL": "https://example.test/graph",
		},
		{"status": "resolved"},
		{"status": "unknown"},
		{"status": 123},
	}

	got := BuildTemplateData(
		"firing",
		alerts,
		map[string]interface{}{"group": "ignored"},
		map[string]interface{}{
			"alertname": "CommonAlert",
			"env":       "production",
			"severity":  "critical",
			"namespace": "monitoring",
		},
		map[string]interface{}{"description": "ignored"},
		"https://alertmanager.example.test",
		formatOptions,
	)

	if got.Status != "firing" || got.AlertName != "CommonAlert" || got.Env != "production" {
		t.Fatalf("基本欄位轉換不符預期：%+v", got)
	}
	if got.Severity != "critical" || got.Namespace != "monitoring" {
		t.Fatalf("通用標籤未優先套用：%+v", got)
	}
	if got.TotalAlerts != 4 || got.FiringCount != 1 || got.ResolvedCount != 1 {
		t.Fatalf("警報計數不符預期：total=%d firing=%d resolved=%d", got.TotalAlerts, got.FiringCount, got.ResolvedCount)
	}
	if got.ExternalURL != "https://alertmanager.example.test" {
		t.Fatalf("ExternalURL 不符預期：%q", got.ExternalURL)
	}
	if !reflect.DeepEqual(got.FormatOptions, formatOptions) {
		t.Fatalf("FormatOptions 未完整保留：got=%+v want=%+v", got.FormatOptions, formatOptions)
	}

	wantAlert := template.AlertData{
		Status:       "firing",
		Labels:       map[string]string{"alertname": "FirstAlert", "instance": "server-1"},
		Annotations:  map[string]string{"summary": "CPU 過高"},
		StartsAt:     "2026-10-21T07:00:00Z",
		EndsAt:       "2026-10-21T08:00:00Z",
		GeneratorURL: "https://example.test/graph",
	}
	if !reflect.DeepEqual(got.Alerts[0], wantAlert) {
		t.Fatalf("警報資料轉換不符預期：got=%+v want=%+v", got.Alerts[0], wantAlert)
	}
	if len(got.Alerts[1].Labels) != 0 || len(got.Alerts[1].Annotations) != 0 {
		t.Fatalf("缺少 labels 或 annotations 時應建立空 map：%+v", got.Alerts[1])
	}
}

func TestBuildTemplateDataFallsBackToFirstAlertLabels(t *testing.T) {
	tests := []struct {
		name         string
		commonLabels map[string]interface{}
	}{
		{name: "nil common labels", commonLabels: nil},
		{
			name: "empty and non-string common labels",
			commonLabels: map[string]interface{}{
				"alertname": "",
				"env":       42,
				"severity":  "",
				"namespace": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := []map[string]interface{}{{
				"labels": map[string]interface{}{
					"alertname": "FallbackAlert",
					"env":       "staging",
					"severity":  "warning",
					"namespace": "default",
				},
			}}

			got := BuildTemplateData("firing", alerts, nil, tt.commonLabels, nil, "", template.FormatOptions{})

			if got.AlertName != "FallbackAlert" || got.Env != "staging" || got.Severity != "warning" || got.Namespace != "default" {
				t.Fatalf("未從第一筆警報回退基本欄位：%+v", got)
			}
		})
	}
}

func TestBuildTemplateDataHandlesEmptyAndMalformedAlerts(t *testing.T) {
	tests := []struct {
		name   string
		alerts []map[string]interface{}
	}{
		{name: "empty alerts", alerts: nil},
		{name: "malformed first alert", alerts: []map[string]interface{}{{"labels": "invalid", "annotations": 1}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildTemplateData("resolved", tt.alerts, nil, nil, nil, "", template.FormatOptions{})

			if got.TotalAlerts != len(tt.alerts) || got.FiringCount != 0 || got.ResolvedCount != 0 {
				t.Fatalf("空白或格式錯誤警報的計數不符預期：%+v", got)
			}
			if got.AlertName != "" || got.Env != "" || got.Severity != "" || got.Namespace != "" {
				t.Fatalf("不應從格式錯誤的 labels 取得欄位：%+v", got)
			}
			if len(tt.alerts) == 0 && got.Alerts != nil {
				t.Fatalf("空 alerts 應保留 nil slice：%+v", got.Alerts)
			}
		})
	}
}
