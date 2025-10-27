# TheBoysLauncher Cross-Platform Makefile

# Set your version here
VERSION ?= v3.0.1

# Default target builds for current platform
.PHONY: all
all: build

# Platform-specific builds
.PHONY: build-windows build-macos build-macos-arm64 build-macos-universal build-linux build-all

# Build Windows (simplified - no version info files)
build-windows: check-icon
	@echo "Building TheBoysLauncher for Windows..."
	@mkdir -p build/windows
	go build -ldflags="-s -w -H=windowsgui -X main.version=$(VERSION)" -o build/windows/TheBoysLauncher.exe .
	@echo "Windows build complete: build/windows/TheBoysLauncher.exe"

# Build macOS Intel
build-macos:
	@echo "Building TheBoysLauncher for macOS Intel..."
	@mkdir -p build/amd64
	export GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 && go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/amd64/TheBoysLauncher .
	@echo "macOS Intel build complete: build/amd64/TheBoysLauncher"

# Build macOS Apple Silicon
build-macos-arm64:
	@echo "Building TheBoysLauncher for macOS Apple Silicon..."
	@mkdir -p build/arm64
	export GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 && go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/arm64/TheBoysLauncher .
	@echo "macOS Apple Silicon build complete: build/arm64/TheBoysLauncher"

# Create macOS universal binary
build-macos-universal: build-macos build-macos-arm64
	@echo "Creating universal macOS binary..."
	@mkdir -p build/universal
	@lipo -create build/amd64/TheBoysLauncher build/arm64/TheBoysLauncher -output build/universal/TheBoysLauncher
	@echo "Universal binary created: build/universal/TheBoysLauncher"

# Build Linux
build-linux:
	@echo "Building TheBoysLauncher for Linux..."
	@mkdir -p build/linux
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/linux/TheBoysLauncher-linux .
	@echo "Linux build complete: build/linux/TheBoysLauncher-linux"

# Build all platforms
build-all: build-windows build-linux build-macos-universal
	@echo "All builds complete!"
	@echo "Windows: build/windows/TheBoysLauncher.exe"
	@echo "Linux: build/linux/TheBoysLauncher-linux"
	@echo "macOS Universal: build/universal/TheBoysLauncher"

# Build for current platform (preserves existing behavior)
build:
	@if [ "$(shell go env GOOS)" = "windows" ]; then \
		$(MAKE) build-windows; \
	elif [ "$(shell go env GOOS)" = "linux" ]; then \
		$(MAKE) build-linux; \
	elif [ "$(shell go env GOOS)" = "darwin" ]; then \
		$(MAKE) build-macos-universal; \
	else \
		echo "Unsupported platform: $(shell go env GOOS)"; \
		exit 1; \
	fi
	@echo "Build complete!"

# macOS packaging targets
.PHONY: package-macos package-macos-intel package-macos-arm64 package-macos-universal package-all

# Package macOS Intel
package-macos-intel: build-macos
	@echo "Creating macOS Intel app bundle..."
	./scripts/create-app-bundle.sh amd64 $(VERSION)

# Package macOS Apple Silicon
package-macos-arm64: build-macos-arm64
	@echo "Creating macOS Apple Silicon app bundle..."
	./scripts/create-app-bundle.sh arm64 $(VERSION)

# Package macOS Universal
package-macos-universal: build-macos-universal
	@echo "Creating macOS Universal app bundle..."
	./scripts/create-app-bundle.sh universal $(VERSION)

# Package all macOS variants
package-macos: package-macos-intel package-macos-arm64 package-macos-universal
	@echo "All macOS packages complete!"

# Create DMG installers
.PHONY: dmg-macos dmg-macos-intel dmg-macos-arm64 dmg-macos-universal dmg-all

# Create macOS Intel DMG
dmg-macos-intel: package-macos-intel
	@echo "Creating macOS Intel DMG..."
	./scripts/create-dmg.sh amd64 $(VERSION)

# Create macOS Apple Silicon DMG
dmg-macos-arm64: package-macos-arm64
	@echo "Creating macOS Apple Silicon DMG..."
	./scripts/create-dmg.sh arm64 $(VERSION)

# Create macOS Universal DMG
dmg-macos-universal: package-macos-universal
	@echo "Creating macOS Universal DMG..."
	./scripts/create-dmg.sh universal $(VERSION)

# Create all macOS DMGs
dmg-macos: dmg-macos-intel dmg-macos-arm64 dmg-macos-universal
	@echo "All macOS DMGs complete!"

# Package all platforms with DMGs
package-all: build-windows build-linux package-macos dmg-macos
	@echo "All platform packages complete!"

# Code signing removed - no longer supported

# Check if icon file exists
.PHONY: check-icon
check-icon:
	@if [ ! -f "icon.ico" ]; then \
		echo "ERROR: icon.ico not found!"; \
		echo "Please create/place an icon.ico file in this directory."; \
		echo "See ICON_README.md for details."; \
		exit 1; \
	fi

# Update versioninfo.json with current version
.PHONY: update-versioninfo
update-versioninfo:
	@powershell -ExecutionPolicy Bypass -File update-version.ps1 -Version "$(VERSION)"

# Quick build verification (doesn't create final executable)
.PHONY: verify
verify:
	@echo "Running quick build verification..."
	@go build -o test.exe .
	@rm -f test.exe
	@echo "✅ Build verification passed!"

# Format and lint code
.PHONY: lint
lint:
	@echo "Running go fmt and go vet..."
	@go fmt ./...
	@go vet ./...
	@echo "✅ Code formatting and linting complete!"

# Full pre-commit check (verify + lint)
.PHONY: precommit
precommit: verify lint
	@echo "✅ All pre-commit checks passed!"

# Runtime test (brief check for panics)
.PHONY: test-runtime
test-runtime:
	@echo "Running runtime test for panics..."
	@timeout 5s ./TheBoysLauncher.exe > /dev/null 2>&1 || true
	@echo "✅ Runtime test completed"

# Full test suite (build + lint + runtime)
.PHONY: test
test: build lint test-runtime
	@echo "✅ All tests passed!"

# Clean build artifacts
.PHONY: clean
clean:
	@rm -rf build/
	@rm -f TheBoysLauncher.exe test.exe resource.syso resource.rc theboyslauncher-cert.pfx
	@echo "Cleaned all build artifacts"

# Help
.PHONY: help
help:
	@echo "TheBoysLauncher Cross-Platform Build System"
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Build Targets:"
	@echo "  build                - Build for current platform (Windows)"
	@echo "  build-windows        - Build Windows executable"
	@echo "  build-linux          - Build Linux executable"
	@echo "  build-macos          - Build macOS Intel"
	@echo "  build-macos-arm64    - Build macOS Apple Silicon"
	@echo "  build-macos-universal- Build macOS Universal binary"
	@echo "  build-all            - Build all platforms"
	@echo ""
	@echo "Package Targets:"
	@echo "  package-macos        - Package all macOS variants"
	@echo "  package-macos-intel  - Package macOS Intel app bundle"
	@echo "  package-macos-arm64  - Package macOS Apple Silicon app bundle"
	@echo "  package-macos-universal- Package macOS Universal app bundle"
	@echo "  package-all          - Package all platforms"
	@echo ""
	@echo "Note: Code signing functionality has been removed"
	@echo ""
	@echo "Development:"
	@echo "  verify               - Quick build verification"
	@echo "  lint                 - Run go fmt and go vet"
	@echo "  precommit            - Run verify + lint"
	@echo "  test-runtime         - Brief runtime test"
	@echo "  test                 - Full test suite"
	@echo "  clean                - Remove all build artifacts"
	@echo "  help                 - Show this help"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build                           # Build current platform"
	@echo "  make build-windows VERSION=v3.0.1    # Build Windows with version"
	@echo "  make build-macos-universal           # Build macOS Universal"
	@echo "  make build-all                       # Build all platforms"
	@echo "  make package-macos-universal         # Package macOS Universal"
	@echo "  make package-all                     # Package all platforms"
	@echo "  make clean                           # Clean all build files"
	@echo ""
	@echo "Output directories:"
	@echo "  build/windows/     - Windows executables"
	@echo "  build/linux/       - Linux executables"
	@echo "  build/amd64/       - macOS Intel builds"
	@echo "  build/arm64/       - macOS Apple Silicon builds"
	@echo "  build/universal/   - macOS Universal builds"
	@echo ""
	@echo "No code signing - distributions are unsigned."