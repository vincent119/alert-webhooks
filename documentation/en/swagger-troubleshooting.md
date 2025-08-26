# Swagger Documentation Troubleshooting Guide

This guide explains how to resolve common issues during Swagger documentation generation and compilation.

## 🚨 Common Issues

### 1. `LeftDelim` / `RightDelim` Compilation Error

**Error Message:**
```
docs/docs.go:787:2: unknown field LeftDelim in struct literal of type "github.com/swaggo/swag".Spec
docs/docs.go:788:2: unknown field RightDelim in struct literal of type "github.com/swaggo/swag".Spec
```

**Cause:** Newer versions of `swaggo/swag` no longer support `LeftDelim` and `RightDelim` fields.

**Solutions:**

#### Method 1: Use Automatic Fix (Recommended)
```bash
# Use Make command to auto-fix
make swagger-fix

# Or run the fix script directly
go run scripts/fix_swagger_docs.go
```

#### Method 2: Regenerate with Auto-fix
```bash
# Clean and regenerate (will auto-apply fixes)
make swagger-clean
make swagger-generate
```

#### Method 3: Manual Fix
Edit `docs/docs.go` file and remove these lines:
```go
LeftDelim:  "{{",
RightDelim: "}}",
```

### 2. `swag init` Tool Crash

**Error Message:**
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**Solutions:**

#### Step 1: Upgrade swag tool
```bash
make upgrade-swag
```

#### Step 2: Use existing documentation
If the tool continues to crash, use verified manual documentation:
```bash
# Check if existing documentation is complete
ls -la docs/

# If documentation exists, try to fix it
make swagger-fix
```

### 3. Go 1.25.0 Compilation Warning

**Warning Message:**
```
warning: failed to evaluate const mProfCycleWrap at .../runtime/mprof.go:179:7
```

**Solution:**
This is a known issue with Go 1.25.0 that doesn't affect program functionality, but we recommend using a stable version:

#### Method 1: Downgrade Go Version (Recommended)
```bash
# Use in go.mod
go 1.23
```

#### Method 2: Ignore Warning
The warning doesn't affect actual functionality and can be safely ignored.

## 🛠️ Prevention Measures

### 1. Automated Fix Process

We've established an automated process to prevent these issues:

```bash
# swagger-generate in Makefile automatically applies fixes
make swagger-generate

# Shell script also includes auto-fix
./scripts/regenerate_swagger.sh
```

### 2. CI/CD Integration

GitHub Actions workflow automatically checks and fixes Swagger documentation:
- `.github/workflows/swagger-check.yml`

### 3. Version Control Strategy

Recommended `go.mod` configuration:
```go
go 1.23  // Use stable version

require (
    github.com/swaggo/swag v1.8.12
    // Other dependencies...
)
```

## 🔧 Available Commands

### Make Commands
```bash
make swagger-generate    # Generate and auto-fix
make swagger-clean      # Clean documentation
make swagger-fix        # Fix existing documentation only
make swagger-manual     # Manual generation (fallback)
make upgrade-swag       # Upgrade swag tool
```

### Direct Commands
```bash
# Generate Swagger documentation
swag init -g cmd/main.go --output docs --parseDependency --parseInternal

# Fix generated documentation
go run scripts/fix_swagger_docs.go

# Verify fix effectiveness
go build -o /tmp/test ./cmd/main.go
```

## 📋 Checklist

Before deployment, please confirm:

- [ ] Swagger documentation generated successfully
- [ ] Fix script executed without errors
- [ ] Application compiles normally
- [ ] Swagger UI accessible
- [ ] API endpoints respond correctly

### Verification Steps

```bash
# 1. Clean and regenerate
make swagger-clean
make swagger-generate

# 2. Verify compilation
go build ./cmd/main.go

# 3. Start service
go run cmd/main.go -e development

# 4. Check Swagger UI
curl -f http://localhost:9999/swagger/index.html
```

## 🆘 Emergency Fix

If all automated methods fail, use these emergency steps:

### 1. Restore to Last Known Good State
```bash
git checkout HEAD~1 -- docs/
```

### 2. Manually Create Minimal Documentation
```bash
# Create basic docs.go
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

### 3. Create Basic JSON Documentation
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

## 📞 Support

If issues persist:

1. Check GitHub Issues
2. Update to latest dependency versions
3. Review detailed error logs
4. Consider using manual documentation as temporary solution

## 📚 Related Resources

- [Swag Official Documentation](https://github.com/swaggo/swag)
- [Gin-Swagger Documentation](https://github.com/swaggo/gin-swagger)
- [Go Version Compatibility](https://golang.org/doc/devel/release.html)
