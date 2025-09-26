# Build stage
FROM golang:1.25.1-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Install git for fetching dependencies
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates

# Create a non-root user
RUN adduser -D -g '' appuser

# Set the working directory
WORKDIR /build

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy the source code
COPY . .

# Build the application
RUN go build -ldflags="-w -s" -o main cmd/server/main.go

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates wget

# Create a non-root user
RUN adduser -D -g '' appuser

# Set the working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/main .

# Copy migrations and other necessary files
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/openapi.yml .

# Change ownership of the app directory
RUN chown -R appuser:appuser /app

# Use the non-root user
USER appuser

# Expose port 8080
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Command to run the application
CMD ["./main"]
