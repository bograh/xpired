# Build stage
FROM golang:1.25.1-alpine AS builder
WORKDIR /build
COPY go.* ./
RUN go mod download
COPY . .
RUN go build -ldflags="-w -s" -o main cmd/server/main.go

# Production stage
FROM alpine:latest
RUN apk add --no-cache ca-certificates wget && \
    adduser -D appuser
WORKDIR /app
COPY --from=builder /build/main .
COPY --from=builder /build/migrations ./migrations
COPY --from=builder /build/openapi.yml .
RUN chown -R appuser:appuser /app
USER appuser
EXPOSE 8080
HEALTHCHECK CMD wget -q --spider http://localhost:8080/health || exit 1
CMD ["./main"]