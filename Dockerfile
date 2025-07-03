# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application (production version)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o crypto-signal-bot main.go

# Final stage
FROM alpine:latest

# Install ca-certificates and wget for health checks
RUN apk --no-cache add ca-certificates tzdata wget

# Set timezone and network settings
ENV TZ=Asia/Jakarta
ENV GODEBUG=netdns=go+1

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/crypto-signal-bot .

# Copy any additional files if needed
COPY --from=builder /app/supabase_schema.sql ./

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port (Railway will set PORT env var)
EXPOSE 8080

# No health check needed for simple personal bot

# Run the application
CMD ["./crypto-signal-bot"]
