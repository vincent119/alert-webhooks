SHELL := /bin/bash

# Makefile for Alert Webhooks project

.PHONY: swagger-generate print-swag-dirs swagger-clean swagger-manual run dev test build deps fmt lint upgrade-swag \
        docker-build docker-build-dev docker-run docker-dev docker-stop docker-clean docker-logs docker-logs-dev docker-shell help

# Generate Swagger documentation
SWAG_DIRS := cmd,$(shell \
	find . -name "*.go" -not -path "./vendor/*" -exec dirname {} \; \
	| sed 's|^\./||' | grep -v '^cmd$$' | sort -u | paste -sd "," - \
)


print-swag-dirs:
	@echo "$(SWAG_DIRS)"

swagger-generate:
	@echo "Generating Swagger documentation..."
	@test -n "$(SWAG_DIRS)" || (echo "SWAG_DIRS is empty"; exit 1)
	@swag init -g main.go -o docs --parseDependency --parseInternal -d $(SWAG_DIRS)
	@echo "✅ Swagger generation successful, applying fixes..."
	@go run scripts/fix_swagger_docs.go || echo "⚠️  Fix script failed but documentation may still work"


# Clean Swagger documentation
swagger-clean:
	@echo "Cleaning Swagger documentation..."
	@rm -f docs/docs.go docs/swagger.json docs/swagger.yaml

# Manually regenerate Swagger documentation (use if swag tool fails)
swagger-manual:
	@echo "Manually regenerating Swagger documentation..."
	@go run scripts/generate_swagger.go

# Fix Swagger documentation issues
swagger-fix:
	@echo "Fixing Swagger documentation issues..."
	@go run scripts/fix_swagger_docs.go

# Run in development mode
dev:
	@echo "Starting development environment..."
	@go run cmd/main.go -e development

# Run in production mode
run:
	@echo "Starting production environment..."
	@go run cmd/main.go -e production

# Build project
build:
	@echo "Building project..."
	@go build -o bin/alert-webhooks cmd/main.go

# Run tests
test:
	@echo "Running tests..."
	@go test ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
lint:
	@echo "Running lint checks..."
	@golangci-lint run

# Upgrade swag tool
upgrade-swag:
	@echo "Upgrading swag tool..."
	@go install github.com/swaggo/swag/cmd/swag@latest

# Docker related commands

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t alert-webhooks:latest .

# Build Docker development image
docker-build-dev:
	@echo "Building Docker development image..."
	@docker build -t alert-webhooks:dev --target builder .

# Run Docker container (production mode)
docker-run:
	@echo "Starting Docker container (production mode)..."
	@docker-compose up -d

# Run Docker container (development mode)
docker-dev:
	@echo "Starting Docker container (development mode)..."
	@docker-compose -f docker-compose.dev.yml up -d

# Stop Docker containers
docker-stop:
	@echo "Stopping Docker containers..."
	@docker-compose down
	@docker-compose -f docker-compose.dev.yml down

# Clean Docker resources
docker-clean:
	@echo "Cleaning Docker resources..."
	@docker-compose down -v --rmi all
	@docker-compose -f docker-compose.dev.yml down -v --rmi all
	@docker system prune -f

# View Docker logs
docker-logs:
	@echo "Viewing Docker logs..."
	@docker-compose logs -f

# View development mode logs
docker-logs-dev:
	@echo "Viewing development mode logs..."
	@docker-compose -f docker-compose.dev.yml logs -f

# Enter container shell
docker-shell:
	@echo "Entering container shell..."
	@docker exec -it alert-webhooks /bin/sh

# Show help
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development commands:"
	@echo "  dev              - Start development environment"
	@echo "  run              - Start production environment"
	@echo "  build            - Build project"
	@echo "  test             - Run tests"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-build-dev - Build Docker development image"
	@echo "  docker-run       - Start Docker container (production mode)"
	@echo "  docker-dev       - Start Docker container (development mode)"
	@echo "  docker-stop      - Stop Docker containers"
	@echo "  docker-clean     - Clean Docker resources"
	@echo "  docker-logs      - View Docker logs"
	@echo "  docker-logs-dev  - View development mode logs"
	@echo "  docker-shell     - Enter container shell"
	@echo ""
	@echo "Documentation commands:"
	@echo "  swagger-generate - Generate Swagger documentation"
	@echo "  swagger-clean    - Clean Swagger documentation"
	@echo "  swagger-manual   - Manually regenerate Swagger documentation"
	@echo "  swagger-fix      - Fix Swagger documentation issues"
	@echo "  upgrade-swag     - Upgrade swag tool"
	@echo ""
	@echo "Utility commands:"
	@echo "  deps             - Install dependencies"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run lint checks"