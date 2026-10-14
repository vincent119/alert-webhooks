package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthCheckReturnsStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/healthz", HealthCheck)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.Code, http.StatusOK)
	}
	if got, want := resp.Body.String(), `{"status":"OK"}`; got != want {
		t.Errorf("body = %s, want %s", got, want)
	}
}
