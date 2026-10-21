package middleware

import (
	"bytes"
	"errors"
	"io"
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

func TestRequestBodyLimitSkipsWhenLimitOrBodyDoesNotRequireReading(t *testing.T) {
	tests := []struct {
		name      string
		limit     int64
		body      io.Reader
		setNoBody bool
	}{
		{name: "非正數上限", limit: 0, body: strings.NewReader("payload")},
		{name: "沒有 request body", limit: 10, setNoBody: true},
		{name: "空 request body", limit: 10, body: strings.NewReader("")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(RequestBodyLimit(tt.limit))
			router.POST("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			request := httptest.NewRequest(http.MethodPost, "/", tt.body)
			if tt.setNoBody {
				request.Body = nil
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
			}
		})
	}
}

func TestRequestBodyLimitRejectsReadOrCloseFailure(t *testing.T) {
	tests := []struct {
		name string
		body io.ReadCloser
	}{
		{name: "讀取失敗", body: &failingRequestBody{readErr: errors.New("read failed")}},
		{name: "關閉失敗", body: &failingRequestBody{content: []byte("ok"), closeErr: errors.New("close failed")}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(RequestBodyLimit(10))
			router.POST("/", func(c *gin.Context) { c.Status(http.StatusNoContent) })

			request := httptest.NewRequest(http.MethodPost, "/", nil)
			request.Body = tt.body
			request.ContentLength = -1
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

type failingRequestBody struct {
	content  []byte
	readErr  error
	closeErr error
	read     bool
}

func (b *failingRequestBody) Read(p []byte) (int, error) {
	if b.readErr != nil {
		return 0, b.readErr
	}
	if b.read {
		return 0, io.EOF
	}
	b.read = true
	return copy(p, b.content), nil
}

func (b *failingRequestBody) Close() error {
	return b.closeErr
}
