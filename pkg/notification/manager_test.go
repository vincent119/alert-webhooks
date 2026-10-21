package notification

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"alert-webhooks/config"
	"alert-webhooks/pkg/notification/types"
	"alert-webhooks/pkg/template"
)

type fakeProvider struct {
	name        string
	enabled     bool
	sendErr     error
	requests    []*types.NotificationRequest
	status      *types.ProviderStatus
	validateErr error
	connectErr  error
}

func (f *fakeProvider) GetName() string {
	return f.name
}

func (f *fakeProvider) SendMessage(_ context.Context, req *types.NotificationRequest) error {
	f.requests = append(f.requests, req)
	return f.sendErr
}

func (f *fakeProvider) GetStatus() *types.ProviderStatus {
	return f.status
}

func (f *fakeProvider) GetCapabilities() *types.ProviderCapabilities {
	return &types.ProviderCapabilities{}
}

func (f *fakeProvider) IsEnabled() bool {
	return f.enabled
}

func (f *fakeProvider) ValidateConfig() error {
	return f.validateErr
}

func (f *fakeProvider) TestConnection() error {
	return f.connectErr
}

type managerTelegramService struct{}

func (managerTelegramService) SendMessage(context.Context, int, string) error { return nil }
func (managerTelegramService) GetBotInfo() (interface{}, error)               { return struct{}{}, nil }

type managerSlackService struct{}

func (managerSlackService) SendMessage(context.Context, string, string) error { return nil }
func (managerSlackService) SendMessageToLevel(context.Context, string, string) error {
	return nil
}
func (managerSlackService) TestConnection() error { return nil }

type managerDiscordService struct{}

func (managerDiscordService) SendMessage(context.Context, string, string) error { return nil }
func (managerDiscordService) SendMessageToChannel(string, string) error         { return nil }
func (managerDiscordService) SendMessageToLevel(context.Context, string, string) error {
	return nil
}
func (managerDiscordService) TestConnection() error            { return nil }
func (managerDiscordService) ValidateChannel(string) error     { return nil }
func (managerDiscordService) GetBotInfo() (interface{}, error) { return struct{}{}, nil }

func loadManagerConfig(t *testing.T, content string) {
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

func TestGetNotificationManagerReturnsSingleton(t *testing.T) {
	first := GetNotificationManager()
	second := GetNotificationManager()
	if first != second {
		t.Fatal("GetNotificationManager 應回傳相同 instance")
	}
}

func TestNotificationManagerInitializeAndReloadProviders(t *testing.T) {
	loadManagerConfig(t, `
telegram:
  enable: true
slack:
  enable: true
discord:
  enable: true
`)
	engine := template.NewTemplateEngine()
	manager := &NotificationManager{}

	if err := manager.Initialize(engine, managerTelegramService{}, managerSlackService{}, managerDiscordService{}); err != nil {
		t.Fatalf("初始化 manager 失敗：%v", err)
	}
	for _, name := range []string{"telegram", "slack", "discord"} {
		if _, exists := manager.GetProvider(name); !exists {
			t.Fatalf("未註冊 %s provider", name)
		}
	}

	loadManagerConfig(t, "telegram:\n  enable: false\nslack:\n  enable: false\ndiscord:\n  enable: false\n")
	if err := manager.Reload(engine, managerTelegramService{}, managerSlackService{}, managerDiscordService{}); err != nil {
		t.Fatalf("重新載入 manager 失敗：%v", err)
	}
	if got := manager.GetAllProviders(); len(got) != 0 {
		t.Fatalf("停用所有 provider 後仍有註冊項目：%v", got)
	}
}

func TestNotificationManagerSendNotificationResponses(t *testing.T) {
	wantErr := errors.New("send failed")
	tests := []struct {
		name             string
		providers        map[string]types.NotificationProvider
		providerName     string
		req              types.NotificationRequest
		wantSuccess      bool
		wantErr          string
		wantMessage      string
		wantRequestCount int
	}{
		{
			name:         "missing provider",
			providers:    map[string]types.NotificationProvider{},
			providerName: "missing",
			wantErr:      "not found",
			wantMessage:  "Provider 'missing' not found",
		},
		{
			name: "disabled provider",
			providers: map[string]types.NotificationProvider{
				"test": &fakeProvider{name: "test", enabled: false},
			},
			providerName: "test",
			wantErr:      "disabled",
			wantMessage:  "Provider 'test' is disabled",
		},
		{
			name: "send error",
			providers: map[string]types.NotificationProvider{
				"test": &fakeProvider{name: "test", enabled: true, sendErr: wantErr},
			},
			providerName:     "test",
			req:              types.NotificationRequest{Level: "L1", ChatID: "room", Channel: "channel", Message: "test"},
			wantErr:          "send failed",
			wantMessage:      "Failed to send message: send failed",
			wantRequestCount: 1,
		},
		{
			name: "success",
			providers: map[string]types.NotificationProvider{
				"test": &fakeProvider{name: "test", enabled: true},
			},
			providerName:     "test",
			req:              types.NotificationRequest{Level: "L2", ChatID: "room", Channel: "channel", Message: "test"},
			wantSuccess:      true,
			wantMessage:      "Message sent successfully",
			wantRequestCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := &NotificationManager{providers: tt.providers}
			response, err := manager.SendNotification(context.Background(), tt.providerName, &tt.req)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("發送通知失敗：%v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
			if response.Success != tt.wantSuccess || response.Message != tt.wantMessage || response.Provider != tt.providerName {
				t.Fatalf("response 不符預期：%+v", response)
			}
			if tt.wantRequestCount > 0 {
				provider := tt.providers[tt.providerName].(*fakeProvider)
				if len(provider.requests) != tt.wantRequestCount {
					t.Fatalf("provider 呼叫次數不符預期：%d", len(provider.requests))
				}
				if response.Level != tt.req.Level || response.ChatID != tt.req.ChatID || response.Channel != tt.req.Channel {
					t.Fatalf("response metadata 不符預期：%+v", response)
				}
			}
		})
	}
}

func TestNotificationManagerPreprocessRendersAlertMessage(t *testing.T) {
	engine := template.NewTemplateEngine()
	if err := engine.LoadTemplates(filepath.Join("..", "..", "templates", "alerts")); err != nil {
		t.Fatalf("載入測試模板失敗：%v", err)
	}
	provider := &fakeProvider{name: "test", enabled: true}
	manager := &NotificationManager{
		providers:      map[string]types.NotificationProvider{"test": provider},
		templateEngine: engine,
	}
	req := &types.NotificationRequest{
		TemplateLanguage: "eng",
		AlertData: &types.AlertManagerData{
			Status: "firing",
			Alerts: []map[string]interface{}{{
				"status":      "firing",
				"labels":      map[string]interface{}{"alertname": "HighCPU"},
				"annotations": map[string]interface{}{"summary": "CPU high"},
			}},
		},
	}

	response, err := manager.SendNotification(context.Background(), "test", req)
	if err != nil || !response.Success {
		t.Fatalf("渲染後發送失敗：response=%+v err=%v", response, err)
	}
	if req.Message == "" || len(provider.requests) != 1 {
		t.Fatalf("模板未產生訊息或未呼叫 provider：message=%q requests=%d", req.Message, len(provider.requests))
	}
}

func TestNotificationManagerPreprocessReportsRenderError(t *testing.T) {
	provider := &fakeProvider{name: "test", enabled: true}
	manager := &NotificationManager{
		providers:      map[string]types.NotificationProvider{"test": provider},
		templateEngine: template.NewTemplateEngine(),
	}
	req := &types.NotificationRequest{
		AlertData: &types.AlertManagerData{Status: "firing"},
	}

	response, err := manager.SendNotification(context.Background(), "test", req)
	if err == nil || !strings.Contains(err.Error(), "template for language") {
		t.Fatalf("模板錯誤不符預期：%v", err)
	}
	if response.Success || !strings.Contains(response.Message, "Failed to preprocess request") {
		t.Fatalf("response 不符預期：%+v", response)
	}
	if len(provider.requests) != 0 {
		t.Fatal("預處理失敗後不應呼叫 provider")
	}
}

func TestNotificationManagerTemplateLanguagePrecedence(t *testing.T) {
	loadManagerConfig(t, `
telegram:
  template_language: tw
slack:
  template_language: ja
`)
	manager := &NotificationManager{}
	tests := []struct {
		name      string
		provider  string
		requested string
		want      string
	}{
		{name: "request wins", provider: "telegram", requested: "ko", want: "ko"},
		{name: "telegram config", provider: "telegram", want: "tw"},
		{name: "slack config", provider: "slack", want: "ja"},
		{name: "default language", provider: "discord", want: "eng"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := manager.getProviderTemplateLanguage(tt.provider, tt.requested); got != tt.want {
				t.Fatalf("語言選擇不符預期：got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestNotificationManagerConvertsAlertManagerData(t *testing.T) {
	manager := &NotificationManager{}
	if _, err := manager.convertAlertManagerData(nil); err == nil || !strings.Contains(err.Error(), "is nil") {
		t.Fatalf("nil data 應回傳錯誤：%v", err)
	}

	data := &types.AlertManagerData{
		Status: "firing",
		CommonLabels: map[string]interface{}{
			"alertname": "HighCPU",
			"env":       "production",
			"severity":  "critical",
		},
		ExternalURL: "https://alertmanager.example.test",
		Alerts: []map[string]interface{}{
			{
				"status": "firing",
				"labels": map[string]interface{}{
					"namespace": "monitoring",
					"pod":       "api-1",
					"ignored":   42,
				},
				"annotations":  map[string]interface{}{"summary": "CPU 過高", "ignored": true},
				"startsAt":     "2026-10-21T07:00:00Z",
				"endsAt":       "2026-10-21T08:00:00Z",
				"generatorURL": "https://example.test/graph",
			},
			{"status": "resolved"},
			{"status": 42},
		},
	}

	got, err := manager.convertAlertManagerData(data)
	if err != nil {
		t.Fatalf("轉換 AlertManager data 失敗：%v", err)
	}
	if got.Status != "firing" || got.AlertName != "HighCPU" || got.Env != "production" || got.Severity != "critical" {
		t.Fatalf("基本欄位不符預期：%+v", got)
	}
	if got.Namespace != "monitoring" || got.TotalAlerts != 3 || got.FiringCount != 1 || got.ResolvedCount != 1 {
		t.Fatalf("fallback 或計數不符預期：%+v", got)
	}
	if got.ExternalURL != data.ExternalURL {
		t.Fatalf("ExternalURL 不符預期：%q", got.ExternalURL)
	}
	wantFirst := template.AlertData{
		Status:       "firing",
		Labels:       map[string]string{"namespace": "monitoring", "pod": "api-1"},
		Annotations:  map[string]string{"summary": "CPU 過高"},
		StartsAt:     "2026-10-21T07:00:00Z",
		EndsAt:       "2026-10-21T08:00:00Z",
		GeneratorURL: "https://example.test/graph",
	}
	if !reflect.DeepEqual(got.Alerts[0], wantFirst) {
		t.Fatalf("第一筆警報轉換不符預期：got=%+v want=%+v", got.Alerts[0], wantFirst)
	}
}

func TestNotificationManagerProviderAccessorsReturnExpectedValues(t *testing.T) {
	wantStatus := &types.ProviderStatus{Name: "test", Enabled: true}
	provider := &fakeProvider{name: "test", enabled: true, status: wantStatus}
	manager := &NotificationManager{providers: map[string]types.NotificationProvider{"test": provider}}

	gotProvider, exists := manager.GetProvider("test")
	if !exists || gotProvider != provider {
		t.Fatalf("GetProvider 不符預期：provider=%v exists=%t", gotProvider, exists)
	}
	if _, exists := manager.GetProvider("missing"); exists {
		t.Fatal("不存在的 provider 不應回傳 exists=true")
	}

	providers := manager.GetAllProviders()
	delete(providers, "test")
	if _, exists := manager.GetProvider("test"); !exists {
		t.Fatal("修改 GetAllProviders 回傳 map 不應影響 manager")
	}

	status, err := manager.GetProviderStatus("test")
	if err != nil || status != wantStatus {
		t.Fatalf("GetProviderStatus 不符預期：status=%+v err=%v", status, err)
	}
	if _, err := manager.GetProviderStatus("missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("不存在 provider 的錯誤不符預期：%v", err)
	}
}
