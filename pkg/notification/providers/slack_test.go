package providers

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"alert-webhooks/pkg/notification/types"
)

type fakeSlackService struct {
	channels      []string
	levels        []string
	messages      []string
	sendErr       error
	connectionErr error
}

func (f *fakeSlackService) SendMessage(_ context.Context, channel, message string) error {
	f.channels = append(f.channels, channel)
	f.messages = append(f.messages, message)
	return f.sendErr
}

func (f *fakeSlackService) SendMessageToLevel(_ context.Context, level, message string) error {
	f.levels = append(f.levels, level)
	f.messages = append(f.messages, message)
	return f.sendErr
}

func (f *fakeSlackService) TestConnection() error {
	return f.connectionErr
}

func TestNewSlackProviderValidatesDependencies(t *testing.T) {
	engine := &fakeTemplateEngine{languages: []string{"tw"}}
	service := &fakeSlackService{}

	tests := []struct {
		name    string
		service SlackService
		engine  types.TemplateEngine
		wantErr string
	}{
		{name: "missing service", engine: engine, wantErr: "slack service not available"},
		{name: "missing template engine", service: service, wantErr: "template engine not available"},
		{name: "valid dependencies", service: service, engine: engine},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewSlackProvider(tt.service, tt.engine)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("錯誤不符預期：got=%v want=%q", err, tt.wantErr)
				}
				return
			}
			if err != nil || provider.GetName() != "slack" {
				t.Fatalf("建立 provider 失敗：provider=%v err=%v", provider, err)
			}
		})
	}
}

func TestSlackProviderRoutesMessages(t *testing.T) {
	tests := []struct {
		name         string
		config       string
		req          types.NotificationRequest
		wantChannels []string
		wantLevels   []string
	}{
		{
			name:         "explicit channel adds prefix",
			config:       "slack:\n  channel: '#default'\n",
			req:          types.NotificationRequest{Channel: "alerts", Message: "test"},
			wantChannels: []string{"#alerts"},
		},
		{
			name:         "direct message keeps prefix",
			config:       "slack:\n  channel: '#default'\n",
			req:          types.NotificationRequest{Channel: "@operator", Message: "test"},
			wantChannels: []string{"@operator"},
		},
		{
			name:       "level routing",
			config:     "slack:\n  channels:\n    l2: '#warning'\n",
			req:        types.NotificationRequest{Level: "L2", Message: "test"},
			wantLevels: []string{"L2"},
		},
		{
			name:         "configured default channel",
			config:       "slack:\n  channel: '#configured'\n",
			req:          types.NotificationRequest{Message: "test"},
			wantChannels: []string{"#configured"},
		},
		{
			name:         "built-in default channel",
			config:       "slack: {}\n",
			req:          types.NotificationRequest{Message: "test"},
			wantChannels: []string{"#alerts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadProviderConfig(t, tt.config)
			service := &fakeSlackService{}
			providerValue, err := NewSlackProvider(service, &fakeTemplateEngine{})
			if err != nil {
				t.Fatalf("建立 provider 失敗：%v", err)
			}
			provider := providerValue.(*SlackProvider)

			if err := provider.SendMessage(context.Background(), &tt.req); err != nil {
				t.Fatalf("發送訊息失敗：%v", err)
			}
			if !reflect.DeepEqual(service.channels, tt.wantChannels) || !reflect.DeepEqual(service.levels, tt.wantLevels) {
				t.Fatalf("routing 不符預期：channels=%v levels=%v", service.channels, service.levels)
			}
			if provider.stats.MessagesSent != 1 || provider.stats.MessagesError != 0 || provider.stats.LastMessageTime == 0 {
				t.Fatalf("成功統計不符預期：%+v", provider.stats)
			}
		})
	}
}

func TestSlackProviderRecordsServiceError(t *testing.T) {
	loadProviderConfig(t, "slack:\n  channel: '#alerts'\n")
	wantErr := errors.New("send failed")
	service := &fakeSlackService{sendErr: wantErr}
	providerValue, err := NewSlackProvider(service, &fakeTemplateEngine{})
	if err != nil {
		t.Fatalf("建立 provider 失敗：%v", err)
	}
	provider := providerValue.(*SlackProvider)

	err = provider.SendMessage(context.Background(), &types.NotificationRequest{Message: "test"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("應保留 service error：%v", err)
	}
	if provider.stats.MessagesSent != 0 || provider.stats.MessagesError != 1 || provider.stats.LastMessageTime != 0 {
		t.Fatalf("失敗統計不符預期：%+v", provider.stats)
	}
}

func TestSlackProviderLevelChannelFallbacks(t *testing.T) {
	tests := []struct {
		name   string
		config string
		level  string
		want   string
	}{
		{
			name:   "normalized key",
			config: "slack:\n  channels:\n    l2: '#normalized'\n",
			level:  "2",
			want:   "#normalized",
		},
		{
			name:   "original key",
			config: "slack:\n  channels:\n    Critical: '#critical'\n",
			level:  "Critical",
			want:   "#critical",
		},
		{
			name:   "uppercase key",
			config: "slack:\n  channels:\n    L3: '#upper'\n",
			level:  "l3",
			want:   "#upper",
		},
		{
			name:   "trimmed key",
			config: "slack:\n  channels:\n    '4': '#numeric'\n",
			level:  "L4",
			want:   "#numeric",
		},
		{
			name:   "configured default",
			config: "slack:\n  channel: '#default'\n",
			level:  "L5",
			want:   "#default",
		},
		{name: "built-in default", config: "slack: {}\n", level: "L6", want: "#alerts"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadProviderConfig(t, tt.config)
			provider := &SlackProvider{}
			if got := provider.getLevelChannel(tt.level); got != tt.want {
				t.Fatalf("channel 不符預期：got=%q want=%q", got, tt.want)
			}
		})
	}
}

func TestSlackProviderConfigurationCapabilitiesAndStatus(t *testing.T) {
	loadProviderConfig(t, `
slack:
  enable: true
  token: test-token
  channel: '#default'
  channels:
    l0: '#critical'
`)
	service := &fakeSlackService{}
	providerValue, err := NewSlackProvider(service, &fakeTemplateEngine{languages: []string{"tw", "eng"}})
	if err != nil {
		t.Fatalf("建立 provider 失敗：%v", err)
	}
	provider := providerValue.(*SlackProvider)

	if !provider.IsEnabled() {
		t.Fatal("設定 enable=true 時 provider 應啟用")
	}
	if err := provider.ValidateConfig(); err != nil {
		t.Fatalf("有效設定不應驗證失敗：%v", err)
	}
	capabilities := provider.GetCapabilities()
	if !capabilities.SupportsLevels || !capabilities.SupportsChannels || !capabilities.SupportsAttachments || capabilities.MaxMessageLength != 40000 {
		t.Fatalf("能力描述不符預期：%+v", capabilities)
	}
	if !reflect.DeepEqual(capabilities.SupportedLanguages, []string{"tw", "eng"}) {
		t.Fatalf("支援語言不符預期：%v", capabilities.SupportedLanguages)
	}
	status := provider.GetStatus()
	if !status.Enabled || !status.Connected || status.LastError != "" {
		t.Fatalf("狀態不符預期：%+v", status)
	}
	if !reflect.DeepEqual(status.Channels, map[string]string{"default": "#default", "l0": "#critical"}) {
		t.Fatalf("channel mapping 不符預期：%v", status.Channels)
	}
	if err := provider.TestConnection(); err != nil {
		t.Fatalf("連線測試不應失敗：%v", err)
	}

	service.connectionErr = errors.New("offline")
	status = provider.GetStatus()
	if status.Connected || status.LastError != "offline" {
		t.Fatalf("失敗狀態不符預期：%+v", status)
	}
	if err := provider.TestConnection(); !errors.Is(err, service.connectionErr) {
		t.Fatalf("應保留連線錯誤：%v", err)
	}
}

func TestSlackProviderRejectsIncompleteConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  string
		wantErr string
	}{
		{name: "missing token", config: "slack:\n  channel: '#alerts'\n", wantErr: "token is required"},
		{name: "missing channel", config: "slack:\n  token: test-token\n", wantErr: "channel must be configured"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loadProviderConfig(t, tt.config)
			provider := &SlackProvider{}
			if err := provider.ValidateConfig(); err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("設定錯誤不符預期：got=%v want 包含 %q", err, tt.wantErr)
			}
		})
	}

	provider := &SlackProvider{}
	if err := provider.TestConnection(); err == nil || !strings.Contains(err.Error(), "not available") {
		t.Fatalf("缺少 service 應失敗：%v", err)
	}
}
