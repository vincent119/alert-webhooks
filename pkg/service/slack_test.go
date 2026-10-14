package service

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

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
			t.Fatal("Slack HTTP request 未使用 caller context")
		}
	case <-time.After(time.Second):
		t.Fatal("Slack HTTP request 未開始")
	}

	cancel()
	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("取消 context 後的錯誤無法辨識：%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("取消 context 後 Slack 傳送未結束")
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
		t.Fatal("Slack HTTP request 未開始")
	}

	updated := make(chan struct{})
	go func() {
		service.SetChannelForLevel("critical", "#critical")
		close(updated)
	}()

	select {
	case <-updated:
	case <-time.After(time.Second):
		t.Fatal("Slack HTTP I/O 期間阻塞 channel 更新")
	}

	close(release)
	if err := <-sendErr; err != nil {
		t.Fatalf("Slack 傳送失敗：%v", err)
	}
}

func TestSlackSendMessagePreservesPayloadAndHeaders(t *testing.T) {
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if got, want := req.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Fatalf("Authorization 不正確：got %q, want %q", got, want)
		}
		if got, want := req.Header.Get("Content-Type"), "application/json"; got != want {
			t.Fatalf("Content-Type 不正確：got %q, want %q", got, want)
		}
		body, err := io.ReadAll(req.Body)
		if err != nil {
			t.Fatalf("讀取 request body 失敗：%v", err)
		}
		if got, want := string(body), `{"channel":"#alerts","text":"測試訊息","username":"覆寫名稱"}`; got != want {
			t.Fatalf("Slack payload 不正確：got %q, want %q", got, want)
		}
		return slackTestResponse(`{"ok":true}`), nil
	}))

	if err := service.SendMessageWithOptions(context.Background(), "#alerts", "測試訊息", &SlackMessage{Username: "覆寫名稱"}); err != nil {
		t.Fatalf("Slack 傳送失敗：%v", err)
	}
}

func TestSlackTestConnectionClosesResponseBody(t *testing.T) {
	body := &closeTrackingReadCloser{Reader: strings.NewReader(`{"ok":true}`)}
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponseWithBody(body), nil
	}))

	if err := service.TestConnection(); err != nil {
		t.Fatalf("Slack 連線測試失敗：%v", err)
	}
	if got, want := body.closeCount, 1; got != want {
		t.Fatalf("response body 關閉次數不正確：got %d, want %d", got, want)
	}
}

func TestSlackSendMessageClosesResponseBody(t *testing.T) {
	body := &closeTrackingReadCloser{Reader: strings.NewReader(`{"ok":true}`)}
	service := newSlackServiceForTest("test-token", roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return slackTestResponseWithBody(body), nil
	}))

	if err := service.SendMessageWithOptions(context.Background(), "#alerts", "測試訊息", nil); err != nil {
		t.Fatalf("Slack 傳送失敗：%v", err)
	}
	if got, want := body.closeCount, 1; got != want {
		t.Fatalf("response body 關閉次數不正確：got %d, want %d", got, want)
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
		t.Fatal("Slack API 回應失敗時預期應回傳錯誤")
	}
	if got, want := err.Error(), "Invalid Slack token. Please check if the token in configuration is correct"; got != want {
		t.Fatalf("主要錯誤被 cleanup error 改變：got %q, want %q", got, want)
	}
	if got, want := body.closeCount, 1; got != want {
		t.Fatalf("response body 關閉次數不正確：got %d, want %d", got, want)
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
