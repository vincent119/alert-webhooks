package slack

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"alert-webhooks/config"
	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

func TestSlackSendMessageToChannel(t *testing.T) {
	tests := []struct {
		name       string
		service    *slackServiceForTest
		channel    string
		body       string
		wantStatus int
	}{
		{name: "unavailable", channel: "alerts", body: `{"message":"hello"}`, wantStatus: http.StatusServiceUnavailable},
		{name: "missing channel", service: &slackServiceForTest{}, body: `{"message":"hello"}`, wantStatus: http.StatusBadRequest},
		{name: "invalid JSON", service: &slackServiceForTest{}, channel: "alerts", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "empty payload", service: &slackServiceForTest{}, channel: "alerts", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "direct success", service: &slackServiceForTest{}, channel: "alerts", body: `{"message":"hello","username":"bot"}`, wantStatus: http.StatusOK},
		{name: "direct failure", service: &slackServiceForTest{sendOptionsErr: errors.New("send failed")}, channel: "@user", body: `{"message":"hello"}`, wantStatus: http.StatusInternalServerError},
		{name: "raw AlertManager", service: &slackServiceForTest{}, channel: "alerts", body: `{"status":"firing","alerts":[{"status":"firing","annotations":{"summary":"CPU"}}]}`, wantStatus: http.StatusOK},
		{name: "wrapped AlertManager", service: &slackServiceForTest{}, channel: "alerts", body: `{"alertmanager_data":{"status":"firing"}}`, wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newSlackHandlerForTest(tt.service)
			response := invokeSlackHandler(t, http.MethodPost, tt.body, map[string]string{"channel": tt.channel}, handler.SendMessageToChannel)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantStatus == http.StatusOK && tt.service.lastChannel != "#alerts" {
				t.Fatalf("channel = %q", tt.service.lastChannel)
			}
		})
	}
}

func TestSlackSendMessageToLevel(t *testing.T) {
	setSlackHandlerConfig(t, config.SlackConf{})
	tests := []struct {
		name       string
		service    *slackServiceForTest
		level      string
		body       string
		wantStatus int
		wantLevel  string
	}{
		{name: "unavailable", level: "0", body: `{"message":"hello"}`, wantStatus: http.StatusServiceUnavailable},
		{name: "missing level", service: &slackServiceForTest{}, body: `{"message":"hello"}`, wantStatus: http.StatusBadRequest},
		{name: "invalid JSON", service: &slackServiceForTest{}, level: "0", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "empty payload", service: &slackServiceForTest{}, level: "0", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "direct success", service: &slackServiceForTest{}, level: "2", body: `{"message":"hello"}`, wantStatus: http.StatusOK, wantLevel: "L2"},
		{name: "named level", service: &slackServiceForTest{}, level: "critical", body: `{"message":"hello"}`, wantStatus: http.StatusOK, wantLevel: "critical"},
		{name: "send failure", service: &slackServiceForTest{sendLevelErr: errors.New("send failed")}, level: "1", body: `{"message":"hello"}`, wantStatus: http.StatusInternalServerError, wantLevel: "L1"},
		{name: "raw AlertManager", service: &slackServiceForTest{}, level: "3", body: `{"status":"resolved","alerts":[{"status":"resolved"}]}`, wantStatus: http.StatusOK, wantLevel: "L3"},
		{name: "wrapped AlertManager", service: &slackServiceForTest{}, level: "4", body: `{"alertmanager_data":{"status":"firing"}}`, wantStatus: http.StatusOK, wantLevel: "L4"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := invokeSlackHandler(t, http.MethodPost, tt.body, map[string]string{"level": tt.level}, newSlackHandlerForTest(tt.service).SendMessageToLevel)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantLevel != "" && tt.service.lastLevel != tt.wantLevel {
				t.Fatalf("level = %q, want %q", tt.service.lastLevel, tt.wantLevel)
			}
		})
	}
}

func TestSlackSendRichMessage(t *testing.T) {
	tests := []struct {
		name       string
		service    *slackServiceForTest
		channel    string
		body       string
		wantStatus int
		wantColor  string
	}{
		{name: "unavailable", channel: "alerts", body: `{"title":"title"}`, wantStatus: http.StatusServiceUnavailable},
		{name: "missing channel", service: &slackServiceForTest{}, body: `{"title":"title"}`, wantStatus: http.StatusBadRequest},
		{name: "invalid JSON", service: &slackServiceForTest{}, channel: "alerts", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "missing title", service: &slackServiceForTest{}, channel: "alerts", body: `{}`, wantStatus: http.StatusBadRequest},
		{name: "default color", service: &slackServiceForTest{}, channel: "alerts", body: `{"title":"title","message":"body"}`, wantStatus: http.StatusOK, wantColor: "good"},
		{name: "custom color", service: &slackServiceForTest{}, channel: "@user", body: `{"title":"title","color":"danger"}`, wantStatus: http.StatusOK, wantColor: "danger"},
		{name: "send failure", service: &slackServiceForTest{richErr: errors.New("send failed")}, channel: "alerts", body: `{"title":"title"}`, wantStatus: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := invokeSlackHandler(t, http.MethodPost, tt.body, map[string]string{"channel": tt.channel}, newSlackHandlerForTest(tt.service).SendRichMessage)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantColor != "" && tt.service.lastColor != tt.wantColor {
				t.Fatalf("color = %q, want %q", tt.service.lastColor, tt.wantColor)
			}
		})
	}
}

func TestSlackStatusAndChannels(t *testing.T) {
	setSlackHandlerConfig(t, config.SlackConf{Enable: true})
	unavailable := newSlackHandlerForTest(nil)
	status := invokeSlackHandler(t, http.MethodGet, "", nil, unavailable.GetStatus)
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), `"connected":false`) {
		t.Fatalf("status response = %d %s", status.Code, status.Body.String())
	}
	channels := invokeSlackHandler(t, http.MethodGet, "", nil, unavailable.GetChannels)
	if channels.Code != http.StatusServiceUnavailable {
		t.Fatalf("channels status = %d", channels.Code)
	}

	available := newSlackHandlerForTest(&slackServiceForTest{channels: map[string]string{"critical": "#critical"}})
	status = invokeSlackHandler(t, http.MethodGet, "", nil, available.GetStatus)
	if status.Code != http.StatusOK || !strings.Contains(status.Body.String(), "#critical") {
		t.Fatalf("status response = %d %s", status.Code, status.Body.String())
	}
	channels = invokeSlackHandler(t, http.MethodGet, "", nil, available.GetChannels)
	if channels.Code != http.StatusOK || !strings.Contains(channels.Body.String(), "#critical") {
		t.Fatalf("channels response = %d %s", channels.Code, channels.Body.String())
	}
}

func TestSlackTestConnection(t *testing.T) {
	tests := []struct {
		name       string
		cfg        config.SlackConf
		service    *slackServiceForTest
		wantStatus int
		wantTarget string
	}{
		{name: "unavailable", wantStatus: http.StatusServiceUnavailable},
		{name: "missing channel", service: &slackServiceForTest{}, wantStatus: http.StatusBadRequest},
		{name: "default channel", cfg: config.SlackConf{Channel: "#default"}, service: &slackServiceForTest{}, wantStatus: http.StatusOK, wantTarget: "#default"},
		{name: "first configured channel", service: &slackServiceForTest{channels: map[string]string{"first": "#first"}}, wantStatus: http.StatusOK, wantTarget: "#first"},
		{name: "send failure", cfg: config.SlackConf{Channel: "#default"}, service: &slackServiceForTest{sendErr: errors.New("send failed")}, wantStatus: http.StatusInternalServerError, wantTarget: "#default"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setSlackHandlerConfig(t, tt.cfg)
			response := invokeSlackHandler(t, http.MethodPost, "", nil, newSlackHandlerForTest(tt.service).TestConnection)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
			if tt.wantTarget != "" && tt.service.lastChannel != tt.wantTarget {
				t.Fatalf("target = %q, want %q", tt.service.lastChannel, tt.wantTarget)
			}
		})
	}
}

func TestSlackValidateChannel(t *testing.T) {
	tests := []struct {
		name       string
		service    *slackServiceForTest
		channel    string
		wantStatus int
	}{
		{name: "unavailable", channel: "alerts", wantStatus: http.StatusServiceUnavailable},
		{name: "missing channel", service: &slackServiceForTest{}, wantStatus: http.StatusBadRequest},
		{name: "not in channel", service: &slackServiceForTest{sendErr: errors.New("not_in_channel")}, channel: "alerts", wantStatus: http.StatusBadRequest},
		{name: "channel not found", service: &slackServiceForTest{sendErr: errors.New("channel_not_found")}, channel: "alerts", wantStatus: http.StatusBadRequest},
		{name: "other failure", service: &slackServiceForTest{sendErr: errors.New("other")}, channel: "alerts", wantStatus: http.StatusInternalServerError},
		{name: "success", service: &slackServiceForTest{}, channel: "@user", wantStatus: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := invokeSlackHandler(t, http.MethodPost, "", map[string]string{"channel": tt.channel}, newSlackHandlerForTest(tt.service).ValidateChannel)
			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d, body = %s", response.Code, tt.wantStatus, response.Body.String())
			}
		})
	}
}

func TestSlackAlertManagerFormatting(t *testing.T) {
	setSlackHandlerConfig(t, config.SlackConf{TemplateLanguage: "eng"})
	handler := newSlackHandlerForTest(&slackServiceForTest{})
	req := &SendMessageRequest{
		Status: "firing",
		Alerts: []map[string]interface{}{
			{"status": "firing", "startsAt": "2026-10-21T00:00:00Z", "generatorURL": "https://generator", "labels": map[string]interface{}{"namespace": "payments", "pod": "api-0", "ignored": 1}, "annotations": map[string]interface{}{"summary": "high CPU", "ignored": 1}},
			{"status": "resolved", "endsAt": "2026-10-21T01:00:00Z", "labels": map[string]interface{}{"pod": "api-1"}, "annotations": map[string]interface{}{"summary": "resolved"}},
		},
		CommonLabels: map[string]interface{}{"alertname": "CPU", "env": "prod", "severity": "critical"},
		ExternalURL:  "https://alerts.example",
	}
	message := handler.formatAlertManagerMessage(req)
	for _, expected := range []string{"CPU", "payments", "high CPU", "resolved"} {
		if !strings.Contains(message, expected) {
			t.Fatalf("message %q does not contain %q", message, expected)
		}
	}

	service.GetServiceManager().ReloadTemplateEngine()
	if message := handler.formatAlertManagerMessage(req); message == "" {
		t.Fatal("template fallback message is empty")
	}
	if len(slackHandlerLogFields("channel", "#alerts", nil)) != 2 {
		t.Fatal("unexpected log field count without error")
	}
	if len(slackHandlerLogFields("channel", "#alerts", errors.New("failed"))) != 3 {
		t.Fatal("unexpected log field count with error")
	}
	if len(slackTemplateLogFields("eng", []string{"eng"})) != 3 {
		t.Fatal("unexpected template log field count")
	}
}

func TestSlackRoutesRejectUnavailableService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterSlackRoutes(router.Group(""), service.NewStaticAccessor(nil))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/slack/status", nil))
	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusServiceUnavailable)
	}
}

func TestSlackCurrentServiceWithoutLease(t *testing.T) {
	response := invokeSlackHandler(t, http.MethodGet, "", nil, NewHandler(nil).GetStatus)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"connected":false`) {
		t.Fatalf("response = %d %s", response.Code, response.Body.String())
	}
}

func newSlackHandlerForTest(value *slackServiceForTest) *Handler {
	return &Handler{resolveService: func(*gin.Context) slackService {
		if value == nil {
			return nil
		}
		return value
	}}
}

func invokeSlackHandler(t *testing.T, method, body string, params map[string]string, callback gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	contextValue, _ := gin.CreateTestContext(response)
	contextValue.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	contextValue.Request.Header.Set("Content-Type", "application/json")
	for key, value := range params {
		contextValue.Params = append(contextValue.Params, gin.Param{Key: key, Value: value})
	}
	callback(contextValue)
	return response
}

func setSlackHandlerConfig(t *testing.T, value config.SlackConf) {
	t.Helper()
	original := config.Slack
	config.Slack = value
	t.Cleanup(func() { config.Slack = original })
}

type slackServiceForTest struct {
	sendErr        error
	sendOptionsErr error
	sendLevelErr   error
	richErr        error
	channels       map[string]string
	lastChannel    string
	lastLevel      string
	lastMessage    string
	lastColor      string
}

func (s *slackServiceForTest) SendMessage(_ context.Context, channel, message string) error {
	s.lastChannel, s.lastMessage = channel, message
	return s.sendErr
}
func (s *slackServiceForTest) SendMessageWithOptions(_ context.Context, channel, message string, _ *service.SlackMessage) error {
	s.lastChannel, s.lastMessage = channel, message
	return s.sendOptionsErr
}
func (s *slackServiceForTest) SendMessageToLevel(_ context.Context, level, message string) error {
	s.lastLevel, s.lastMessage = level, message
	return s.sendLevelErr
}
func (s *slackServiceForTest) SendRichMessage(_ context.Context, channel, _, _, color string, _ []service.Field) error {
	s.lastChannel, s.lastColor = channel, color
	return s.richErr
}
func (s *slackServiceForTest) GetChannels() map[string]string { return s.channels }
