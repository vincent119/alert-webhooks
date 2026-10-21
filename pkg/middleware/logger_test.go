package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vincent119/zlogger"
)

func TestLoggerMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logPath := t.TempDir()
	level := "info"
	format := "json"
	outputs := []string{"file"}
	fileName := "middleware.log"
	cleanup, err := zlogger.Configure(&zlogger.ConfigPatch{
		Level:    &level,
		Format:   &format,
		Outputs:  &outputs,
		LogPath:  &logPath,
		FileName: &fileName,
	})
	if err != nil {
		t.Fatalf("Configure() error = %v", err)
	}
	t.Cleanup(func() {
		if err := cleanup(); err != nil {
			t.Errorf("cleanup() error = %v", err)
		}
	})
	logFile := filepath.Join(logPath, fileName)

	t.Run("Logger 記錄 request ID 與預設 category", func(t *testing.T) {
		router := gin.New()
		router.Use(Logger())
		router.GET("/ready", func(c *gin.Context) { c.Status(http.StatusNoContent) })

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}

		entry := readLastMiddlewareLogEntry(t, logFile)
		if got := entry["msg"]; got != "HTTP request" {
			t.Fatalf("msg = %v, want HTTP request", got)
		}
		if got := entry["category"]; got != "default" {
			t.Fatalf("category = %v, want default", got)
		}
		if got, ok := entry["request_id"].(string); !ok || got == "" {
			t.Fatalf("request_id = %v, want non-empty string", entry["request_id"])
		}
	})

	t.Run("Logger 記錄 Gin error", func(t *testing.T) {
		router := gin.New()
		router.Use(Logger())
		router.GET("/failure", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
			c.Status(http.StatusBadRequest)
		})

		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/failure?source=test", nil))

		entry := readLastMiddlewareLogEntry(t, logFile)
		if got := entry["level"]; got != "ERROR" {
			t.Fatalf("level = %v, want ERROR", got)
		}
		if got := entry["query"]; got != "source=test" {
			t.Fatalf("query = %v, want source=test", got)
		}
	})

	t.Run("LoggerWithSkipPaths 跳過指定路徑並記錄其他路徑", func(t *testing.T) {
		router := gin.New()
		router.Use(LoggerWithSkipPaths([]string{"/health"}))
		router.GET("/health", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		router.GET("/ready", func(c *gin.Context) { c.Status(http.StatusNoContent) })
		router.GET("/failure", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
			c.Status(http.StatusBadRequest)
		})

		skippedResponse := httptest.NewRecorder()
		router.ServeHTTP(skippedResponse, httptest.NewRequest(http.MethodGet, "/health", nil))
		if got := skippedResponse.Header().Get("X-Request-ID"); got != "" {
			t.Fatalf("skipped request ID = %q, want empty", got)
		}

		readyResponse := httptest.NewRecorder()
		router.ServeHTTP(readyResponse, httptest.NewRequest(http.MethodGet, "/ready", nil))
		if got := readyResponse.Header().Get("X-Request-ID"); got == "" {
			t.Fatal("logged request ID is empty")
		}

		failureResponse := httptest.NewRecorder()
		router.ServeHTTP(failureResponse, httptest.NewRequest(http.MethodGet, "/failure", nil))

		entry := readLastMiddlewareLogEntry(t, logFile)
		if got := entry["level"]; got != "ERROR" {
			t.Fatalf("level = %v, want ERROR", got)
		}
	})
}

func readLastMiddlewareLogEntry(t *testing.T, path string) map[string]interface{} {
	t.Helper()
	if err := zlogger.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("log is not valid JSON: %v", err)
	}
	return entry
}
