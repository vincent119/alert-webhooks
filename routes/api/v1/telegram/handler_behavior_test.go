package telegram

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"alert-webhooks/config"
	"alert-webhooks/pkg/service"
	"alert-webhooks/pkg/template"

	"github.com/gin-gonic/gin"
	"github.com/go-telegram/bot/models"
)

func TestTelegramSendMessageValidationAndDirectMessage(t *testing.T) {
	tests := []struct {
		name       string
		service    *telegramServiceForTest
		chatID     string
		body       string
		wantStatus int
	}{
		{name: "invalid chat ID", service: &telegramServiceForTest{}, chatID: "L9", body: `{"message":"hello"}`, wantStatus: http.StatusBadRequest},
		{name: "invalid chat ID text", service: &telegramServiceForTest{}, chatID: "invalid", body: `{"message":"hello"}`, wantStatus: http.StatusBadRequest},
		{name: "invalid JSON", service: &telegramServiceForTest{}, chatID: "L0", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "empty payload", service: &telegramServiceForTest{}, chatID: "0", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "direct success", service: &telegramServiceForTest{}, chatID: "L2", body: `{"message":"hello"}`, wantStatus: http.StatusOK},
		{name: "direct failure", service: &telegramServiceForTest{sendErrors: []error{errors.New("send failed")}}, chatID: "2", body: `{"message":"hello"}`, wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := invokeTelegramHandlerReplay(t, http.MethodPost, tt.body, map[string]string{"chatid": tt.chatID}, newTelegramHandlerForTest(tt.service, nil).SendMessage)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantStatus == http.StatusOK && (tt.service.lastLevel != 2 || tt.service.messages[0] != "hello") {
				t.Fatalf("level = %d, messages = %#v", tt.service.lastLevel, tt.service.messages)
			}
		})
	}
}

func TestTelegramSendAlertManagerMessages(t *testing.T) {
	setTelegramHandlerConfig(t, config.TelegramConf{TemplateLanguage: "tw", TemplateMode: "full"})
	body := `{
		"alertmanager_data": {
			"status":"firing",
			"commonLabels":{"namespace":"payments"},
			"externalURL":"https://alerts.example",
			"alerts":[
				{"status":"firing","labels":{"alertname":"CPU.High","env":"prod","severity":"critical","pod":"api-0"},"annotations":{"summary":"high_cpu","description":"description"},"startsAt":"2026-10-21T00:00:00Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"https://generator"},
				{"status":"resolved","labels":{"alertname":"CPU.High","env":"prod","severity":"critical","pod":"api-1"},"annotations":{"summary":"resolved","description":"done"},"startsAt":"2026-10-21T00:00:00Z","endsAt":"2026-10-21T01:00:00Z","generatorURL":"https://generator"}
			]
		}
	}`

	t.Run("fallback 分離傳送成功", func(t *testing.T) {
		fake := &telegramServiceForTest{}
		response := invokeTelegramHandlerReplay(t, http.MethodPost, body, map[string]string{"chatid": "L1"}, newTelegramHandlerForTest(fake, nil).SendMessage)
		if response.Code != http.StatusOK || len(fake.messages) != 2 {
			t.Fatalf("status = %d, messages = %#v, body = %s", response.Code, fake.messages, response.Body.String())
		}
		if !strings.Contains(fake.messages[0], "警報通知") || !strings.Contains(fake.messages[1], "警報已解決") {
			t.Fatalf("messages = %#v", fake.messages)
		}
	})

	t.Run("firing 傳送失敗", func(t *testing.T) {
		fake := &telegramServiceForTest{sendErrors: []error{errors.New("firing failed")}}
		response := invokeTelegramHandlerReplay(t, http.MethodPost, body, map[string]string{"chatid": "1"}, newTelegramHandlerForTest(fake, nil).SendMessage)
		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
	})

	t.Run("resolved 傳送失敗", func(t *testing.T) {
		fake := &telegramServiceForTest{sendErrors: []error{nil, errors.New("resolved failed")}}
		response := invokeTelegramHandlerReplay(t, http.MethodPost, body, map[string]string{"chatid": "1"}, newTelegramHandlerForTest(fake, nil).SendMessage)
		if response.Code != http.StatusInternalServerError || len(fake.messages) != 2 {
			t.Fatalf("status = %d, messages = %#v", response.Code, fake.messages)
		}
	})

	t.Run("template engine render fallback", func(t *testing.T) {
		fake := &telegramServiceForTest{}
		response := invokeTelegramHandlerReplay(t, http.MethodPost, body, map[string]string{"chatid": "1"}, newTelegramHandlerForTest(fake, template.NewTemplateEngine()).SendMessage)
		if response.Code != http.StatusOK {
			t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
		}
	})
}

func TestTelegramDirectAlertManagerFallbackAndDebugPaths(t *testing.T) {
	originalApp := config.App
	originalLog := config.Log
	config.App.Mode = "development"
	config.Log.Level = "debug"
	t.Cleanup(func() {
		config.App = originalApp
		config.Log = originalLog
	})

	body := `{"receiver":"alerts","status":"firing","alerts":[{"status":"firing","labels":{"alertname":"CPU"},"annotations":{"summary":"high"}}]}`
	fake := &telegramServiceForTest{}
	response := invokeTelegramHandler(t, http.MethodPost, body, map[string]string{"chatid": "L0"}, newTelegramHandlerForTest(fake, nil).SendMessage)
	if response.Code != http.StatusOK || len(fake.messages) != 1 {
		t.Fatalf("status = %d, messages = %#v, body = %s", response.Code, fake.messages, response.Body.String())
	}

	invalid := invokeTelegramHandlerReplay(t, http.MethodPost, `{`, map[string]string{"chatid": "L0"}, newTelegramHandlerForTest(&telegramServiceForTest{}, nil).SendMessage)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid status = %d, body = %s", invalid.Code, invalid.Body.String())
	}
}

func TestTelegramTemplateRenderSuccessAndFailure(t *testing.T) {
	engine := template.NewTemplateEngine()
	if err := engine.LoadTemplates("../../../../templates/alerts"); err != nil {
		t.Fatalf("LoadTemplates() error = %v", err)
	}
	body := `{"template_language":"eng","alertmanager_data":{"status":"firing","alerts":[{"status":"firing","labels":{"alertname":"CPU"},"annotations":{"summary":"high"}}]}}`

	success := &telegramServiceForTest{}
	response := invokeTelegramHandlerReplay(t, http.MethodPost, body, map[string]string{"chatid": "L0"}, newTelegramHandlerForTest(success, engine).SendMessage)
	if response.Code != http.StatusOK || len(success.messages) != 1 {
		t.Fatalf("status = %d, messages = %#v, body = %s", response.Code, success.messages, response.Body.String())
	}

	failure := &telegramServiceForTest{sendErrors: []error{errors.New("send failed")}}
	response = invokeTelegramHandlerReplay(t, http.MethodPost, body, map[string]string{"chatid": "L0"}, newTelegramHandlerForTest(failure, engine).SendMessage)
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestTelegramGetBotInfo(t *testing.T) {
	tests := []struct {
		name       string
		service    *telegramServiceForTest
		wantStatus int
		want       string
	}{
		{name: "provider error", service: &telegramServiceForTest{botInfoErr: errors.New("info failed")}, wantStatus: http.StatusInternalServerError, want: "Failed to get bot info"},
		{name: "invalid type", service: &telegramServiceForTest{botInfo: map[string]interface{}{}}, wantStatus: http.StatusInternalServerError, want: "Invalid bot info type"},
		{name: "success", service: &telegramServiceForTest{botInfo: &models.User{ID: 42, Username: "alert_bot", FirstName: "Alert", CanJoinGroups: true}}, wantStatus: http.StatusOK, want: "alert_bot"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := invokeTelegramHandler(t, http.MethodGet, "", nil, newTelegramHandlerForTest(tt.service, nil).GetBotInfo)
			if response.Code != tt.wantStatus || !strings.Contains(response.Body.String(), tt.want) {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestTelegramConversionAndFormattingHelpers(t *testing.T) {
	alert := Alert{
		Status:       "firing",
		Labels:       map[string]string{"alertname": "CPU.High", "env": "prod", "severity": "critical", "pod": "api-0"},
		Annotations:  map[string]string{"summary": "high_cpu", "description": "description"},
		StartsAt:     "2026-10-21T00:00:00Z",
		EndsAt:       "2026-10-21T01:00:00Z",
		GeneratorURL: "https://generator",
		Fingerprint:  "fingerprint",
	}
	converted := convertAlertSliceToMap([]Alert{alert})
	if len(converted) != 1 || converted[0]["fingerprint"] != "fingerprint" {
		t.Fatalf("converted alerts = %#v", converted)
	}
	if convertStringMapToInterface(nil) != nil {
		t.Fatal("nil map must remain nil")
	}
	if convertedMap := convertStringMapToInterface(map[string]string{"key": "value"}); convertedMap["key"] != "value" {
		t.Fatalf("converted map = %#v", convertedMap)
	}

	webhook := &AlertManagerWebhook{
		Status:       "firing",
		Alerts:       []Alert{alert, {Status: "resolved", Labels: alert.Labels, Annotations: alert.Annotations, StartsAt: alert.StartsAt, EndsAt: alert.EndsAt, GeneratorURL: alert.GeneratorURL}},
		CommonLabels: map[string]string{"namespace": "payments"},
		ExternalURL:  "https://alerts.example",
	}
	for _, language := range []string{"tw", "eng"} {
		handler := newTelegramHandlerForTest(&telegramServiceForTest{}, nil)
		message := handler.generateAlertManagerMessage(webhook, language)
		for _, expected := range []string{"CPU\\.High", "payments", "high\\_cpu", "api\\-0"} {
			if !strings.Contains(message, expected) {
				t.Fatalf("%s message %q does not contain %q", language, message, expected)
			}
		}
	}

	handler := newTelegramHandlerForTest(&telegramServiceForTest{}, nil)
	if got := handler.formatTime(""); got != "未設定" {
		t.Fatalf("empty time = %q", got)
	}
	if got := handler.formatTime("invalid"); got != "invalid" {
		t.Fatalf("invalid time = %q", got)
	}
	if got := handler.formatTime("2026-10-21T00:00:00Z"); got != "2026-10-21 00:00:00" {
		t.Fatalf("formatted time = %q", got)
	}
}

func TestTelegramFormatOptionsModes(t *testing.T) {
	setTelegramHandlerConfig(t, config.TelegramConf{TemplateMode: "minimal"})
	minimal := newTelegramHandlerForTest(&telegramServiceForTest{}, nil).getFormatOptionsForTelegram()
	if !minimal.CompactMode.Enabled || minimal.ShowExternalURL.Enabled {
		t.Fatalf("minimal options = %#v", minimal)
	}

	setTelegramHandlerConfig(t, config.TelegramConf{TemplateMode: "full"})
	full := newTelegramHandlerForTest(&telegramServiceForTest{}, nil).getFormatOptionsForTelegram()
	if full.CompactMode.Enabled || !full.ShowExternalURL.Enabled {
		t.Fatalf("full options = %#v", full)
	}

	setTelegramHandlerConfig(t, config.TelegramConf{TemplateMode: "minimal"})
	engine := template.NewTemplateEngine()
	minimalWithEngine := newTelegramHandlerForTest(&telegramServiceForTest{}, engine).getFormatOptionsForTelegram()
	if !minimalWithEngine.CompactMode.Enabled {
		t.Fatalf("minimal engine options = %#v", minimalWithEngine)
	}
}

func TestTelegramRoutesRejectUnavailableService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group(""), service.NewStaticAccessor(nil))
	for _, request := range []*http.Request{
		httptest.NewRequest(http.MethodPost, "/telegram/chatid_L0", strings.NewReader(`{"message":"hello"}`)),
		httptest.NewRequest(http.MethodGet, "/telegram/info", nil),
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusServiceUnavailable {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
		}
	}
}

func TestTelegramDirectMessageCurrentBodyConsumptionBehavior(t *testing.T) {
	fake := &telegramServiceForTest{}
	response := invokeTelegramHandler(t, http.MethodPost, `{"message":"hello"}`, map[string]string{"chatid": "L0"}, newTelegramHandlerForTest(fake, nil).SendMessage)
	if response.Code != http.StatusOK || len(fake.messages) != 0 {
		t.Fatalf("status = %d, messages = %#v, body = %s", response.Code, fake.messages, response.Body.String())
	}
}

func TestTelegramCurrentServiceWithoutLease(t *testing.T) {
	handler := NewHandler(nil)
	response := invokeTelegramHandler(t, http.MethodPost, "", map[string]string{"chatid": "invalid"}, handler.SendMessage)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestTelegramNewHandlerUsesServiceManagerTemplateEngine(t *testing.T) {
	service.GetServiceManager().ReloadTemplateEngine()
	handler := NewHandler(nil)
	if handler.templateEngine == nil || handler.resolveService == nil {
		t.Fatalf("handler = %#v", handler)
	}
}

func newTelegramHandlerForTest(value *telegramServiceForTest, engine *template.TemplateEngine) *Handler {
	return &Handler{
		templateEngine: engine,
		resolveService: func(*gin.Context) telegramService {
			if value == nil {
				return nil
			}
			return value
		},
	}
}

func invokeTelegramHandler(t *testing.T, method, body string, params map[string]string, callback gin.HandlerFunc) *httptest.ResponseRecorder {
	return invokeTelegramHandlerWithBody(t, method, body, params, callback, false)
}

func invokeTelegramHandlerReplay(t *testing.T, method, body string, params map[string]string, callback gin.HandlerFunc) *httptest.ResponseRecorder {
	return invokeTelegramHandlerWithBody(t, method, body, params, callback, true)
}

func invokeTelegramHandlerWithBody(t *testing.T, method, body string, params map[string]string, callback gin.HandlerFunc, replay bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	contextValue, _ := gin.CreateTestContext(response)
	contextValue.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	if replay {
		contextValue.Request.Body = &replayRequestBody{content: []byte(body)}
		contextValue.Request.ContentLength = int64(len(body))
	}
	contextValue.Request.Header.Set("Content-Type", "application/json")
	for key, value := range params {
		contextValue.Params = append(contextValue.Params, gin.Param{Key: key, Value: value})
	}
	callback(contextValue)
	return response
}

type replayRequestBody struct {
	content []byte
	offset  int
}

func (b *replayRequestBody) Read(destination []byte) (int, error) {
	if b.offset == len(b.content) {
		b.offset = 0
		return 0, io.EOF
	}
	written := copy(destination, b.content[b.offset:])
	b.offset += written
	return written, nil
}

func (b *replayRequestBody) Close() error { return nil }

func setTelegramHandlerConfig(t *testing.T, value config.TelegramConf) {
	t.Helper()
	original := config.Telegram
	config.Telegram = value
	t.Cleanup(func() { config.Telegram = original })
}

type telegramServiceForTest struct {
	sendErrors []error
	messages   []string
	lastLevel  int
	botInfo    interface{}
	botInfoErr error
}

func (s *telegramServiceForTest) SendMessage(_ context.Context, level int, message string) error {
	s.lastLevel = level
	s.messages = append(s.messages, message)
	if len(s.sendErrors) == 0 {
		return nil
	}
	err := s.sendErrors[0]
	s.sendErrors = s.sendErrors[1:]
	return err
}

func (s *telegramServiceForTest) GetBotInfo() (interface{}, error) {
	return s.botInfo, s.botInfoErr
}
