# 1. Builder 阶段：运行在 Mac 原生环境 (arm64)
FROM --platform=$BUILDPLATFORM golang:1.24-bookworm AS builder

# 声明自动注入的变量
ARG TARGETOS
ARG TARGETARCH

# 修正：Debian 使用 apt-get 而不是 apk
RUN apt-get update && apt-get install -y \
    git \
    protobuf-compiler \
    libprotobuf-dev \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# 设置代理（如果在国内构建建议保留）
ENV GOPROXY=https://goproxy.cn,direct

# 复制依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . .

# 安装工具并生成代码
# 这些工具会在原生 arm64 下运行，速度极快
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest && \
    protoc --go_out=. --go-grpc_out=. api/proto/plugin.proto || true

# 关键：交叉编译
# GOARCH 会根据你构建时的 --platform 自动变为 amd64 或 arm64
# 注意：确保 ./cmd/bot 是你的入口路径
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -ldflags="-s -w" -o bot ./cmd/main.go

# --- 2. 运行阶段 ---
# 这里不需要 --platform，它会自动匹配构建目标
FROM alpine:3.19

# 安装运行时基础包
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# 从 builder 拷贝产物
COPY --from=builder /app/bot .
# 如果你有默认配置文件，取消下面这行的注释
# COPY --from=builder /app/config.yaml .

RUN mkdir -p /app/plugins-bin /app/plugins-config
EXPOSE 8080 50051
VOLUME ["/app/plugins-bin", "/app/plugins-config"]

ENTRYPOINT ["./bot"]