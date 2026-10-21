package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"alert-webhooks/config"

	"github.com/gin-gonic/gin"
)

func TestBasicAuthAllowsRequestWithoutHeaderWhenCredentialsAreEmpty(t *testing.T) {
	router := newBasicAuthTestRouter(t, true, "", "")
	request := httptest.NewRequest(http.MethodGet, "/protected", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestBasicAuthAllowsRequestWithoutHeaderWhenEitherCredentialIsEmpty(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "僅設定帳號",
			username: "test-user",
		},
		{
			name:     "僅設定密碼",
			password: "test-password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newBasicAuthTestRouter(t, true, tt.username, tt.password)
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
			}
		})
	}
}

func TestBasicAuthRequiresCredentialsWhenBothConfigured(t *testing.T) {
	tests := []struct {
		name       string
		setAuth    func(*http.Request)
		wantStatus int
	}{
		{
			name:       "缺少認證資訊",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "錯誤認證資訊",
			setAuth: func(request *http.Request) {
				request.SetBasicAuth("invalid", "invalid")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "正確認證資訊",
			setAuth: func(request *http.Request) {
				request.SetBasicAuth("test-user", "test-password")
			},
			wantStatus: http.StatusNoContent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newBasicAuthTestRouter(t, true, "test-user", "test-password")
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.setAuth != nil {
				tt.setAuth(request)
			}
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}

func TestCredentialsMatchRequiresBothCredentials(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		want     bool
	}{
		{
			name:     "帳密皆正確",
			username: "test-user",
			password: "test-password",
			want:     true,
		},
		{
			name:     "帳號錯誤",
			username: "invalid-user",
			password: "test-password",
		},
		{
			name:     "密碼錯誤",
			username: "test-user",
			password: "invalid-password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := credentialsMatch(tt.username, tt.password, "test-user", "test-password"); got != tt.want {
				t.Fatalf("credentialsMatch() = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestBasicAuthWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BasicAuthWithConfig("test-user", "test-password"))
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
	}{
		{name: "正確認證", username: "test-user", password: "test-password", wantStatus: http.StatusNoContent},
		{name: "錯誤認證", username: "test-user", password: "invalid", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/protected", nil)
			request.SetBasicAuth(tt.username, tt.password)
			response := httptest.NewRecorder()

			router.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}

func newBasicAuthTestRouter(t *testing.T, enabled bool, username, password string) *gin.Engine {
	t.Helper()
	originalWebhooks := config.Webhooks
	config.Webhooks = config.WebhooksConf{
		Enable:           enabled,
		BaseAuthUser:     username,
		BaseAuthPassword: password,
	}
	t.Cleanup(func() {
		config.Webhooks = originalWebhooks
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(BasicAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}
