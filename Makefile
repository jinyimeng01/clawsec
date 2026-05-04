# ClawSec Makefile

BINARY_NAME=clawsec
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:ssZ'")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
GIT_BRANCH=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "dev")

LDFLAGS=-s -w \
	-X github.com/clawsec/clawsec/internal/constants.Version=$(VERSION) \
	-X github.com/clawsec/clawsec/internal/constants.BuildTime=$(BUILD_TIME) \
	-X github.com/clawsec/clawsec/internal/constants.GitCommit=$(GIT_COMMIT) \
	-X github.com/clawsec/clawsec/internal/constants.GitBranch=$(GIT_BRANCH)

BUILD_DIR=bin

.PHONY: all build clean test lint fmt install cross-compile help

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/clawsec

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean -cache

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo "Coverage report generated: coverage.out"

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "golangci-lint not installed"

fmt:
	@echo "Formatting code..."
	go fmt ./...

install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/ || cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/go/bin/ || echo "Please add $(BUILD_DIR) to your PATH"

cross-compile:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/clawsec
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/clawsec
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/clawsec
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/clawsec
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/clawsec
	@echo "Cross-compilation complete!"

dev-run: build
	./$(BUILD_DIR)/$(BINARY_NAME) version

dev-help: build
	./$(BUILD_DIR)/$(BINARY_NAME) --help

help:
	@echo "ClawSec Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build binary for current platform"
	@echo "  clean          - Remove build artifacts"
	@echo "  test           - Run all tests with coverage"
	@echo "  lint           - Run code linter"
	@echo "  fmt            - Format Go source code"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  cross-compile  - Build for all supported platforms"
	@echo "  dev-run        - Build and run version command"
	@echo "  dev-help       - Build and show help"
	@echo "  help           - Show this help message"
