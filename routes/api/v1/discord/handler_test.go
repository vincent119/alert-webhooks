package discord

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRespondErrorDoesNotExposeInternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)

	respondError(context, http.StatusInternalServerError, "Failed to send message", "DISCORD_MESSAGE_SEND_FAILED", errors.New("discord token rejected"))

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if got := response.Body.String(); strings.Contains(got, "discord token rejected") {
		t.Fatalf("response unexpectedly exposes internal error: %s", got)
	}
	if !strings.Contains(response.Body.String(), "DISCORD_MESSAGE_SEND_FAILED") {
		t.Fatalf("response does not contain stable error code: %s", response.Body.String())
	}
}
