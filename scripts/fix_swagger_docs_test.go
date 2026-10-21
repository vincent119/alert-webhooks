package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateSwaggerDocsAcceptsValidFile(t *testing.T) {
	path := writeSwaggerValidationFile(t, "package docs\nvar value = struct { SwaggerTemplate string }{\nSwaggerTemplate: `valid`,\n}\n")
	if err := validateSwaggerDocs(path); err != nil {
		t.Fatalf("有效 Swagger 文件驗證失敗：%v", err)
	}
}

func TestValidateSwaggerDocsRejectsKnownInvalidPatterns(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{name: "left delimiter", content: "LeftDelim: \"{{\",\n"},
		{name: "right delimiter", content: "RightDelim: \"}}\",\n"},
		{name: "double comma", content: "Value: 1,,\n"},
		{name: "missing trailing comma", content: "SwaggerTemplate: `value`\n"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if err := validateSwaggerDocs(writeSwaggerValidationFile(t, testCase.content)); err == nil {
				t.Fatal("已知無效 Swagger pattern 應驗證失敗")
			}
		})
	}
}

func TestValidateSwaggerDocsReturnsOpenAndScannerErrors(t *testing.T) {
	if err := validateSwaggerDocs(filepath.Join(t.TempDir(), "missing.go")); err == nil {
		t.Fatal("不存在的 Swagger 文件應回傳錯誤")
	}

	tooLongLine := strings.Repeat("a", 70*1024)
	if err := validateSwaggerDocs(writeSwaggerValidationFile(t, tooLongLine)); err == nil {
		t.Fatal("超過 Scanner token 限制應回傳錯誤")
	}
}

func TestCloseSwaggerValidationFilePreservesPrimaryValidationError(t *testing.T) {
	validationErr := errors.New("validation failed")
	closeErr := errors.New("close failed")
	file := &swaggerValidationCloserForTest{closeErr: closeErr}

	err := closeSwaggerValidationFile(file, validationErr)
	if !errors.Is(err, validationErr) {
		t.Fatalf("primary validation error was not preserved: %v", err)
	}
	if errors.Is(err, closeErr) {
		t.Fatalf("close error must not replace the primary validation error: %v", err)
	}
	if got, want := file.closeCount, 1; got != want {
		t.Fatalf("validation file close count is incorrect: got %d, want %d", got, want)
	}
}

func TestCloseSwaggerValidationFileReturnsCloseErrorWithoutPrimaryError(t *testing.T) {
	closeErr := errors.New("close failed")
	file := &swaggerValidationCloserForTest{closeErr: closeErr}

	err := closeSwaggerValidationFile(file, nil)
	if !errors.Is(err, closeErr) {
		t.Fatalf("want close error: %v", err)
	}
	if got, want := file.closeCount, 1; got != want {
		t.Fatalf("validation file close count is incorrect: got %d, want %d", got, want)
	}
}

func TestRemoveSwaggerBackupTreatsMissingFileAsSuccess(t *testing.T) {
	backupFile := filepath.Join(t.TempDir(), "docs.go.backup")

	if err := removeSwaggerBackup(backupFile); err != nil {
		t.Fatalf("missing backup file should be treated as success: %v", err)
	}
}

func TestRemoveSwaggerBackupReturnsCleanupError(t *testing.T) {
	backupDir := filepath.Join(t.TempDir(), "docs.go.backup")
	if err := os.Mkdir(backupDir, 0755); err != nil {
		t.Fatalf("create temporary directory failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "keep"), []byte("test"), 0644); err != nil {
		t.Fatalf("create temporary file failed: %v", err)
	}

	if err := removeSwaggerBackup(backupDir); err == nil {
		t.Fatal("expected cleanup error when deleting a non-empty directory")
	}
}

type swaggerValidationCloserForTest struct {
	closeErr   error
	closeCount int
}

func (f *swaggerValidationCloserForTest) Close() error {
	f.closeCount++
	return f.closeErr
}

func writeSwaggerValidationFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "docs.go")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("建立 Swagger 驗證檔案失敗：%v", err)
	}
	return path
}
