package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"alert-webhooks/config"

	"github.com/vincent119/zlogger"
)

var slackLogFile string

func TestMain(m *testing.M) {
	logPath, err := os.MkdirTemp("", "alert-webhooks-slack-log-*")
	if err != nil {
		panic(err)
	}

	level := "info"
	format := "json"
	outputs := []string{"file"}
	fileName := "slack.log"
	cleanup, err := zlogger.Configure(&zlogger.ConfigPatch{
		Level:    &level,
		Format:   &format,
		Outputs:  &outputs,
		LogPath:  &logPath,
		FileName: &fileName,
	})
	if err != nil {
		panic(err)
	}

	slackLogFile = filepath.Join(logPath, fileName)
	code := m.Run()
	if err := cleanup(); err != nil && code == 0 {
		code = 1
	}
	if err := os.RemoveAll(logPath); err != nil && code == 0 {
		code = 1
	}
	os.Exit(code)
}

func TestSlackSendMessageUsesRequestContext(t *testing.T) {
	started := make(chan context.Context, 1)
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		started <- req.Context()
		<-req.Context().Done()
		return nil, req.Context().Err()
	}))

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	errCh := make(chan error, 1)
	go func() {
		errCh <- service.SendMessageWithOptions(ctx, "#alerts", "測試訊息", nil)
	}()

	select {
	case requestCtx := <-started:
		if requestCtx != ctx {
			t.Fatal("Slack HTTP request did not use caller context")
		}
	case <-time.After(time.Second):
		t.Fatal("Slack HTTP request did not start")
	}

	cancel()
	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error after context cancellation is not identifiable: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("Slack send did not finish after context cancellation")
	}
}

func TestSlackSendDoesNotHoldLockDuringHTTPIO(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		close(started)
		<-release
		return slackTestResponse(`{"ok":true}`), nil
	}))

	sendErr := make(chan error, 1)
	go func() {
		sendErr <- service.SendMessageWithOptions(context.Background(), "#alerts", "測試訊息", nil)
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("Slack HTTP request did not start")
	}

	updated := make(chan struct{})
	go func() {
		service.SetChannelForLevel("critical", "#critical")
		close(updated)
	}()

	select {
	case <-updated:
	case <-time.After(time.Second):
		t.Fatal("Slack HTTP I/O blocked channel update")
	}

	close(release)
	if err := <-sendErr; err != nil {
		t.Fatalf("Slack send failed: %v", err)
	}
}

func TestSlackSendMessagePreservesPayloadAndHeaders(t *testing.T) {
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Fatalf("Authorization is incorrect: got %q, want %q", got, want)
		}
		if got, want := req.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("Content-Type is incorrect: got %q, want %q", got, want)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("read request body failed: %v", err)
		}
		if got, want := string(body), `{"channel":"#alerts","text":"測試訊息","username":"覆寫名稱"}`; got != want {
			t.Fatalf("Slack payload is incorrect: got %q, want %q", got, want)
		}
		return slackTestResponse(`{"ok":true}`), nil
	}))

	if err := service.SendMessageWithOptions(context.Background(), "#alerts", "測試訊息", &SlackMessage{Username: "覆寫名稱"}); err != nil {
		t.Fatalf("Slack send failed: %v", err)
	}
}

func TestSlackSendMessageDoesNotLogMessageText(t *testing.T) {
	const sensitiveMessage = "S-02-敏感正文-成功"
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponse(`{"ok":true}`), nil
	}))

	if err := service.SendMessageWithOptions(context.Background(), "#alerts", sensitiveMessage, nil); err != nil {
		t.Fatalf("Slack send failed: %v", err)
	}

	assertSlackLogDoesNotContainMessageText(t, sensitiveMessage)
}

func TestSlackSendMessageFailureDoesNotLogMessageText(t *testing.T) {
	const sensitiveMessage = "S-02-敏感正文-失敗"
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("Slack transport failed")
	}))

	if err := service.SendMessageWithOptions(context.Background(), "#alerts", sensitiveMessage, nil); err == nil {
		t.Fatal("expected Slack send error")
	}

	assertSlackLogDoesNotContainMessageText(t, sensitiveMessage)
}

func TestSlackTestConnectionClosesResponseBody(t *testing.T) {
	body := &closeTrackingReadCloser{Reader: strings.NewReader(`{"ok":true}`)}
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponseWithBody(body), nil
	}))

	if err := service.TestConnection(); err != nil {
		t.Fatalf("Slack connection test failed: %v", err)
	}
	if got, want := body.closeCount, 1; got != want {
		t.Fatalf("response body close count is incorrect: got %d, want %d", got, want)
	}
}

func TestSlackSendMessageClosesResponseBody(t *testing.T) {
	body := &closeTrackingReadCloser{Reader: strings.NewReader(`{"ok":true}`)}
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponseWithBody(body), nil
	}))

	if err := service.SendMessageWithOptions(context.Background(), "#alerts", "測試訊息", nil); err != nil {
		t.Fatalf("Slack send failed: %v", err)
	}
	if got, want := body.closeCount, 1; got != want {
		t.Fatalf("response body close count is incorrect: got %d, want %d", got, want)
	}
}

func TestSlackSendMessagePreservesAPIErrorWhenResponseBodyCloseFails(t *testing.T) {
	closeErr := errors.New("response body close failed")
	body := &closeTrackingReadCloser{
		Reader:   strings.NewReader(`{"ok":false,"error":"invalid_auth"}`),
		closeErr: closeErr,
	}
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponseWithBody(body), nil
	}))

	err := service.SendMessageWithOptions(context.Background(), "#alerts", "測試訊息", nil)
	if err == nil {
		t.Fatal("expected an error for failed Slack API response")
	}
	if got, want := err.Error(), "Invalid Slack token. Please check if the token in configuration is correct"; got != want {
		t.Fatalf("primary error was changed by cleanup error: got %q, want %q", got, want)
	}
	if got, want := body.closeCount, 1; got != want {
		t.Fatalf("response body close count is incorrect: got %d, want %d", got, want)
	}
}

func TestNewSlackService(t *testing.T) {
	t.Run("拒絕空 token", func(t *testing.T) {
		_, err := newSlackService("", config.Config{}, &http.Client{})
		if err == nil {
			t.Fatal("expected missing token error")
		}
	})

	t.Run("使用多頻道設定", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return slackTestResponse(`{"ok":true}`), nil
		})}
		cfg := config.Config{Slack: config.SlackConf{Channels: map[string]string{"chat_ids0": "#critical"}}}
		service, err := newSlackService("token", cfg, client)
		if err != nil {
			t.Fatalf("newSlackService() error = %v", err)
		}
		if got := service.channels["chat_ids0"]; got != "#critical" {
			t.Fatalf("channel = %q", got)
		}
	})

	t.Run("使用預設頻道", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return slackTestResponse(`{"ok":true}`), nil
		})}
		cfg := config.Config{Slack: config.SlackConf{Channel: "#default"}}
		service, err := newSlackService("token", cfg, client)
		if err != nil {
			t.Fatalf("newSlackService() error = %v", err)
		}
		if got := service.channels["default"]; got != "#default" {
			t.Fatalf("default channel = %q", got)
		}
	})

	t.Run("連線測試失敗", func(t *testing.T) {
		client := &http.Client{Transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, errors.New("transport failed")
		})}
		_, err := newSlackService("token", config.Config{}, client)
		if err == nil || !strings.Contains(err.Error(), "failed to test slack connection") {
			t.Fatalf("error = %v", err)
		}
	})
}

func TestSlackTestConnectionErrors(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		want      string
	}{
		{
			name: "transport error",
			transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("transport failed")
			}),
			want: "transport failed",
		},
		{
			name: "read error",
			transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return slackTestResponseWithBody(&failingSlackBody{readErr: errors.New("read failed")}), nil
			}),
			want: "read failed",
		},
		{
			name: "invalid JSON",
			transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return slackTestResponse("not-json"), nil
			}),
			want: "invalid character",
		},
		{
			name: "API error",
			transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
				return slackTestResponse(`{"ok":false,"error":"invalid_auth"}`), nil
			}),
			want: "slack auth test failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newSlackServiceForTest("token", tt.transport)
			err := service.TestConnection()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func TestSlackSendMessageOptionsAndWrapper(t *testing.T) {
	setSlackConfigForTest(t, config.SlackConf{
		Username:    "default-name",
		IconURL:     "https://example.com/default.png",
		IconEmoji:   ":bell:",
		LinkNames:   true,
		UnfurlLinks: true,
		UnfurlMedia: true,
	})

	requestCount := 0
	service := newSlackServiceForTest("token", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		requestCount++
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("ReadAll() error = %v", err)
		}
		payload := string(body)
		expectedValues := []string{"default-name", "default.png", ":bell:"}
		if requestCount == 2 {
			expectedValues = []string{"override-name", "override.png", ":warning:", "thread", "attachments", "blocks"}
		}
		for _, expected := range expectedValues {
			if !strings.Contains(payload, expected) {
				t.Fatalf("payload %q does not contain %q", payload, expected)
			}
		}
		return slackTestResponse(`{"ok":true}`), nil
	}))
	options := &SlackMessage{
		Username:  "override-name",
		IconURL:   "https://example.com/override.png",
		IconEmoji: ":warning:",
		ThreadTS:  "thread",
		Attachments: []Attachment{{
			Text: "attachments",
		}},
		Blocks: []Block{{Type: "section", Text: "blocks"}},
	}
	if err := service.SendMessage(context.Background(), "#alerts", "message"); err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if err := service.SendMessageWithOptions(context.Background(), "#alerts", "message", options); err != nil {
		t.Fatalf("SendMessageWithOptions() error = %v", err)
	}
}

func TestSlackSendMessageToLevel(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.SlackConf
		level   string
		channel string
		wantErr bool
	}{
		{name: "L level mapping", cfg: config.SlackConf{Channels: map[string]string{"chat_ids2": "#level2"}}, level: "L2", channel: "#level2"},
		{name: "original mapping fallback", cfg: config.SlackConf{Channels: map[string]string{"warning": "#warning"}}, level: "warning", channel: "#warning"},
		{name: "lowercase mapping fallback", cfg: config.SlackConf{Channels: map[string]string{"critical": "#critical"}}, level: "CRITICAL", channel: "#critical"},
		{name: "default channel fallback", cfg: config.SlackConf{Channel: "#default"}, level: "L9", channel: "#default"},
		{name: "missing channel", cfg: config.SlackConf{}, level: "L9", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setSlackConfigForTest(t, tt.cfg)
			var gotChannel string
			service := newSlackServiceForTest("token", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
				body, _ := io.ReadAll(req.Body)
				gotChannel = string(body)
				return slackTestResponse(`{"ok":true}`), nil
			}))
			err := service.SendMessageToLevel(context.Background(), tt.level, "message")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected channel error")
				}
				return
			}
			if err != nil {
				t.Fatalf("SendMessageToLevel() error = %v", err)
			}
			if !strings.Contains(gotChannel, `"channel":"`+tt.channel+`"`) {
				t.Fatalf("payload = %q", gotChannel)
			}
		})
	}
}

func TestSlackRichMessageAndChannelAccessors(t *testing.T) {
	setSlackConfigForTest(t, config.SlackConf{})
	service := newSlackServiceForTest("token", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(req.Body)
		if !strings.Contains(string(body), "Alert Webhooks") || !strings.Contains(string(body), "title") {
			t.Fatalf("rich payload = %q", body)
		}
		return slackTestResponse(`{"ok":true}`), nil
	}))
	if err := service.SendRichMessage(context.Background(), "#alerts", "title", "message", "danger", []Field{{Title: "key", Value: "value"}}); err != nil {
		t.Fatalf("SendRichMessage() error = %v", err)
	}

	service.SetChannelForLevel("critical", "#critical")
	if channel, ok := service.GetChannelForLevel("critical"); !ok || channel != "#critical" {
		t.Fatalf("channel = %q, %t", channel, ok)
	}
	channels := service.GetChannels()
	channels["critical"] = "changed"
	if channel, _ := service.GetChannelForLevel("critical"); channel != "#critical" {
		t.Fatal("GetChannels must return a defensive copy")
	}
}

func TestSlackSendMessageResponseErrors(t *testing.T) {
	tests := []struct {
		name      string
		transport http.RoundTripper
		want      string
	}{
		{name: "read", transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return slackTestResponseWithBody(&failingSlackBody{readErr: errors.New("read failed")}), nil
		}), want: "failed to read response"},
		{name: "parse", transport: roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return slackTestResponse("not-json"), nil
		}), want: "failed to parse response"},
		{name: "not in channel", transport: slackAPIErrorTransport("not_in_channel"), want: "Bot is not in channel"},
		{name: "channel not found", transport: slackAPIErrorTransport("channel_not_found"), want: "does not exist"},
		{name: "missing scope", transport: slackAPIErrorTransport("missing_scope"), want: "lacks necessary permissions"},
		{name: "unknown", transport: slackAPIErrorTransport("unknown_error"), want: "Slack API error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := newSlackServiceForTest("token", tt.transport)
			err := service.SendMessageWithOptions(context.Background(), "#alerts", "message", nil)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error = %v, want containing %q", err, tt.want)
			}
		})
	}
}

func newSlackServiceForTest(token string, transport http.RoundTripper) *SlackService {
	return &SlackService{
		token:    token,
		client:   &http.Client{Transport: transport},
		channels: map[string]string{},
	}
}

func slackTestResponse(body string) *http.Response {
	return slackTestResponseWithBody(io.NopCloser(strings.NewReader(body)))
}

func slackTestResponseWithBody(body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       body,
		Header:     make(http.Header),
	}
}

type closeTrackingReadCloser struct {
	io.Reader
	closeErr   error
	closeCount int
}

func (r *closeTrackingReadCloser) Close() error {
	r.closeCount++
	return r.closeErr
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

type failingSlackBody struct {
	readErr error
}

func (b *failingSlackBody) Read([]byte) (int, error) { return 0, b.readErr }
func (b *failingSlackBody) Close() error             { return nil }

func slackAPIErrorTransport(apiError string) http.RoundTripper {
	return roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponse(`{"ok":false,"error":"` + apiError + `"}`), nil
	})
}

func setSlackConfigForTest(t *testing.T, value config.SlackConf) {
	t.Helper()
	original := config.Slack
	config.Slack = value
	t.Cleanup(func() { config.Slack = original })
}

func assertSlackLogDoesNotContainMessageText(t *testing.T, sensitiveMessage string) {
	t.Helper()
	if err := zlogger.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	content, err := os.ReadFile(slackLogFile)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	if strings.Contains(string(content), sensitiveMessage) {
		t.Fatalf("log contains sensitive message: %s", content)
	}
	if strings.Contains(string(content), `"text"`) {
		t.Fatalf("log contains text field: %s", content)
	}
}
