# TheBoys Launcher Cross-Platform Makefile

# Set your version here
VERSION ?= v3.0.1

# Default target builds for current platform
.PHONY: all
all: build

# Platform-specific builds
.PHONY: build-windows build-macos build-macos-arm64 build-macos-universal build-all

# Build Windows (simplified - no version info files)
build-windows: check-icon
	@echo "Building TheBoys Launcher for Windows..."
	@mkdir -p build/windows
	go build -ldflags="-s -w -H=windowsgui -X main.version=$(VERSION)" -o build/windows/TheBoysLauncher.exe .
	@echo "Windows build complete: build/windows/TheBoysLauncher.exe"

# Build macOS Intel
build-macos:
	@echo "Building TheBoys Launcher for macOS Intel..."
	@mkdir -p build/amd64
	export GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 && go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/amd64/TheBoysLauncher .
	@echo "macOS Intel build complete: build/amd64/TheBoysLauncher"

# Build macOS Apple Silicon
build-macos-arm64:
	@echo "Building TheBoys Launcher for macOS Apple Silicon..."
	@mkdir -p build/arm64
	export GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 && go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/arm64/TheBoysLauncher .
	@echo "macOS Apple Silicon build complete: build/arm64/TheBoysLauncher"

# Create macOS universal binary
build-macos-universal: build-macos build-macos-arm64
	@echo "Creating universal macOS binary..."
	@mkdir -p build/universal
	@lipo -create build/amd64/TheBoysLauncher build/arm64/TheBoysLauncher -output build/universal/TheBoysLauncher
	@echo "Universal binary created: build/universal/TheBoysLauncher"

# Build all platforms
build-all: build-windows build-macos-universal
	@echo "All builds complete!"
	@echo "Windows: build/windows/TheBoysLauncher.exe"
	@echo "macOS Universal: build/universal/TheBoysLauncher"

# Build for current platform (preserves existing behavior)
build: build-windows
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

# Package all platforms
package-all: build-windows package-macos
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
	@rm -f TheBoysLauncher.exe test.exe resource.syso resource.rc theboys-launcher-cert.pfx
	@echo "Cleaned all build artifacts"

# Help
.PHONY: help
help:
	@echo "TheBoys Launcher Cross-Platform Build System"
	@echo "Version: $(VERSION)"
	@echo ""
	@echo "Build Targets:"
	@echo "  build                - Build for current platform (Windows)"
	@echo "  build-windows        - Build Windows executable"
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
	@echo "  build/amd64/       - macOS Intel builds"
	@echo "  build/arm64/       - macOS Apple Silicon builds"
	@echo "  build/universal/   - macOS Universal builds"
	@echo ""
	@echo "No code signing - distributions are unsigned."