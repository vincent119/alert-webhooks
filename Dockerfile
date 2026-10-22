ARG ALPINE_VERSION=3.24.2
ARG GOLANG_VERSION=1.27.3
ARG GOLANG_ALPINE_VERSION=3.24

# -----------------------------------------------------------------------------
# Stage 1: Builder（使用本機架構編譯，Go cross-compile 產出目標平台 binary）
# -----------------------------------------------------------------------------
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION}-alpine${GOLANG_ALPINE_VERSION}@sha256:d530cf60723abf1ffae0cd16bb546b379ec0550d2f3815e5b16ed69dbc00fba4 AS builder

ARG TARGETOS
ARG TARGETARCH

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    file

ARG timezone=UTC
ENV TZ=${timezone}

# Set working directory
WORKDIR /app

# Copy dependency files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download && go mod verify

# Copy source code
COPY . .

# 依 BuildKit 目標平台交叉編譯，避免 arm64 image 包入 amd64 binary。
RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -trimpath -ldflags="-s -w" -o /build/alert-webhooks ./cmd/main.go

# Verify the binary
RUN file /build/alert-webhooks && ls -la /build/alert-webhooks

# -----------------------------------------------------------------------------
# Stage 2: Runtime（使用目標平台架構）
# -----------------------------------------------------------------------------
FROM alpine:${ALPINE_VERSION}@sha256:294b683cb724975bec92580e1e685676bd4b50bda910ddb8c51d4cabeaec77e6 AS runtime

ARG timezone=UTC
ENV TZ=${timezone}

# Install runtime dependencies
RUN apk add --no-cache \
    "musl>=1.2.6-r3" \
    "musl-utils>=1.2.6-r3" \
    "zlib>=1.3.2-r1" \
    ca-certificates \
    tzdata \
    && update-ca-certificates

# Create non-root user for security
RUN addgroup -g 1500 -S appgroup && \
    adduser -u 1500 -S appuser -G appgroup && \
    ln -snf /usr/share/zoneinfo/"${TZ}" /etc/localtime && echo "${TZ}" > /etc/timezone

# Set working directory
WORKDIR /app

# Create necessary directories
RUN mkdir -p /app/configs /app/templates/alerts /app/logs  && \
    chown -R appuser:appgroup /app

# Copy binary from builder stage
COPY --from=builder --chown=appuser:appgroup /build/alert-webhooks /app/alert-webhooks

# Copy configuration files
COPY --chown=appuser:appgroup configs/ /app/configs/
COPY --chown=appuser:appgroup templates/alerts/ /app/templates/alerts/

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
    CMD wget -qO- http://localhost:9999/healthy >/dev/null || exit 1

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
