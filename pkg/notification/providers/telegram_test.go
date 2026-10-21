package providers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"alert-webhooks/pkg/notification/types"
)

type fakeTelegramService struct {
	levels   []int
	messages []string
	sendErr  error
	botInfo  interface{}
	botErr   error
}

func (f *fakeTelegramService) SendMessage(_ context.Context, level int, message string) error {
	f.levels = append(f.levels, level)
	f.messages = append(f.messages, message)
	return f.sendErr
}

func (f *fakeTelegramService) GetBotInfo() (interface{}, error) {
	return f.botInfo, f.botErr
}

func TestNewTelegramProviderValidatesDependencies(t *testing.T) {
	engine := &fakeTemplateEngine{languages: []string{"tw"}}
	service := &fakeTelegramService{}

	tests := []struct {
		name    string
		service TelegramService
		engine  types.TemplateEngine
		wantErr string
	}{
		{name: "missing service", engine: engine, wantErr: "telegram service not available"},
		{name: "missing template engine", service: service, wantErr: "template engine not available"},
		{name: "valid dependencies", service: service, engine: engine},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewTelegramProvider(tt.service, tt.engine)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("錯誤不符預期：got=%v want=%q", err, tt.wantErr)
				}
				return
			}
			if err != nil || provider.GetName() != "telegram" {
				t.Fatalf("建立 provider 失敗：provider=%v err=%v", provider, err)
			}
		})
	}
}

func TestTelegramProviderParsesLevel(t *testing.T) {
	provider := &TelegramProvider{}
	tests := []struct {
		name    string
		req     types.NotificationRequest
		want    int
		wantErr string
	}{
		{name: "prefixed level", req: types.NotificationRequest{Level: "L3"}, want: 3},
		{name: "numeric level", req: types.NotificationRequest{Level: "6"}, want: 6},
		{name: "chat ID fallback", req: types.NotificationRequest{ChatID: "L2"}, want: 2},
		{name: "invalid text", req: types.NotificationRequest{Level: "critical"}, wantErr: "invalid level format"},
		{name: "out of range", req: types.NotificationRequest{Level: "L7"}, wantErr: "level out of range"},
		{name: "invalid chat ID", req: types.NotificationRequest{ChatID: "room"}, wantErr: "no valid level"},
		{name: "empty request", req: types.NotificationRequest{}, wantErr: "no valid level"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.parseLevelFromRequest(&tt.req)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("level 解析不符預期：got=%d err=%v want=%d", got, err, tt.want)
			}
		})
	}
}

func TestTelegramProviderSendMessageUpdatesStatistics(t *testing.T) {
	loadProviderConfig(t, "app:\n  mode: production\nlog:\n  level: info\n")
	engine := &fakeTemplateEngine{languages: []string{"tw"}}

	tests := []struct {
		name       string
		req        types.NotificationRequest
		sendErr    error
		wantErr    string
		wantSent   int64
		wantErrors int64
		wantLevels []int
	}{
		{
			name:       "success",
			req:        types.NotificationRequest{Level: "L1", Message: "test"},
			wantSent:   1,
			wantLevels: []int{1},
		},
		{
			name:       "service error",
			req:        types.NotificationRequest{Level: "2", Message: "test"},
			sendErr:    errors.New("send failed"),
			wantErr:    "send failed",
			wantErrors: 1,
			wantLevels: []int{2},
		},
		{
			name:       "invalid level",
			req:        types.NotificationRequest{Level: "L9", Message: "test"},
			wantErr:    "invalid level format",
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &fakeTelegramService{sendErr: tt.sendErr}
			providerValue, err := NewTelegramProvider(service, engine)
			if err != nil {
				t.Fatalf("建立 provider 失敗：%v", err)
			}
			provider := providerValue.(*TelegramProvider)

			err = provider.SendMessage(context.Background(), &tt.req)
			if tt.wantErr == "" && err != nil {
				t.Fatalf("發送訊息失敗：%v", err)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)) {
				t.Fatalf("錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
			if !reflect.DeepEqual(service.levels, tt.wantLevels) {
				t.Fatalf("送出 level 不符預期：got=%v want=%v", service.levels, tt.wantLevels)
			}
			if provider.stats.MessagesSent != tt.wantSent || provider.stats.MessagesError != tt.wantErrors {
				t.Fatalf("統計不符預期：%+v", provider.stats)
			}
			if tt.wantSent == 1 && provider.stats.LastMessageTime == 0 {
				t.Fatal("成功發送後應記錄 LastMessageTime")
			}
		})
	}
}

func TestTelegramProviderConfigurationCapabilitiesAndStatus(t *testing.T) {
	loadProviderConfig(t, `
telegram:
  enable: true
  token: test-token
  chat_ids0: room-0
  chat_ids3: room-3
`)
	service := &fakeTelegramService{botInfo: map[string]interface{}{"username": "bot"}}
	providerValue, err := NewTelegramProvider(service, &fakeTemplateEngine{languages: []string{"tw", "eng"}})
	if err != nil {
		t.Fatalf("建立 provider 失敗：%v", err)
	}
	provider := providerValue.(*TelegramProvider)

	if !provider.IsEnabled() {
		t.Fatal("設定 enable=true 時 provider 應啟用")
	}
	if err := provider.ValidateConfig(); err != nil {
		t.Fatalf("有效設定不應驗證失敗：%v", err)
	}
	capabilities := provider.GetCapabilities()
	if !capabilities.SupportsLevels || capabilities.SupportsChannels || capabilities.MaxMessageLength != 4096 {
		t.Fatalf("能力描述不符預期：%+v", capabilities)
	}
	if !reflect.DeepEqual(capabilities.SupportedLanguages, []string{"tw", "eng"}) {
		t.Fatalf("支援語言不符預期：%v", capabilities.SupportedLanguages)
	}
	status := provider.GetStatus()
	if !status.Enabled || !status.Connected || status.LastError != "" {
		t.Fatalf("狀態不符預期：%+v", status)
	}
	if !reflect.DeepEqual(status.Channels, map[string]string{"L0": "room-0", "L3": "room-3"}) {
		t.Fatalf("ChatID mapping 不符預期：%v", status.Channels)
	}
	if err := provider.TestConnection(); err != nil {
		t.Fatalf("連線測試不應失敗：%v", err)
	}

	service.botErr = errors.New("offline")
	status = provider.GetStatus()
	if status.Connected || status.LastError != "offline" {
		t.Fatalf("失敗狀態不符預期：%+v", status)
	}
	if err := provider.TestConnection(); err == nil || !strings.Contains(err.Error(), "failed to connect") {
		t.Fatalf("連線錯誤不符預期：%v", err)
	}
}

func TestTelegramProviderRejectsIncompleteConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "missing token", config: "telegram:\n  chat_ids0: room\n", wantErr: "token is required"},
		{name: "missing chat ID", config: "telegram:\n  token: test-token\n", wantErr: "chat ID must be configured"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadProviderConfig(t, tt.config)
			provider := &TelegramProvider{}
			if err := provider.ValidateConfig(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("設定錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
		})
	}

	provider := &TelegramProvider{}
	if err := provider.TestConnection(); err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("缺少 service 應失敗：%v", err)
	}
}
