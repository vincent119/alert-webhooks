package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSwaggerFixMainRepairsAndValidatesDocument(t *testing.T) {
	workingDir := useSwaggerFixWorkingDirectory(t)
	docsDir := filepath.Join(workingDir, "docs")
	if err := os.MkdirAll(docsDir, 0o700); err != nil {
		t.Fatalf("建立 temporary docs 目錄失敗：%v", err)
	}
	docsFile := filepath.Join(docsDir, "docs.go")
	content := `package docs

import "github.com/swaggo/swag"

var SwaggerInfo = &swag.Spec{
	Version: "1.0",,
	SwaggerTemplate: ` + "`valid`" + `,
	LeftDelim: "{{",
	RightDelim: "}}",
}
`
	if err := os.WriteFile(docsFile, []byte(content), 0o600); err != nil {
		t.Fatalf("建立 temporary Swagger 文件失敗：%v", err)
	}

	output := captureSwaggerFixOutput(t, main)
	for _, expected := range []string{
		"移除 LeftDelim 欄位",
		"移除 RightDelim 欄位",
		"清理多餘逗號",
		"Swagger 文檔修復完成",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("Swagger fix output 未包含 %q：%s", expected, output)
		}
	}

	updated, err := os.ReadFile(docsFile)
	if err != nil {
		t.Fatalf("讀取修復後 Swagger 文件失敗：%v", err)
	}
	updatedText := string(updated)
	if strings.Contains(updatedText, "LeftDelim") || strings.Contains(updatedText, "RightDelim") || strings.Contains(updatedText, ",,") {
		t.Fatalf("Swagger 文件仍包含已知問題：%s", updatedText)
	}
	if !strings.Contains(updatedText, "SwaggerTemplate: `valid`,") {
		t.Fatalf("SwaggerTemplate 未補上尾隨逗號：%s", updatedText)
	}
	if _, err := os.Stat(docsFile + ".backup"); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("成功後 backup 應移除：%v", err)
	}
}

func TestSwaggerFixMainReturnsWithoutChangingValidDocument(t *testing.T) {
	workingDir := useSwaggerFixWorkingDirectory(t)
	docsDir := filepath.Join(workingDir, "docs")
	if err := os.MkdirAll(docsDir, 0o700); err != nil {
		t.Fatalf("建立 temporary docs 目錄失敗：%v", err)
	}
	docsFile := filepath.Join(docsDir, "docs.go")
	content := "package docs\nvar value = struct { SwaggerTemplate string }{\nSwaggerTemplate: `valid`,\n}\n"
	if err := os.WriteFile(docsFile, []byte(content), 0o600); err != nil {
		t.Fatalf("建立有效 Swagger 文件失敗：%v", err)
	}

	output := captureSwaggerFixOutput(t, main)
	if !strings.Contains(output, "文檔已經是正確的，無需修復") {
		t.Fatalf("no-op output 不符預期：%s", output)
	}
	updated, err := os.ReadFile(docsFile)
	if err != nil {
		t.Fatalf("讀取 no-op Swagger 文件失敗：%v", err)
	}
	if string(updated) != content {
		t.Fatalf("有效 Swagger 文件不應被修改：%s", updated)
	}
}

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

func useSwaggerFixWorkingDirectory(t *testing.T) string {
	t.Helper()
	workingDir := t.TempDir()
	previousDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("取得工作目錄失敗：%v", err)
	}
	if err := os.Chdir(workingDir); err != nil {
		t.Fatalf("切換 temporary directory 失敗：%v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(previousDir); err != nil {
			t.Errorf("還原工作目錄失敗：%v", err)
		}
	})
	return workingDir
}

func captureSwaggerFixOutput(t *testing.T, run func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("建立 stdout pipe 失敗：%v", err)
	}
	previousStdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = previousStdout })

	run()
	if err := writer.Close(); err != nil {
		t.Fatalf("關閉 stdout writer 失敗：%v", err)
	}
	os.Stdout = previousStdout

	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("讀取 stdout 失敗：%v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("關閉 stdout reader 失敗：%v", err)
	}
	return string(output)
}
