package discord

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

func TestDiscordSendMessageToChannel(t *testing.T) {
	tests := []struct {
		name       string
		service    *discordServiceForTest
		body       string
		channel    string
		wantStatus int
		wantCode   string
	}{
		{name: "missing channel", service: &discordServiceForTest{}, body: `{"message":"hello"}`, wantStatus: http.StatusBadRequest},
		{name: "unavailable", body: `{"message":"hello"}`, channel: "alerts", wantStatus: http.StatusServiceUnavailable},
		{name: "invalid JSON", service: &discordServiceForTest{}, body: `{`, channel: "alerts", wantStatus: http.StatusBadRequest, wantCode: "DISCORD_INVALID_REQUEST_FORMAT"},
		{name: "empty payload", service: &discordServiceForTest{}, body: `{}`, channel: "alerts", wantStatus: http.StatusBadRequest},
		{name: "direct success", service: &discordServiceForTest{}, body: `{"message":"hello"}`, channel: "alerts", wantStatus: http.StatusOK},
		{name: "direct send failure", service: &discordServiceForTest{sendChannelErr: errors.New("send failed")}, body: `{"message":"hello"}`, channel: "alerts", wantStatus: http.StatusInternalServerError, wantCode: "DISCORD_MESSAGE_SEND_FAILED"},
		{name: "wrapped AlertManager", service: &discordServiceForTest{}, body: `{"alertmanager_data":{"status":"firing","alerts":[{"status":"firing","labels":{"alertname":"CPU"}}]}}`, channel: "alerts", wantStatus: http.StatusOK},
		{name: "wrapped AlertManager send failure", service: &discordServiceForTest{sendChannelErr: errors.New("send failed")}, body: `{"alertmanager_data":{"status":"firing","alerts":[{"status":"firing"}]}}`, channel: "alerts", wantStatus: http.StatusInternalServerError},
		{name: "direct AlertManager", service: &discordServiceForTest{}, body: `{"status":"resolved","alerts":[{"status":"resolved","labels":{"alertname":"CPU"}}]}`, channel: "alerts", wantStatus: http.StatusOK},
		{name: "direct AlertManager send failure", service: &discordServiceForTest{sendChannelErr: errors.New("send failed")}, body: `{"status":"resolved","alerts":[{"status":"resolved"}]}`, channel: "alerts", wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newDiscordHandlerForTest(tt.service)
			response := invokeDiscordHandler(t, http.MethodPost, "/", tt.body, map[string]string{"channel": tt.channel}, handler.SendMessageToChannel)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantCode != "" && !strings.Contains(response.Body.String(), tt.wantCode) {
				t.Fatalf("body %q does not contain %q", response.Body.String(), tt.wantCode)
			}
		})
	}
}

func TestDiscordSendMessageToLevel(t *testing.T) {
	tests := []struct {
		name       string
		service    *discordServiceForTest
		body       string
		level      string
		wantStatus int
		wantLevel  string
	}{
		{name: "missing level", service: &discordServiceForTest{}, body: `{"message":"hello"}`, wantStatus: http.StatusBadRequest},
		{name: "unavailable", body: `{"message":"hello"}`, level: "0", wantStatus: http.StatusServiceUnavailable},
		{name: "invalid JSON", service: &discordServiceForTest{}, body: `{`, level: "0", wantStatus: http.StatusBadRequest},
		{name: "empty payload", service: &discordServiceForTest{}, body: `{}`, level: "0", wantStatus: http.StatusBadRequest},
		{name: "direct success", service: &discordServiceForTest{}, body: `{"message":"hello"}`, level: "2", wantStatus: http.StatusOK, wantLevel: "L2"},
		{name: "direct failure", service: &discordServiceForTest{sendErr: errors.New("send failed")}, body: `{"message":"hello"}`, level: "2", wantStatus: http.StatusInternalServerError, wantLevel: "L2"},
		{name: "wrapped AlertManager", service: &discordServiceForTest{}, body: `{"alertmanager_data":{"status":"firing","alerts":[{"status":"firing"}]}}`, level: "1", wantStatus: http.StatusOK, wantLevel: "L1"},
		{name: "wrapped AlertManager send failure", service: &discordServiceForTest{sendErr: errors.New("send failed")}, body: `{"alertmanager_data":{"status":"firing","alerts":[{"status":"firing"}]}}`, level: "1", wantStatus: http.StatusInternalServerError, wantLevel: "L1"},
		{name: "direct AlertManager", service: &discordServiceForTest{}, body: `{"status":"resolved","alerts":[{"status":"resolved"}]}`, level: "3", wantStatus: http.StatusOK, wantLevel: "L3"},
		{name: "direct AlertManager send failure", service: &discordServiceForTest{sendErr: errors.New("send failed")}, body: `{"status":"resolved","alerts":[{"status":"resolved"}]}`, level: "3", wantStatus: http.StatusInternalServerError, wantLevel: "L3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newDiscordHandlerForTest(tt.service)
			response := invokeDiscordHandler(t, http.MethodPost, "/", tt.body, map[string]string{"level": tt.level}, handler.SendMessageToLevel)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantLevel != "" && tt.service != nil && tt.service.lastLevel != tt.wantLevel {
				t.Fatalf("level = %q, want %q", tt.service.lastLevel, tt.wantLevel)
			}
		})
	}
}

func TestDiscordGetStatus(t *testing.T) {
	tests := []struct {
		name       string
		service    *discordServiceForTest
		wantStatus int
		wantBody   string
	}{
		{name: "unavailable", wantStatus: http.StatusServiceUnavailable, wantBody: "unavailable"},
		{name: "connection failure", service: &discordServiceForTest{testErr: errors.New("connection failed")}, wantStatus: http.StatusInternalServerError, wantBody: "DISCORD_CONNECTION_TEST_FAILED"},
		{name: "bot info failure", service: &discordServiceForTest{botInfoErr: errors.New("info failed")}, wantStatus: http.StatusOK, wantBody: "partial"},
		{name: "healthy", service: &discordServiceForTest{botInfo: map[string]interface{}{"username": "bot"}}, wantStatus: http.StatusOK, wantBody: "healthy"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := invokeDiscordHandler(t, http.MethodGet, "/", "", nil, newDiscordHandlerForTest(tt.service).GetStatus)
			if response.Code != tt.wantStatus || !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
			}
		})
	}
}

func TestDiscordTestAndValidateChannel(t *testing.T) {
	tests := []struct {
		name       string
		validate   bool
		service    *discordServiceForTest
		channel    string
		wantStatus int
	}{
		{name: "test missing channel", service: &discordServiceForTest{}, wantStatus: http.StatusBadRequest},
		{name: "test unavailable", channel: "alerts", wantStatus: http.StatusServiceUnavailable},
		{name: "test failure", service: &discordServiceForTest{sendChannelErr: errors.New("send failed")}, channel: "alerts", wantStatus: http.StatusInternalServerError},
		{name: "test success", service: &discordServiceForTest{}, channel: "alerts", wantStatus: http.StatusOK},
		{name: "validate missing channel", validate: true, service: &discordServiceForTest{}, wantStatus: http.StatusBadRequest},
		{name: "validate unavailable", validate: true, channel: "alerts", wantStatus: http.StatusServiceUnavailable},
		{name: "validate failure", validate: true, service: &discordServiceForTest{validateErr: errors.New("invalid")}, channel: "alerts", wantStatus: http.StatusBadRequest},
		{name: "validate success", validate: true, service: &discordServiceForTest{}, channel: "alerts", wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newDiscordHandlerForTest(tt.service)
			callback := handler.TestChannel
			if tt.validate {
				callback = handler.ValidateChannel
			}
			response := invokeDiscordHandler(t, http.MethodPost, "/", "", map[string]string{"channel": tt.channel}, callback)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
		})
	}
}

func TestDiscordMessageFormattingHelpers(t *testing.T) {
	handler := newDiscordHandlerForTest(&discordServiceForTest{})
	payload := `{"status":"firing","commonLabels":{"namespace":"payments"},"externalURL":"https://alerts.example","alerts":[` +
		`{"status":"firing","labels":{"alertname":"CPU*High","env":"prod","severity":"critical"},"annotations":{"summary":"load_high"}},` +
		`{"status":"resolved","labels":{},"annotations":{"summary":"resolved"}}]}`
	message, err := handler.generateAlertManagerMessage([]byte(payload))
	if err != nil {
		t.Fatalf("generateAlertManagerMessage() error = %v", err)
	}
	for _, expected := range []string{"CPU\\*High", "payments", "load\\_high", "resolved"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("message %q does not contain %q", message, expected)
		}
	}
	if _, err := handler.generateBuiltInMessage([]byte("invalid")); err == nil {
		t.Fatal("expected invalid AlertManager data error")
	}
	if _, err := handler.generateAlertManagerMessage([]byte("invalid")); err == nil {
		t.Fatal("expected invalid generated AlertManager message error")
	}
	service.GetServiceManager().ReloadTemplateEngine()
	if message, err := handler.generateAlertManagerMessage([]byte(payload)); err != nil || message == "" {
		t.Fatalf("template fallback message = %q, error = %v", message, err)
	}
	if got := handler.escapeText(""); got != "" {
		t.Fatalf("empty escaped text = %q", got)
	}
	if got := handler.escapeText("*a_b`c~d|e"); got != "\\*a\\_b\\`c\\~d\\|e" {
		t.Fatalf("escaped text = %q", got)
	}
	if got := handler.getStringValue(nil); got != "" {
		t.Fatalf("nil value = %q", got)
	}
	if got := handler.getStringValue("value"); got != "value" {
		t.Fatalf("string value = %q", got)
	}
	if got := handler.getStringValue(42); got != "42" {
		t.Fatalf("numeric value = %q", got)
	}
}

func TestDiscordCurrentServiceFromLeaseWithoutContextValue(t *testing.T) {
	handler := NewHandler(nil)
	response := invokeDiscordHandler(t, http.MethodGet, "/", "", nil, handler.GetStatus)
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestDiscordRouteLeaseWithConcreteService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router.Group(""), service.NewStaticAccessor(&service.DiscordService{}))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/discord/status", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d, body = %s", response.Code, http.StatusInternalServerError, response.Body.String())
	}
}

func newDiscordHandlerForTest(service *discordServiceForTest) *Handler {
	return &Handler{resolveService: func(*gin.Context) discordService {
		if service == nil {
			return nil
		}
		return service
	}}
}

func invokeDiscordHandler(t *testing.T, method, path, body string, params map[string]string, callback gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	contextValue, _ := gin.CreateTestContext(response)
	contextValue.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	contextValue.Request.Header.Set("Content-Type", "application/json")
	for key, value := range params {
		contextValue.Params = append(contextValue.Params, gin.Param{Key: key, Value: value})
	}
	callback(contextValue)
	return response
}

type discordServiceForTest struct {
	sendErr        error
	sendChannelErr error
	testErr        error
	botInfo        interface{}
	botInfoErr     error
	validateErr    error
	lastLevel      string
	lastChannel    string
	lastMessage    string
}

func (s *discordServiceForTest) SendMessage(_ context.Context, level, message string) error {
	s.lastLevel = level
	s.lastMessage = message
	return s.sendErr
}
func (s *discordServiceForTest) SendMessageToChannel(channel, message string) error {
	s.lastChannel = channel
	s.lastMessage = message
	return s.sendChannelErr
}
func (s *discordServiceForTest) TestConnection() error { return s.testErr }
func (s *discordServiceForTest) GetBotInfo() (interface{}, error) {
	return s.botInfo, s.botInfoErr
}
func (s *discordServiceForTest) ValidateChannel(channel string) error {
	s.lastChannel = channel
	return s.validateErr
}
