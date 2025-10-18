# TheBoys Launcher Makefile
# Cross-platform build system for the modern TheBoys Launcher

# Version and build information
VERSION := 2.0.0
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GO_VERSION := $(shell go version | awk '{print $$3}')

# Build configuration
BINARY_NAME := theboys-launcher
OUTPUT_DIR := build
PLATFORMS := windows/amd64 linux/amd64 darwin/amd64 darwin/arm64

# LDFlags for build info
LDFLAGS := -ldflags "-X theboys-launcher/pkg/version.Version=$(VERSION) \
                    -X theboys-launcher/pkg/version.BuildTime=$(BUILD_TIME) \
                    -X theboys-launcher/pkg/version.GitCommit=$(GIT_COMMIT) \
                    -X theboys-launcher/pkg/version.GoVersion=$(GO_VERSION) \
                    -s -w"

# Default target
.PHONY: all
all: clean build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(OUTPUT_DIR)
	@rm -f theboys-launcher*
	@go clean -cache
	@echo "Clean completed."

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies installed."

# Build for current platform
.PHONY: build
build:
	@echo "Building for current platform..."
	@mkdir -p $(OUTPUT_DIR)
	@go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME) ./cmd/winterpack
	@echo "Build completed: $(OUTPUT_DIR)/$(BINARY_NAME)"

# Cross-platform build
.PHONY: build-all
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(OUTPUT_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		echo "Building for $$GOOS/$$GOARCH..."; \
		GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) -o $(OUTPUT_DIR)/$$output_name ./cmd/winterpack; \
	done
	@echo "Cross-platform build completed."

# Build for specific platform
.PHONY: build-windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/winterpack
	@echo "Windows build completed: $(OUTPUT_DIR)/$(BINARY_NAME)-windows-amd64.exe"

.PHONY: build-linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/winterpack
	@echo "Linux build completed: $(OUTPUT_DIR)/$(BINARY_NAME)-linux-amd64"

.PHONY: build-darwin
build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(OUTPUT_DIR)
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/winterpack
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(OUTPUT_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/winterpack
	@echo "macOS builds completed"

# Development targets
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@go run ./cmd/winterpack

.PHONY: dev-debug
dev-debug:
	@echo "Running in debug mode with verbose logging..."
	@DEBUG=true go run ./cmd/winterpack

# Testing
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Code quality
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted."

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "Go vet completed."

.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
		echo "Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Security
.PHONY: security
security:
	@echo "Running security checks..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not installed, skipping..."; \
		echo "Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

# Dependencies management
.PHONY: deps-update
deps-update:
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "Dependencies updated."

.PHONY: deps-verify
deps-verify:
	@echo "Verifying dependencies..."
	@go mod verify
	@echo "Dependencies verified."

# Help
.PHONY: help
help:
	@echo "TheBoys Launcher Build System"
	@echo "============================="
	@echo ""
	@echo "Build targets:"
	@echo "  build         Build for current platform"
	@echo "  build-all     Build for all supported platforms"
	@echo "  build-windows Build for Windows (amd64)"
	@echo "  build-linux   Build for Linux (amd64)"
	@echo "  build-darwin  Build for macOS (amd64, arm64)"
	@echo ""
	@echo "Development targets:"
	@echo "  dev           Run in development mode"
	@echo "  dev-debug     Run in debug mode"
	@echo "  test          Run tests"
	@echo "  test-coverage Run tests with coverage report"
	@echo ""
	@echo "Code quality targets:"
	@echo "  fmt           Format code"
	@echo "  vet           Run go vet"
	@echo "  lint          Run linter (requires golangci-lint)"
	@echo "  security      Run security checks (requires gosec)"
	@echo ""
	@echo "Dependency targets:"
	@echo "  deps          Install dependencies"
	@echo "  deps-update   Update dependencies"
	@echo "  deps-verify   Verify dependencies"
	@echo ""
	@echo "Other targets:"
	@echo "  clean         Clean build artifacts"
	@echo "  help          Show this help message"
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
	@echo "Git commit: $(GIT_COMMIT)"
	@echo "Go version: $(GO_VERSION)"

# Release targets
.PHONY: release
release: clean test fmt vet build-all
	@echo "Creating release packages..."
	@mkdir -p $(OUTPUT_DIR)/release
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d'/' -f1); \
		GOARCH=$$(echo $$platform | cut -d'/' -f2); \
		output_name=$(BINARY_NAME)-$$GOOS-$$GOARCH; \
		if [ $$GOOS = "windows" ]; then output_name=$$output_name.exe; fi; \
		cd $(OUTPUT_DIR) && tar -czf release/$$output_name-$(VERSION).tar.gz $$output_name; \
		cd -; \
	done
	@echo "Release packages created in $(OUTPUT_DIR)/release/"