# Swagger 文檔故障排除指南

本指南說明如何解決 Swagger 文檔生成和編譯過程中的常見問題。

## 🚨 常見問題

### 1. `LeftDelim` / `RightDelim` 編譯錯誤

**錯誤訊息:**
```
docs/docs.go:787:2: unknown field LeftDelim in struct literal of type "github.com/swaggo/swag".Spec
docs/docs.go:788:2: unknown field RightDelim in struct literal of type "github.com/swaggo/swag".Spec
```

**原因:** 新版本的 `swaggo/swag` 不再支援 `LeftDelim` 和 `RightDelim` 欄位。

**解決方案:**

#### 方法 1: 使用自動修復 (推薦)
```bash
# 使用 Make 命令自動修復
make swagger-fix

# 或直接運行修復腳本
go run scripts/fix_swagger_docs.go
```

#### 方法 2: 重新生成並自動修復
```bash
# 清理並重新生成（會自動應用修復）
make swagger-clean
make swagger-generate
```

#### 方法 3: 手動修復
編輯 `docs/docs.go` 文件，移除以下行：
```go
LeftDelim:  "{{",
RightDelim: "}}",
```

### 2. `swag init` 工具崩潰

**錯誤訊息:**
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**解決方案:**

#### 步驟 1: 升級 swag 工具
```bash
make upgrade-swag
```

#### 步驟 2: 使用現有文檔
如果工具持續崩潰，可以使用已驗證的手動文檔：
```bash
# 檢查現有文檔是否完整
ls -la docs/

# 如果文檔存在，嘗試修復
make swagger-fix
```

### 3. Go 1.25.0 編譯警告

**警告訊息:**
```
warning: failed to evaluate const mProfCycleWrap at .../runtime/mprof.go:179:7
```

**解決方案:**
這是 Go 1.25.0 的已知問題，不影響程序運行，但建議使用穩定版本：

#### 方法 1: 降級 Go 版本 (推薦)
```bash
# 在 go.mod 中使用
go 1.23
```

#### 方法 2: 忽略警告
警告不會影響實際功能，可以安全忽略。

## 🛠️ 預防措施

### 1. 自動化修復流程

我們已經建立了自動化流程來預防這些問題：

```bash
# Makefile 中的 swagger-generate 會自動應用修復
make swagger-generate

# Shell 腳本也包含自動修復
./scripts/regenerate_swagger.sh
```

### 2. CI/CD 整合

GitHub Actions 工作流程會自動檢查和修復 Swagger 文檔：
- `.github/workflows/swagger-check.yml`

### 3. 版本控制策略

建議的 `go.mod` 配置：
```go
go 1.23  // 使用穩定版本

require (
    github.com/swaggo/swag v1.8.12
    // 其他依賴...
)
```

## 🔧 可用命令

### Make 命令
```bash
make swagger-generate    # 生成並自動修復
make swagger-clean      # 清理文檔
make swagger-fix        # 僅修復現有文檔
make swagger-manual     # 手動生成（備用）
make upgrade-swag       # 升級 swag 工具
```

### 直接命令
```bash
# 生成 Swagger 文檔
swag init -g cmd/main.go --output docs --parseDependency --parseInternal

# 修復生成的文檔
go run scripts/fix_swagger_docs.go

# 驗證修復效果
go build -o /tmp/test ./cmd/main.go
```

## 📋 檢查清單

在部署前，請確認：

- [ ] Swagger 文檔生成成功
- [ ] 修復腳本執行無錯誤
- [ ] 應用程式可以正常編譯
- [ ] Swagger UI 可以正常訪問
- [ ] API 端點回應正確

### 驗證步驟

```bash
# 1. 清理並重新生成
make swagger-clean
make swagger-generate

# 2. 驗證編譯
go build ./cmd/main.go

# 3. 啟動服務
go run cmd/main.go -e development

# 4. 檢查 Swagger UI
curl -f http://localhost:9999/swagger/index.html
```

## 🆘 緊急修復

如果所有自動化方法都失敗，可以使用以下緊急步驟：

### 1. 恢復到最後已知良好狀態
```bash
git checkout HEAD~1 -- docs/
```

### 2. 手動創建最小化文檔
```bash
# 創建基本的 docs.go
cat > docs/docs.go << 'EOF'
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "swagger": "2.0",
    "info": {
        "title": "Alert Webhooks API",
        "version": "1.0"
    }
}`

var SwaggerInfo = &swag.Spec{
    Version:          "1.0",
    Host:             "",
    BasePath:         "/api/v1",
    Schemes:          []string{"http", "https"},
    Title:            "Alert Webhooks API",
    Description:      "Alert notification service API",
    InfoInstanceName: "swagger",
    SwaggerTemplate:  docTemplate,
}

func init() {
    swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
EOF
```

### 3. 創建基本 JSON 文檔
```bash
cat > docs/swagger.json << 'EOF'
{
    "swagger": "2.0",
    "info": {
        "title": "Alert Webhooks API",
        "version": "1.0"
    },
    "basePath": "/api/v1",
    "schemes": ["http", "https"]
}
EOF
```

## 📞 支援

如果問題持續存在：

1. 檢查 GitHub Issues
2. 更新到最新版本的依賴
3. 查看詳細的錯誤日誌
4. 考慮使用手動文檔作為臨時解決方案

## 📚 相關資源

- [Swag 官方文檔](https://github.com/swaggo/swag)
- [Gin-Swagger 文檔](https://github.com/swaggo/gin-swagger)
- [Go 版本相容性](https://golang.org/doc/devel/release.html)
