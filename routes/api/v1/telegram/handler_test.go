package telegram

import (
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
