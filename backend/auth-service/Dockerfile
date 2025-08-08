# Multi-stage Dockerfile for auth-service
FROM golang:1.21-alpine AS builder

# Set working directory
WORKDIR /app

# Install git and ca-certificates for fetching dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/main

# Final stage - minimal image
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create app directory and non-root user
WORKDIR /root/
RUN adduser -D -s /bin/sh appuser

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Change ownership to non-root user
RUN chown appuser:appuser main
USER appuser

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Run the binary
CMD ["./main"]