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

func TestLoggerRecordsRequestIDAndDefaultCategory(t *testing.T) {
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

	router := gin.New()
	router.Use(Logger())
	router.GET("/ready", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/ready", nil))
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if err := zlogger.Sync(); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	content, err := os.ReadFile(filepath.Join(logPath, fileName))
	if err != nil {
		t.Fatalf("read log file failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &entry); err != nil {
		t.Fatalf("log is not valid JSON: %v", err)
	}
	if got := entry["msg"]; got != "HTTP request" {
		t.Fatalf("msg = %v, want HTTP request", got)
	}
	if got := entry["category"]; got != "default" {
		t.Fatalf("category = %v, want default", got)
	}
	if got, ok := entry["request_id"].(string); !ok || got == "" {
		t.Fatalf("request_id = %v, want non-empty string", entry["request_id"])
	}
}
