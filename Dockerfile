FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Set build environment
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o alert-webhooks \
    ./cmd/main.go

# Verify the binary
RUN file alert-webhooks && ls -la alert-webhooks

# -----------------------------------------------------------------------------
# Stage 2: Runtime stage
# -----------------------------------------------------------------------------
FROM alpine:3.19 AS runtime

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    curl \
    && update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1500 -S appgroup && \
    adduser -u 1500 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p /app/configs /app/templates /app/logs /app/documentation && \
    chown -R appuser:appgroup /app

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appgroup /app/alert-webhooks /app/alert-webhooks

# Copy configuration files
COPY --chown=appuser:appgroup configs/ /app/configs/
COPY --chown=appuser:appgroup templates/ /app/templates/

# Copy Swagger documentation (if exists)
COPY --chown=appuser:appgroup docs/ /app/docs/

# Set proper permissions
RUN chmod +x /app/alert-webhooks

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 9999

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9999/healthz || exit 1

# Set default environment variables
ENV GIN_MODE=release \
    APP_ENV=production \
    LOG_LEVEL=info

# Default command
CMD ["/app/alert-webhooks"]

# -----------------------------------------------------------------------------
# Labels for metadata
# -----------------------------------------------------------------------------
LABEL maintainer="alert-webhooks team" \
      version="1.0.0" \
      description="Alert Webhooks service for Telegram and Slack notifications" \
      org.opencontainers.image.title="Alert-Webhooks" \
      org.opencontainers.image.description="Alert notification service supporting Telegram and Slack" \
      org.opencontainers.image.vendor="alert-webhooks" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/vincent119/alert-webhooks"
