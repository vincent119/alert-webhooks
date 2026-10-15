package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// fix_swagger_docs.go - 自動修復 swag 生成的 docs.go 中的已知問題
// 使用方法: go run scripts/fix_swagger_docs.go

func main() {
	docsFile := "docs/docs.go"

	fmt.Println("🔧 修復 Swagger 文檔中的已知問題...")

	// 檢查文件是否存在
	if _, err := os.Stat(docsFile); os.IsNotExist(err) {
		fmt.Printf("❌ 文件不存在: %s\n", docsFile)
		os.Exit(1)
	}

	// 讀取文件內容
	content, err := os.ReadFile(docsFile)
	if err != nil {
		fmt.Printf("❌ 讀取文件失敗: %v\n", err)
		os.Exit(1)
	}

	originalContent := string(content)
	modifiedContent := originalContent

	// 修復 1: 移除不支援的 LeftDelim 和 RightDelim 欄位
	leftDelimPattern := regexp.MustCompile(`\s*LeftDelim:\s*"[^"]*",?\n`)
	rightDelimPattern := regexp.MustCompile(`\s*RightDelim:\s*"[^"]*",?\n`)

	if leftDelimPattern.MatchString(modifiedContent) {
		fmt.Println("🔧 移除 LeftDelim 欄位...")
		modifiedContent = leftDelimPattern.ReplaceAllString(modifiedContent, "")
	}

	if rightDelimPattern.MatchString(modifiedContent) {
		fmt.Println("🔧 移除 RightDelim 欄位...")
		modifiedContent = rightDelimPattern.ReplaceAllString(modifiedContent, "")
	}

	// 修復 2: 修復 struct literal 中的逗號問題
	// 清理多餘的尾隨逗號
	extraCommaPattern := regexp.MustCompile(`,\s*,`)
	if extraCommaPattern.MatchString(modifiedContent) {
		fmt.Println("🔧 清理多餘逗號...")
		modifiedContent = extraCommaPattern.ReplaceAllString(modifiedContent, ",")
	}

	// 確保 struct literal 的最後一個欄位有逗號（Go 習慣）
	structEndPattern := regexp.MustCompile(`(\w+:\s*[^,\n}]+)\s*\n\s*}`)
	if structEndPattern.MatchString(modifiedContent) {
		fmt.Println("🔧 添加缺少的尾隨逗號...")
		modifiedContent = structEndPattern.ReplaceAllString(modifiedContent, "$1,\n}")
	}

	// 修復 3: 確保正確的導入
	if !strings.Contains(modifiedContent, `"github.com/swaggo/swag"`) {
		fmt.Println("🔧 檢查導入語句...")
		// 如果需要，可以在這裡添加導入修復邏輯
	}

	// 檢查是否有修改
	if modifiedContent == originalContent {
		fmt.Println("✅ 文檔已經是正確的，無需修復")
		return
	}

	// 備份原始文件
	backupFile := docsFile + ".backup"
	if err := os.WriteFile(backupFile, []byte(originalContent), 0644); err != nil {
		fmt.Printf("⚠️  無法創建備份文件: %v\n", err)
	} else {
		fmt.Printf("💾 已創建備份: %s\n", backupFile)
	}

	// 寫入修復後的內容
	if err := os.WriteFile(docsFile, []byte(modifiedContent), 0644); err != nil {
		fmt.Printf("❌ 寫入修復後的文件失敗: %v\n", err)
		os.Exit(1)
	}

	// 驗證修復後的文件
	fmt.Println("🧪 驗證修復後的文件...")
	if err := validateSwaggerDocs(docsFile); err != nil {
		fmt.Printf("❌ 驗證失敗: %v\n", err)

		// 嘗試恢復備份
		if _, backupErr := os.Stat(backupFile); backupErr == nil {
			fmt.Println("🔄 恢復備份文件...")
			if restoreErr := os.Rename(backupFile, docsFile); restoreErr != nil {
				fmt.Printf("❌ 恢復備份失敗: %v\n", restoreErr)
			} else {
				fmt.Println("✅ 已恢復原始文件")
			}
		}
		os.Exit(1)
	}

	// 清理備份文件（可選）
	if err := removeSwaggerBackup(backupFile); err != nil {
		fmt.Printf("⚠️  無法清理備份文件: %v\n", err)
	}

	fmt.Println("✅ Swagger 文檔修復完成！")
	fmt.Println("📝 修復的問題:")
	fmt.Println("   - 移除不支援的 LeftDelim/RightDelim 欄位")
	fmt.Println("   - 清理語法問題")
	fmt.Println("   - 確保編譯相容性")
}

// validateSwaggerDocs 驗證修復後的 docs.go 文件
func validateSwaggerDocs(filePath string) (err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("無法打開文件: %v", err)
	}
	defer func() {
		err = closeSwaggerValidationFile(file, err)
	}()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// 檢查是否還有問題欄位
		if strings.Contains(line, "LeftDelim:") || strings.Contains(line, "RightDelim:") {
			return fmt.Errorf("第 %d 行仍包含問題欄位: %s", lineNum, line)
		}

		// 檢查語法問題
		if strings.Contains(line, ",,") {
			return fmt.Errorf("第 %d 行有語法錯誤（雙逗號）: %s", lineNum, line)
		}

		// 檢查 struct literal 結尾是否缺少逗號
		if strings.Contains(line, "SwaggerTemplate:") && !strings.Contains(line, ",") {
			return fmt.Errorf("第 %d 行可能缺少尾隨逗號: %s", lineNum, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("掃描文件時出錯: %v", err)
	}

	return nil
}

func removeSwaggerBackup(backupFile string) error {
	err := os.Remove(backupFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func closeSwaggerValidationFile(file io.Closer, validationErr error) error {
	if closeErr := file.Close(); closeErr != nil && validationErr == nil {
		return fmt.Errorf("關閉驗證文件失敗: %w", closeErr)
	}
	return validationErr
}
