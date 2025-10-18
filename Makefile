# TheBoys Launcher - Makefile
# This Makefile provides convenient targets for building the launcher

# Variables
APP_NAME := theboys-launcher
EXE_NAME := $(APP_NAME)$(if $(filter windows,$(shell go env GOOS)),.exe,)
VERSION ?= dev
BUILD_DIR := build
DIST_DIR := dist
LDFLAGS := -s -w -X main.version=$(VERSION)

# Default target
.PHONY: all
all: clean deps frontend build-all

# Help target
.PHONY: help
help:
	@echo "TheBoys Launcher - Build Targets"
	@echo "==============================="
	@echo ""
	@echo "Build Targets:"
	@echo "  all              - Clean, install deps, and build for all platforms"
	@echo "  build-all        - Build for all platforms (Windows, Linux, macOS)"
	@echo "  build-windows    - Build for Windows (amd64 + arm64)"
	@echo "  build-linux      - Build for Linux (amd64 + arm64)"
	@echo "  build-macos      - Build for macOS (amd64 + arm64)"
	@echo "  build-current    - Build for current platform only"
	@echo ""
	@echo "Development Targets:"
	@echo "  deps             - Install Go and Node.js dependencies"
	@echo "  frontend         - Build frontend assets"
	@echo "  dev              - Start development server"
	@echo "  test             - Run tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  install-wails    - Install Wails CLI"
	@echo ""
	@echo "Package Targets:"
	@echo "  package          - Create distribution packages"
	@echo "  package-windows  - Create Windows package"
	@echo "  package-linux    - Create Linux packages"
	@echo "  package-macos    - Create macOS packages"
	@echo ""
	@echo "Examples:"
	@echo "  make all VERSION=v1.0.0"
	@echo "  make build-current"
	@echo "  make dev"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f $(APP_NAME)-*
	rm -f TheBoysLauncher.exe
	rm -f theboys-launcher-*
	rm -f $(EXE_NAME)

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	cd frontend && npm install && cd ..

# Install Wails CLI
.PHONY: install-wails
install-wails:
	@echo "Installing Wails CLI..."
	go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Build frontend assets
.PHONY: frontend
frontend:
	@echo "Building frontend assets..."
	cd frontend && npm run build && cd ..

# Build for current platform only
.PHONY: build-current
build-current: clean deps frontend
	@echo "Building for current platform..."
	@if command -v wails >/dev/null 2>&1; then \
		wails build -tags dev -ldflags="$(LDFLAGS)" -o $(EXE_NAME); \
	else \
		echo "❌ Wails CLI not found. Please run 'make install-wails' first."; \
		exit 1; \
	fi
	@echo "Build complete: $(EXE_NAME)"

# Build for all platforms
.PHONY: build-all
build-all: clean deps frontend
	@echo "Building for all platforms..."
	@./build-all.sh $(VERSION)

# Build for Windows
.PHONY: build-windows
build-windows: clean deps frontend
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)/windows-amd64
	@mkdir -p $(BUILD_DIR)/windows-arm64
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/$(APP_NAME)-windows-amd64.exe ./cmd/launcher
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/windows-arm64/$(APP_NAME)-windows-arm64.exe ./cmd/launcher
	@echo "Windows builds complete"

# Build for Linux
.PHONY: build-linux
build-linux: clean deps frontend
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)/linux-amd64
	@mkdir -p $(BUILD_DIR)/linux-arm64
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/$(APP_NAME)-linux-amd64 ./cmd/launcher
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/$(APP_NAME)-linux-arm64 ./cmd/launcher
	chmod +x $(BUILD_DIR)/linux-amd64/$(APP_NAME)-linux-amd64
	chmod +x $(BUILD_DIR)/linux-arm64/$(APP_NAME)-linux-arm64
	@echo "Linux builds complete"

# Build for macOS
.PHONY: build-macos
build-macos: clean deps frontend
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)/darwin-amd64
	@mkdir -p $(BUILD_DIR)/darwin-arm64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/$(APP_NAME)-macos-amd64 ./cmd/launcher
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/darwin-arm64/$(APP_NAME)-macos-arm64 ./cmd/launcher
	chmod +x $(BUILD_DIR)/darwin-amd64/$(APP_NAME)-macos-amd64
	chmod +x $(BUILD_DIR)/darwin-arm64/$(APP_NAME)-macos-arm64
	@echo "macOS builds complete"

# Development server
.PHONY: dev
dev: deps
	@echo "Starting development server..."
	@echo "Note: Press Ctrl+C to stop the development server"
	@wails dev

# Run development build (uses wails dev for proper development)
.PHONY: run
run: deps
	@echo "Running TheBoys Launcher in development mode..."
	@echo "Note: Press Ctrl+C to stop the development server"
	@wails dev

# Run with specific mode
.PHONY: run-gui
run-gui: quick
	@echo "Running TheBoys Launcher in GUI mode..."
	@$(EXE_NAME)

.PHONY: run-cli
run-cli: quick
	@echo "Running TheBoys Launcher in CLI mode..."
	@$(EXE_NAME) --cli

# Run with development options
.PHONY: run-dev
run-dev: quick
	@echo "Running TheBoys Launcher in development mode..."
	@THEBOYS_DEV=1 $(EXE_NAME)

# Run tests in development mode
.PHONY: test-run
test-run: quick
	@echo "Running TheBoys Launcher with test data..."
	@THEBOYS_TEST=1 $(EXE_NAME) --cli --list-modpacks

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Create distribution packages
.PHONY: package
package: build-all
	@echo "Creating distribution packages..."
	@mkdir -p $(DIST_DIR)
	@echo "Packages created in $(DIST_DIR)"

# Build installers for all platforms
.PHONY: installers
installers: build-all
	@echo "Creating installers for all platforms..."
	@mkdir -p installers/dist

# Windows installer (GUI with directory selection)
.PHONY: installer-windows
installer-windows: build-windows
	@echo "Creating Windows GUI installer..."
	@mkdir -p installers/dist
	@if command -v makensis >/dev/null 2>&1; then \
		if [ -f "build/windows-amd64/theboys-launcher-windows-amd64.exe" ]; then \
			cp "build/windows-amd64/theboys-launcher-windows-amd64.exe" "installers/TheBoysLauncher.exe"; \
			cd installers && makensis windows-installer.nsi && cd ..; \
			echo "✓ Windows GUI installer created: installers/dist/TheBoysLauncher-Setup-$(VERSION).exe"; \
			echo "  - Professional installation wizard"; \
			echo "  - Directory selection"; \
			echo "  - Component selection"; \
			echo "  - Desktop shortcuts"; \
			echo "  - File associations"; \
		else \
			echo "❌ Windows build not found"; \
		fi; \
	else \
		echo "❌ NSIS not found. Install NSIS to create Windows installer."; \
		echo "  Download from: https://nsis.sourceforge.io/"; \
	fi

# macOS installer (GUI with proper wizard)
.PHONY: installer-macos
installer-macos: build-macos
	@echo "Creating macOS GUI installer..."
	@mkdir -p installers/dist
	@if command -v pkgbuild >/dev/null 2>&1; then \
		if [ -f "build/darwin-amd64/theboys-launcher-macos-amd64" ]; then \
			cp "build/darwin-amd64/theboys-launcher-macos-amd64" "installers/theboys-launcher-macos"; \
			chmod +x "installers/theboys-launcher-macos"; \
			cd installers && ./macos-build-installer.sh && cd ..; \
			echo "✓ macOS GUI installer created: installers/dist/TheBoys Launcher-$(VERSION).pkg"; \
			echo "  - Professional installation wizard"; \
			echo "  - Drag-and-drop interface"; \
			echo "  - Welcome and completion screens"; \
			echo "  - License agreement"; \
		else \
			echo "❌ macOS build not found"; \
		fi; \
	else \
		echo "❌ macOS build tools not found. Run on macOS to create macOS installer."; \
	fi

# Linux installers (Multiple GUI options)
.PHONY: installer-linux
installer-linux: build-linux
	@echo "Creating Linux GUI installers..."
	@mkdir -p installers/dist
	@if [ -f "build/linux-amd64/theboys-launcher-linux-amd64" ]; then \
		cp "build/linux-amd64/theboys-launcher-linux-amd64" "installers/theboys-launcher-linux"; \
		chmod +x "installers/theboys-launcher-linux"; \
		cd installers && ./linux-gui-installer.sh && cd ..; \
		echo "✓ Linux GUI installers created in installers/dist/"; \
		echo "  - Qt-based GUI installer (primary)"; \
		echo "  - Zenity fallback installer"; \
		echo "  - Dialog terminal installer"; \
		echo "  - AppImage self-contained installer"; \
	else \
		echo "❌ Linux build not found"; \
	fi

# Create all installers
.PHONY: installer-all
installer-all: installer-windows installer-macos installer-linux
	@echo "All installers created in installers/dist/"

# Create Windows package
.PHONY: package-windows
package-windows: build-windows
	@echo "Creating Windows package..."
	@mkdir -p $(DIST_DIR)
	@if [ -f "$(BUILD_DIR)/windows-amd64/$(APP_NAME)-windows-amd64.exe" ]; then \
		cp "$(BUILD_DIR)/windows-amd64/$(APP_NAME)-windows-amd64.exe" "$(DIST_DIR)/TheBoysLauncher.exe"; \
		echo "Created: $(DIST_DIR)/TheBoysLauncher.exe"; \
	fi
	@if [ -f "$(BUILD_DIR)/windows-arm64/$(APP_NAME)-windows-arm64.exe" ]; then \
		cp "$(BUILD_DIR)/windows-arm64/$(APP_NAME)-windows-arm64.exe" "$(DIST_DIR)/TheBoysLauncher-arm64.exe"; \
		echo "Created: $(DIST_DIR)/TheBoysLauncher-arm64.exe"; \
	fi

# Create Linux packages
.PHONY: package-linux
package-linux: build-linux
	@echo "Creating Linux packages..."
	@mkdir -p $(DIST_DIR)
	@if [ -f "$(BUILD_DIR)/linux-amd64/$(APP_NAME)-linux-amd64" ]; then \
		cp "$(BUILD_DIR)/linux-amd64/$(APP_NAME)-linux-amd64" "$(DIST_DIR)/theboys-launcher-linux-amd64"; \
		chmod +x "$(DIST_DIR)/theboys-launcher-linux-amd64"; \
		echo "Created: $(DIST_DIR)/theboys-launcher-linux-amd64"; \
	fi
	@if [ -f "$(BUILD_DIR)/linux-arm64/$(APP_NAME)-linux-arm64" ]; then \
		cp "$(BUILD_DIR)/linux-arm64/$(APP_NAME)-linux-arm64" "$(DIST_DIR)/theboys-launcher-linux-arm64"; \
		chmod +x "$(DIST_DIR)/theboys-launcher-linux-arm64"; \
		echo "Created: $(DIST_DIR)/theboys-launcher-linux-arm64"; \
	fi

# Create macOS packages
.PHONY: package-macos
package-macos: build-macos
	@echo "Creating macOS packages..."
	@mkdir -p $(DIST_DIR)
	@if [ -f "$(BUILD_DIR)/darwin-amd64/$(APP_NAME)-macos-amd64" ]; then \
		cp "$(BUILD_DIR)/darwin-amd64/$(APP_NAME)-macos-amd64" "$(DIST_DIR)/theboys-launcher-macos-amd64"; \
		chmod +x "$(DIST_DIR)/theboys-launcher-macos-amd64"; \
		echo "Created: $(DIST_DIR)/theboys-launcher-macos-amd64"; \
	fi
	@if [ -f "$(BUILD_DIR)/darwin-arm64/$(APP_NAME)-macos-arm64" ]; then \
		cp "$(BUILD_DIR)/darwin-arm64/$(APP_NAME)-macos-arm64" "$(DIST_DIR)/theboys-launcher-macos-arm64"; \
		chmod +x "$(DIST_DIR)/theboys-launcher-macos-arm64"; \
		echo "Created: $(DIST_DIR)/theboys-launcher-macos-arm64"; \
	fi

# Quick build for current platform
.PHONY: quick
quick: frontend
	@echo "Quick build for current platform..."
	@if command -v wails >/dev/null 2>&1; then \
		wails build -tags dev -skipfrontend -ldflags="$(LDFLAGS)" -o $(EXE_NAME); \
	else \
		echo "❌ Wails CLI not found. Please run 'make install-wails' first."; \
		exit 1; \
	fi
	@echo "Quick build complete: $(EXE_NAME)"

# Display version info
.PHONY: version
version:
	@echo "TheBoys Launcher"
	@echo "Version: $(VERSION)"
	@echo "Go: $(shell go version)"
	@echo "Node: $(shell node --version 2>/dev/null || echo 'Not found')"
	@echo "NPM: $(shell npm --version 2>/dev/null || echo 'Not found')"

# Install tools
.PHONY: tools
tools:
	@echo "Installing development tools..."
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest