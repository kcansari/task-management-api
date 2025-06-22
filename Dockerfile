# Multi-stage Docker build for Go application
# This pattern reduces final image size by separating build dependencies from runtime

# Build stage - includes Go compiler and development tools
FROM golang:1.23.4-alpine AS builder

# Install build dependencies
# git: for fetching Go modules from version control
# ca-certificates: for HTTPS requests during module download
RUN apk add --no-cache git ca-certificates

# Set working directory in container
WORKDIR /app

# Copy Go module files first (for better Docker layer caching)
# This allows Docker to cache the module download step if go.mod/go.sum haven't changed
COPY go.mod go.sum ./

# Download dependencies
# This step will be cached if go.mod and go.sum haven't changed
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0: Disable CGO for static binary (better for containers)
# GOOS=linux: Target Linux OS (important for cross-compilation)
# -a: Force rebuilding of packages
# -installsuffix cgo: Use different install suffix when CGO is disabled
# -ldflags '-extldflags "-static"': Create fully static binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o task-api .

# Production stage - minimal runtime environment
FROM alpine:latest

# Install runtime dependencies
# ca-certificates: for HTTPS requests (if app makes external calls)
# tzdata: for proper timezone handling
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
# Running as non-root is a security best practice
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/task-api .

# Copy configuration files if needed
COPY --from=builder /app/.env.example .env.example

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (documented - doesn't actually publish)
EXPOSE 8080

# Health check to ensure container is working
# Docker will periodically run this command to check container health
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./task-api"]