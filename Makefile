# TheBoys Launcher Makefile

# Set your version here
VERSION ?= v1.1.0

# Default target
.PHONY: all
all: build

# Build the launcher
.PHONY: build
build:
	@echo "Building TheBoys Launcher with version $(VERSION)..."
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o TheBoysLauncher.exe .
	@echo "Build successful! Created TheBoysLauncher.exe with version $(VERSION)"

# Clean build artifacts
.PHONY: clean
clean:
	@if exist TheBoysLauncher.exe del TheBoysLauncher.exe
	@if exist test.exe del test.exe
	@echo "Cleaned build artifacts"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build     - Build TheBoysLauncher.exe (default version: v1.1.0)"
	@echo "  clean     - Remove build artifacts"
	@echo "  help      - Show this help"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build                    # Build with default version"
	@echo "  make build VERSION=v1.2.0     # Build with specific version"
	@echo "  make clean                    # Clean build files"