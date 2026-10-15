package telegram

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorResponseUsesStablePublicFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)

	context.JSON(http.StatusInternalServerError, errorResponse("Failed to send message", "TELEGRAM_MESSAGE_SEND_FAILED"))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if got := response.Body.String(); !strings.Contains(got, "TELEGRAM_MESSAGE_SEND_FAILED") {
		t.Fatalf("response does not contain stable error code: %s", got)
	}
}

func TestSanitizeDebugPayloadRedactsSensitiveFieldsRecursively(t *testing.T) {
	body := []byte(`{
		"token":"top-secret",
		"nested":{"Api_Key":"nested-secret","client_secret":"client-secret","safe":"visible"},
		"items":[{"PASSWORD":"array-secret"}]
	}`)

	sanitized, ok := sanitizeDebugPayload(body)
	if !ok {
		t.Fatal("sanitizeDebugPayload returned false for valid JSON")
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(sanitized), &payload); err != nil {
		t.Fatalf("unmarshal sanitized payload: %v", err)
	}
	if payload["token"] != "[REDACTED]" {
		t.Fatalf("token = %v, want [REDACTED]", payload["token"])
	}
	nested := payload["nested"].(map[string]interface{})
	if nested["Api_Key"] != "[REDACTED]" || nested["client_secret"] != "[REDACTED]" || nested["safe"] != "visible" {
		t.Fatalf("nested payload = %#v", nested)
	}
	items := payload["items"].([]interface{})
	item := items[0].(map[string]interface{})
	if item["PASSWORD"] != "[REDACTED]" {
		t.Fatalf("items payload = %#v", items)
	}
}

func TestSanitizeDebugPayloadLimitsOutputLength(t *testing.T) {
	body := []byte(`{"message":"` + strings.Repeat("x", maxSanitizedPayloadLogLength+1) + `"}`)

	sanitized, ok := sanitizeDebugPayload(body)
	if !ok {
		t.Fatal("sanitizeDebugPayload returned false for valid JSON")
	}
	if got, want := len([]rune(sanitized)), maxSanitizedPayloadLogLength+3; got != want {
		t.Fatalf("sanitized payload length = %d, want %d", got, want)
	}
}

func TestSanitizeDebugPayloadRedactsAnnotations(t *testing.T) {
	body := []byte(`{
		"annotations":{"summary":"internal URL https://admin.example","count":2},
		"commonAnnotations":{"owner":"ops@example.com"},
		"commonLabels":{"service":"payments"}
	}`)

	sanitized, ok := sanitizeDebugPayload(body)
	if !ok {
		t.Fatal("sanitizeDebugPayload returned false for valid JSON")
	}
	for _, value := range []string{"admin.example", "ops@example.com"} {
		if strings.Contains(sanitized, value) {
			t.Fatalf("sanitized payload exposes annotation value %q: %s", value, sanitized)
		}
	}
	if !strings.Contains(sanitized, `"service":"payments"`) {
		t.Fatalf("sanitized payload unexpectedly removes non-sensitive labels: %s", sanitized)
	}
}

func TestSanitizeDebugPayloadRejectsInvalidJSON(t *testing.T) {
	if sanitized, ok := sanitizeDebugPayload([]byte(`{"token":"secret"`)); ok || sanitized != "" {
		t.Fatalf("sanitizeDebugPayload = (%q, %t), want empty payload and false", sanitized, ok)
	}
}
