.PHONY: all build clean proto docker test

# Variables
BINARY_NAME=bot
CTL_NAME=botctl
PLUGIN_DIR=plugins-bin
CONFIG_DIR=plugins-config

all: proto build build-ctl

# Generate protobuf code
proto:
	@echo "Generating protobuf code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/plugin.proto

# Build the main binary
build:
	@echo "Building bot platform..."
	go build -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/main.go

# Build the CLI tool
build-ctl:
	@echo "Building botctl..."
	go build -ldflags="-s -w" -o $(CTL_NAME) ./cmd/botctl/main.go

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)_linux_amd64 ./cmd/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)_linux_arm64 ./cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(BINARY_NAME)_darwin_amd64 ./cmd/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(BINARY_NAME)_darwin_arm64 ./cmd/main.go
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(CTL_NAME)_linux_amd64 ./cmd/botctl/main.go
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(CTL_NAME)_linux_arm64 ./cmd/botctl/main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o $(CTL_NAME)_darwin_amd64 ./cmd/botctl/main.go
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o $(CTL_NAME)_darwin_arm64 ./cmd/botctl/main.go

# Build example plugins
build-plugins:
	@echo "Building example plugins..."
	cd examples/plugin-weather && go build -ldflags="-s -w" -o weather-plugin .

# Run the bot
run: build
	./$(BINARY_NAME)

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)_* $(CTL_NAME) $(CTL_NAME)_*
	rm -f examples/plugin-weather/weather-plugin*
	rm -f examples/plugin-echo-external/echo-ext-plugin*

# Clean all including downloaded plugins
clean-all: clean
	find $(PLUGIN_DIR) -type f ! -name '.gitkeep' -delete 2>/dev/null || true

# Setup directories
setup:
	mkdir -p $(PLUGIN_DIR) $(CONFIG_DIR)

# Docker build
docker:
	docker build -t bot-platform:latest .

# Docker compose up
up:
	docker-compose up -d

# Docker compose down
down:
	docker-compose down

# Run tests
test:
	go test -v ./...

# Install protoc plugins
install-proto-tools:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Download dependencies
deps:
	go mod download
	go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  all           - Generate proto and build all binaries"
	@echo "  proto         - Generate protobuf code"
	@echo "  build         - Build the main bot binary"
	@echo "  build-ctl     - Build the botctl CLI tool"
	@echo "  build-all     - Build for all platforms"
	@echo "  build-plugins - Build example plugins"
	@echo "  run           - Build and run the bot"
	@echo "  clean         - Remove build artifacts"
	@echo "  setup         - Create required directories"
	@echo "  docker        - Build Docker image"
	@echo "  up            - Start with docker-compose"
	@echo "  down          - Stop docker-compose"
	@echo "  test          - Run tests"
	@echo "  deps          - Download and tidy dependencies"
