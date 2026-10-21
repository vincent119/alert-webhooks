package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRecoveryHandlesPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Recovery())
	router.GET("/panic", func(c *gin.Context) {
		c.Set("HttpRequest", "test-request")
		panic("test panic")
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))
}

func TestRecoveryPassesThroughNormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Recovery())
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ok", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}

func TestRecoveryWithWriterHandlesPanicWithAndWithoutStack(t *testing.T) {
	for _, stack := range []bool{false, true} {
		t.Run(map[bool]string{false: "不輸出 stack", true: "輸出 stack"}[stack], func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			router := gin.New()
			router.Use(RecoveryWithWriter(stack))
			router.GET("/panic", func(*gin.Context) { panic("test panic") })

			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))
		})
	}
}

func TestRecoveryWithWriterPassesThroughNormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RecoveryWithWriter(true))
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusNoContent) })

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ok", nil))

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
}
