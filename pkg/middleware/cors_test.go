package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"alert-webhooks/config"

	"github.com/gin-gonic/gin"
)

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	router := newCORSRouter([]string{"https://console.example.com"})
	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "https://console.example.com")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "https://console.example.com" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Access-Control-Allow-Credentials = %q", got)
	}
	if got := response.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q, want Origin", got)
	}
}

func TestCORSRejectsUnconfiguredOrigin(t *testing.T) {
	router := newCORSRouter([]string{"https://console.example.com"})
	request := httptest.NewRequest(http.MethodOptions, "/", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("rejected origin still received Allow-Origin: %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("rejected origin still received Allow-Credentials: %q", got)
	}
}

func TestCORSDoesNotAllowWildcardOrigin(t *testing.T) {
	router := newCORSRouter([]string{"*"})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://attacker.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("wildcard must not authorize origin, got %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Fatalf("wildcard must not allow credentials, got %q", got)
	}
}

func TestCORSReadsCurrentConfigurationForEveryRequest(t *testing.T) {
	originalWebhooks := config.Webhooks
	config.Webhooks = config.WebhooksConf{CORSAllowedOrigins: []string{"https://first.example"}}
	t.Cleanup(func() {
		config.Webhooks = originalWebhooks
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS())
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })

	firstRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	firstRequest.Header.Set("Origin", "https://first.example")
	firstResponse := httptest.NewRecorder()
	router.ServeHTTP(firstResponse, firstRequest)
	if got := firstResponse.Header().Get("Access-Control-Allow-Origin"); got != "https://first.example" {
		t.Fatalf("first Allow-Origin = %q", got)
	}

	config.Webhooks.CORSAllowedOrigins = []string{"https://second.example"}
	secondRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	secondRequest.Header.Set("Origin", "https://second.example")
	secondResponse := httptest.NewRecorder()
	router.ServeHTTP(secondResponse, secondRequest)
	if got := secondResponse.Header().Get("Access-Control-Allow-Origin"); got != "https://second.example" {
		t.Fatalf("second Allow-Origin = %q", got)
	}
}

func TestCORSContinuesAllowedNonPreflightRequest(t *testing.T) {
	router := newCORSRouter([]string{"  https://console.example.com  ", ""})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Origin", "https://console.example.com")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if got := response.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("Access-Control-Allow-Methods is empty")
	}
}

func TestIsAllowedOriginRejectsEmptyOrigin(t *testing.T) {
	if isAllowedOrigin("", []string{"https://console.example.com"}) {
		t.Fatal("empty origin must be rejected")
	}
}

func newCORSRouter(allowedOrigins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(corsWithOriginProvider(func() []string { return allowedOrigins }))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}
