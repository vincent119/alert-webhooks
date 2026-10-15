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
