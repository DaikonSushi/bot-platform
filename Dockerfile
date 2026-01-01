# Build stage
FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git protobuf-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Generate proto files
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    protoc --go_out=. --go-grpc_out=. api/proto/plugin.proto || true

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary
COPY --from=builder /app/bot .
COPY --from=builder /app/config.yaml .

# Create directories for plugins
RUN mkdir -p /app/plugins-bin /app/plugins-config

# Expose ports
# 8080 - Admin API
# 50051 - gRPC for plugins
EXPOSE 8080 50051

# Volume for persistent data
VOLUME ["/app/plugins-bin", "/app/plugins-config", "/app/config.yaml"]

ENTRYPOINT ["./bot"]
