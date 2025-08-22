#!/bin/bash

# Alert Webhooks - Swagger documentation regeneration script
# Usage: ./scripts/regenerate_swagger.sh

set -e

echo "🔄 Regenerating Swagger documentation..."

# Ensure docs directory exists
mkdir -p docs

# Try generating with swag init
echo "📝 Attempting to generate with swag init..."
if swag init -g cmd/main.go --output docs --parseDependency --parseInternal 2>/dev/null; then
    echo "✅ Successfully generated documentation with swag init"
    
    # 自動修復已知的 Swagger 問題
    echo "🔧 Applying automatic fixes to generated documentation..."
    if go run scripts/fix_swagger_docs.go; then
        echo "✅ Swagger documentation fixes applied successfully"
    else
        echo "⚠️  Failed to apply fixes, but documentation may still work"
    fi
else
    echo "⚠️  swag init failed, keeping existing manually generated documentation"
fi

# Check if required files exist
required_files=("docs/docs.go" "docs/swagger.json" "docs/swagger.yaml")
missing_files=()

for file in "${required_files[@]}"; do
    if [[ ! -f "$file" ]]; then
        missing_files+=("$file")
    fi
done

if [[ ${#missing_files[@]} -gt 0 ]]; then
    echo "❌ Missing required Swagger documentation files:"
    printf '   - %s\n' "${missing_files[@]}"
    echo "💡 Please ensure manual documentation is complete"
    exit 1
fi

echo "✅ Swagger documentation check completed"

# Verify service is running
if curl -s http://localhost:9999/swagger/doc.json >/dev/null 2>&1; then
    echo "🌐 Swagger API available at: http://localhost:9999/swagger/index.html"
else
    echo "⚠️  Service is not running. Please start the application with:"
    echo "   go run cmd/main.go -e development"
fi

echo "📋 Regeneration process finished!"