package v1

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterApiV1RoutesRegistersHealthWhenServicesAreUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterApiV1Routes(router.Group("/api/v1"))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("health status = %d，預期 %d", response.Code, http.StatusOK)
	}

	for _, path := range []string{
		"/api/v1/telegram/info",
		"/api/v1/slack/status",
		"/api/v1/discord/status",
	} {
		response = httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("未啟用服務 %s status = %d，預期 %d", path, response.Code, http.StatusNotFound)
		}
	}
}
