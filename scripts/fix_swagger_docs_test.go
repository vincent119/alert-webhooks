package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCloseSwaggerValidationFilePreservesPrimaryValidationError(t *testing.T) {
	validationErr := errors.New("validation failed")
	closeErr := errors.New("close failed")
	file := &swaggerValidationCloserForTest{closeErr: closeErr}

	err := closeSwaggerValidationFile(file, validationErr)
	if !errors.Is(err, validationErr) {
		t.Fatalf("主要 validation error 未保留：%v", err)
	}
	if errors.Is(err, closeErr) {
		t.Fatalf("close error 不應覆蓋主要 validation error：%v", err)
	}
	if got, want := file.closeCount, 1; got != want {
		t.Fatalf("驗證檔關閉次數不正確：got %d, want %d", got, want)
	}
}

func TestCloseSwaggerValidationFileReturnsCloseErrorWithoutPrimaryError(t *testing.T) {
	closeErr := errors.New("close failed")
	file := &swaggerValidationCloserForTest{closeErr: closeErr}

	err := closeSwaggerValidationFile(file, nil)
	if !errors.Is(err, closeErr) {
		t.Fatalf("預期回傳 close error：%v", err)
	}
	if got, want := file.closeCount, 1; got != want {
		t.Fatalf("驗證檔關閉次數不正確：got %d, want %d", got, want)
	}
}

func TestRemoveSwaggerBackupTreatsMissingFileAsSuccess(t *testing.T) {
	backupFile := filepath.Join(t.TempDir(), "docs.go.backup")

	if err := removeSwaggerBackup(backupFile); err != nil {
		t.Fatalf("不存在的備份檔應視為成功：%v", err)
	}
}

func TestRemoveSwaggerBackupReturnsCleanupError(t *testing.T) {
	backupDir := filepath.Join(t.TempDir(), "docs.go.backup")
	if err := os.Mkdir(backupDir, 0755); err != nil {
		t.Fatalf("建立暫存目錄失敗：%v", err)
	}
	if err := os.WriteFile(filepath.Join(backupDir, "keep"), []byte("test"), 0644); err != nil {
		t.Fatalf("建立暫存檔失敗：%v", err)
	}

	if err := removeSwaggerBackup(backupDir); err == nil {
		t.Fatal("刪除非空目錄時預期 cleanup error")
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
