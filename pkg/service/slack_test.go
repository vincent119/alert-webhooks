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

	"github.com/vincent119/zlogger"
)

var slackLogFile string

func TestMain(m *testing.M) {
	logPath, err := os.MkdirTemp("", "alert-webhooks-slack-log-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(logPath)

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
