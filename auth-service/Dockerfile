# Multi-stage build for optimized production image

# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies and security updates
RUN apk add --no-cache git ca-certificates tzdata && \
    apk upgrade --no-cache

# Create non-root user for build
RUN addgroup -g 1000 appgroup && \
    adduser -D -s /bin/sh -u 1000 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o auth-service \
    ./cmd/http/main.go

# Test stage (optional - for running tests in CI)
FROM builder AS tester
COPY . .
RUN go test -v -race -coverprofile=coverage.out ./...

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies and security updates
RUN apk --no-cache add ca-certificates tzdata curl && \
    apk upgrade --no-cache && \
    rm -rf /var/cache/apk/*

# Create non-root user for runtime
RUN addgroup -g 1000 appgroup && \
    adduser -D -s /bin/sh -u 1000 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/auth-service .

# Copy configuration and migration files
COPY --from=builder /app/config/ ./config/
COPY --from=builder /app/migrations/ ./migrations/

# Create directories for logs and temporary files
RUN mkdir -p /app/logs /tmp/auth-service && \
    chown -R appuser:appgroup /app /tmp/auth-service

# Set proper permissions
RUN chmod +x ./auth-service && \
    chown appuser:appgroup ./auth-service

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Enhanced health check with proper timeouts and retries
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f --max-time 5 http://localhost:8080/health || exit 1

# Set environment variables
ENV APP_ENV=production \
    LOG_LEVEL=info \
    GIN_MODE=release

# Add labels for better metadata
LABEL maintainer="DevOps Team <devops@company.com>" \
    version="1.0.0" \
    description="Auth Service - Production Optimized" \
    security.scan="enabled"

# Graceful shutdown
STOPSIGNAL SIGTERM

# Run the application
CMD ["./auth-service"]