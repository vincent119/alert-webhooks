package providers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"alert-webhooks/pkg/notification/types"
)

type fakeDiscordService struct {
	levels        []string
	channels      []string
	messages      []string
	sendErr       error
	connectionErr error
}

func (f *fakeDiscordService) SendMessage(_ context.Context, level, message string) error {
	f.levels = append(f.levels, level)
	f.messages = append(f.messages, message)
	return f.sendErr
}

func (f *fakeDiscordService) SendMessageToChannel(channelID, message string) error {
	f.channels = append(f.channels, channelID)
	f.messages = append(f.messages, message)
	return f.sendErr
}

func (f *fakeDiscordService) SendMessageToLevel(ctx context.Context, level, message string) error {
	return f.SendMessage(ctx, level, message)
}

func (f *fakeDiscordService) TestConnection() error {
	return f.connectionErr
}

func (f *fakeDiscordService) ValidateChannel(string) error {
	return nil
}

func (f *fakeDiscordService) GetBotInfo() (interface{}, error) {
	return nil, f.connectionErr
}

func TestDiscordProviderMetadataAndConnection(t *testing.T) {
	service := &fakeDiscordService{}
	provider := NewDiscordProvider(service, &fakeTemplateEngine{languages: []string{"tw", "eng"}})

	if provider.GetName() != "discord" || !provider.IsEnabled() {
		t.Fatalf("provider metadata 不符預期：name=%q enabled=%t", provider.GetName(), provider.IsEnabled())
	}
	capabilities := provider.GetCapabilities()
	if !capabilities.SupportsLevels || !capabilities.SupportsChannels || capabilities.SupportsAttachments || capabilities.MaxMessageLength != 2000 {
		t.Fatalf("能力描述不符預期：%+v", capabilities)
	}
	if !reflect.DeepEqual(capabilities.SupportedLanguages, []string{"tw", "eng"}) {
		t.Fatalf("支援語言不符預期：%v", capabilities.SupportedLanguages)
	}
	if err := provider.ValidateConfig(); err != nil {
		t.Fatalf("有效連線不應驗證失敗：%v", err)
	}
	if err := provider.TestConnection(); err != nil {
		t.Fatalf("連線測試不應失敗：%v", err)
	}
	status := provider.GetStatus()
	if !status.Enabled || !status.Connected || status.LastError != "" {
		t.Fatalf("狀態不符預期：%+v", status)
	}

	service.connectionErr = errors.New("offline")
	if err := provider.ValidateConfig(); !errors.Is(err, service.connectionErr) {
		t.Fatalf("應保留驗證錯誤：%v", err)
	}
	if err := provider.TestConnection(); !errors.Is(err, service.connectionErr) {
		t.Fatalf("應保留連線錯誤：%v", err)
	}
	status = provider.GetStatus()
	if status.Connected || status.LastError != "offline" {
		t.Fatalf("失敗狀態不符預期：%+v", status)
	}
}

func TestDiscordProviderRoutesDirectMessages(t *testing.T) {
	tests := []struct {
		name         string
		req          types.NotificationRequest
		sendErr      error
		wantErr      string
		wantChannels []string
		wantLevels   []string
	}{
		{
			name:         "channel routing",
			req:          types.NotificationRequest{Channel: "channel-1", Message: "test"},
			wantChannels: []string{"channel-1"},
		},
		{
			name:       "level routing",
			req:        types.NotificationRequest{Level: "L2", Message: "test"},
			wantLevels: []string{"L2"},
		},
		{
			name:    "missing destination",
			req:     types.NotificationRequest{Message: "test"},
			wantErr: "channel or level",
		},
		{
			name:       "service error",
			req:        types.NotificationRequest{Level: "L1", Message: "test"},
			sendErr:    errors.New("send failed"),
			wantErr:    "send failed",
			wantLevels: []string{"L1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeDiscordService{sendErr: tt.sendErr}
			provider := NewDiscordProvider(service, &fakeTemplateEngine{})
			err := provider.SendMessage(context.Background(), &tt.req)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("發送訊息失敗：%v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(service.channels, tt.wantChannels) || !reflect.DeepEqual(service.levels, tt.wantLevels) {
				t.Fatalf("routing 不符預期：channels=%v levels=%v", service.channels, service.levels)
			}
		})
	}
}

func TestDiscordProviderFormatsAndRoutesAlertData(t *testing.T) {
	alertData := &types.AlertManagerData{
		Status:      "firing",
		ExternalURL: "https://alertmanager.example.test",
		Alerts: []map[string]interface{}{
			{
				"status": "firing",
				"labels": map[string]interface{}{
					"alertname": "HighCPU",
					"env":       "production",
					"severity":  "critical",
					"namespace": "monitoring",
				},
				"annotations": map[string]interface{}{"summary": "CPU 過高"},
			},
			{
				"status":      "resolved",
				"annotations": map[string]interface{}{"summary": "記憶體恢復"},
			},
		},
	}

	tests := []struct {
		name         string
		req          types.NotificationRequest
		sendErr      error
		wantErr      string
		wantChannels []string
		wantLevels   []string
	}{
		{
			name:         "channel routing",
			req:          types.NotificationRequest{Channel: "channel-1", AlertData: alertData},
			wantChannels: []string{"channel-1"},
		},
		{
			name:       "level routing",
			req:        types.NotificationRequest{Level: "L0", AlertData: alertData},
			wantLevels: []string{"L0"},
		},
		{
			name:    "missing destination",
			req:     types.NotificationRequest{AlertData: alertData},
			wantErr: "channel or level",
		},
		{
			name:         "service error",
			req:          types.NotificationRequest{Channel: "channel-2", AlertData: alertData},
			sendErr:      errors.New("send failed"),
			wantErr:      "send failed",
			wantChannels: []string{"channel-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeDiscordService{sendErr: tt.sendErr}
			provider := NewDiscordProvider(service, &fakeTemplateEngine{})
			err := provider.SendMessage(context.Background(), &tt.req)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("發送訊息失敗：%v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(service.channels, tt.wantChannels) || !reflect.DeepEqual(service.levels, tt.wantLevels) {
				t.Fatalf("routing 不符預期：channels=%v levels=%v", service.channels, service.levels)
			}
			if len(service.messages) > 0 {
				for _, fragment := range []string{"HighCPU", "production", "critical", "monitoring", "CPU 過高", "記憶體恢復", "View Details"} {
					if !strings.Contains(service.messages[0], fragment) {
						t.Fatalf("格式化訊息缺少 %q：%s", fragment, service.messages[0])
					}
				}
			}
		})
	}
}

func TestDiscordProviderRejectsEmptyContent(t *testing.T) {
	provider := NewDiscordProvider(&fakeDiscordService{}, &fakeTemplateEngine{})
	err := provider.SendMessage(context.Background(), &types.NotificationRequest{Level: "L1"})
	if err == nil || !strings.Contains(err.Error(), "message or alert data") {
		t.Fatalf("空內容錯誤不符預期：%v", err)
	}
}

func TestDiscordProviderFormatsEmptyAndNonStringValues(t *testing.T) {
	provider := &DiscordProvider{}
	if got := provider.formatBasicMessage(&types.AlertManagerData{}); got != "🚨 **Alert Notification**\n\n" {
		t.Fatalf("空警報格式不符預期：%q", got)
	}

	alertData := &types.AlertManagerData{
		Status: "resolved",
		Alerts: []map[string]interface{}{{
			"status": "resolved",
			"labels": map[string]interface{}{
				"alertname": 42,
				"env":       nil,
			},
		}},
	}
	got := provider.formatBasicMessage(alertData)
	if !strings.Contains(got, "**Alert Name:** 42") || strings.Contains(got, "**Environment:**") {
		t.Fatalf("非字串與 nil 欄位格式不符預期：%s", got)
	}
}
