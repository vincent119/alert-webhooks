package discord

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"alert-webhooks/config"
	"alert-webhooks/pkg/service"

	"github.com/gin-gonic/gin"
)

type routeCase struct {
	name   string
	method string
	path   string
}

func TestSetupRoutesRequiresBasicAuthWhenEnabled(t *testing.T) {
	routes := []routeCase{
		{name: "傳送指定頻道訊息", method: http.MethodPost, path: "/discord/channel/test"},
		{name: "傳送分級訊息", method: http.MethodPost, path: "/discord/chatid_L0"},
		{name: "查詢服務狀態", method: http.MethodGet, path: "/discord/status"},
		{name: "測試頻道", method: http.MethodPost, path: "/discord/test/test"},
		{name: "驗證頻道", method: http.MethodPost, path: "/discord/validate/test"},
	}

	authCases := []struct {
		name       string
		setAuth    func(*http.Request)
		wantStatus int
	}{
		{name: "缺少認證", wantStatus: http.StatusUnauthorized},
		{
			name: "認證格式錯誤",
			setAuth: func(req *http.Request) {
				req.Header.Set("Authorization", "Bearer invalid")
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "認證資訊錯誤",
			setAuth: func(req *http.Request) {
				req.SetBasicAuth("invalid", "invalid")
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, authCase := range authCases {
		t.Run(authCase.name, func(t *testing.T) {
			router := newTestRouter(t, true)

			for _, route := range routes {
				t.Run(route.name, func(t *testing.T) {
					req := httptest.NewRequest(route.method, route.path, nil)
					if authCase.setAuth != nil {
						authCase.setAuth(req)
					}

					resp := httptest.NewRecorder()
					router.ServeHTTP(resp, req)

					if resp.Code != authCase.wantStatus {
						t.Fatalf("status = %d, want %d", resp.Code, authCase.wantStatus)
					}
					if got := resp.Header().Get("WWW-Authenticate"); got == "" {
						t.Fatal("missing WWW-Authenticate header")
					}
				})
			}
		})
	}
}

func TestSetupRoutesAllowsValidBasicAuthAndDisabledAuthentication(t *testing.T) {
	route := routeCase{
		name:   "傳送指定頻道訊息",
		method: http.MethodPost,
		path:   "/discord/channel/test",
	}

	tests := []struct {
		name        string
		authEnabled bool
		setAuth     func(*http.Request)
	}{
		{
			name:        "有效認證可進入既有 handler",
			authEnabled: true,
			setAuth: func(req *http.Request) {
				req.SetBasicAuth("test-user", "test-password")
			},
		},
		{
			name:        "認證關閉時維持 bypass",
			authEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := newTestRouter(t, tt.authEnabled)
			req := httptest.NewRequest(route.method, route.path, nil)
			if tt.setAuth != nil {
				tt.setAuth(req)
			}

			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, want existing handler status %d", resp.Code, http.StatusServiceUnavailable)
			}
		})
	}
}

func newTestRouter(t *testing.T, authEnabled bool) *gin.Engine {
	t.Helper()

	originalWebhooks := config.Webhooks
	config.Webhooks = config.WebhooksConf{
		Enable:           authEnabled,
		BaseAuthUser:     "test-user",
		BaseAuthPassword: "test-password",
	}
	t.Cleanup(func() {
		config.Webhooks = originalWebhooks
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	SetupRoutes(router.Group(""), service.NewStaticAccessor(nil))
	return router
}
