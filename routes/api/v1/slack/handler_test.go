package slack

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSlackHandlerLogFieldsDoesNotLogAlertContent(t *testing.T) {
	fields := slackHandlerLogFields("channel", "#alerts", errors.New("Slack request failed"))
	fields = append(fields, slackTemplateLogFields("eng", []string{"eng", "zhtw"})...)
	keys := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		keys[field.Key] = struct{}{}
		switch field.Key {
		case "message", "title", "message_preview", "text":
			t.Fatalf("handler log fields contain alert content key %q", field.Key)
		}
	}
	for _, key := range []string{"category", "channel", "error", "language", "available_languages"} {
		if _, ok := keys[key]; !ok {
			t.Errorf("handler log fields do not retain metadata key %q", key)
		}
	}
}

func TestRespondErrorDoesNotExposeInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)

	respondError(context, http.StatusInternalServerError, "Failed to send message", "SLACK_MESSAGE_SEND_FAILED", errors.New("upstream token rejected"))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if got := response.Body.String(); got == "" || contains(got, "upstream token rejected") {
		t.Fatalf("response unexpectedly exposes internal error: %s", got)
	}
	if !contains(response.Body.String(), "SLACK_MESSAGE_SEND_FAILED") {
		t.Fatalf("response does not contain stable error code: %s", response.Body.String())
	}
}

func contains(value, expected string) bool {
	return strings.Contains(value, expected)
}
