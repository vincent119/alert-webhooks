package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestBodyLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		body        string
		limit       int64
		unknownSize bool
		wantStatus  int
		wantHandler bool
	}{
		{
			name:        "允許小於上限的 JSON request",
			body:        `{"message":"ok"}`,
			limit:       64,
			wantStatus:  http.StatusNoContent,
			wantHandler: true,
		},
		{
			name:        "拒絕超過上限的 JSON request",
			body:        `{"message":"` + strings.Repeat("a", 65) + `"}`,
			limit:       64,
			wantStatus:  http.StatusRequestEntityTooLarge,
			wantHandler: false,
		},
		{
			name:        "限制未知大小的 request body",
			body:        `{"message":"` + strings.Repeat("a", 65) + `"}`,
			limit:       64,
			unknownSize: true,
			wantStatus:  http.StatusRequestEntityTooLarge,
			wantHandler: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false
			router := gin.New()
			router.Use(RequestBodyLimit(tt.limit))
			router.POST("/messages", func(c *gin.Context) {
				handlerCalled = true
				var payload struct {
					Message string `json:"message"`
				}
				if err := c.ShouldBindJSON(&payload); err != nil {
					c.Status(http.StatusBadRequest)
					return
				}
				c.Status(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodPost, "/messages", bytes.NewBufferString(tt.body))
			if tt.unknownSize {
				req.ContentLength = -1
			}
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			if resp.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", resp.Code, tt.wantStatus)
			}
			if handlerCalled != tt.wantHandler {
				t.Fatalf("handler called = %t, want %t", handlerCalled, tt.wantHandler)
			}
		})
	}
}
