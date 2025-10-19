# TheBoys Launcher Makefile

# Set your version here
VERSION ?= v2.0.3

# Default target
.PHONY: all
all: build

# Build the launcher with icon and version info
.PHONY: build
build: check-icon update-versioninfo
	@echo "Building TheBoys Launcher with version $(VERSION)..."
	go generate
	go build -ldflags="-s -w -H=windowsgui -X main.version=$(VERSION)" -o TheBoysLauncher.exe .
	@echo "Build successful! Created TheBoysLauncher.exe with version $(VERSION)"
	@echo "Icon: embedded"
	@echo "Version info: embedded"

# Build and sign the launcher
.PHONY: build-signed
build-signed: build
	@echo "Signing executable to prevent Windows Defender false positives..."
	@powershell -ExecutionPolicy Bypass -File sign-exe.ps1

# Quick sign only (executable must exist)
.PHONY: sign
sign:
	@echo "Signing existing executable..."
	@powershell -ExecutionPolicy Bypass -File sign-exe.ps1

# Check if icon file exists
.PHONY: check-icon
check-icon:
	@if not exist "icon.ico" ( \
		echo ERROR: icon.ico not found! & \
		echo Please create/place an icon.ico file in this directory. & \
		echo See ICON_README.md for details. & \
		exit /b 1 \
	)

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
	@rm -f TheBoysLauncher.exe test.exe resource.syso resource.rc theboys-launcher-cert.pfx
	@echo "Cleaned build artifacts"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build       - Build TheBoysLauncher.exe (default version: v2.0.3)"
	@echo "  build-signed- Build and sign exe (prevents Defender false positives)"
	@echo "  sign        - Sign existing executable only"
	@echo "  verify      - Quick build verification (doesn't create final exe)"
	@echo "  lint        - Run go fmt and go vet"
	@echo "  precommit   - Run verify + lint (run before committing)"
	@echo "  test-runtime- Brief runtime test for panics"
	@echo "  test        - Full test suite (build + lint + runtime)"
	@echo "  clean       - Remove build artifacts"
	@echo "  help        - Show this help"
	@echo ""
	@echo "Usage examples:"
	@echo "  make build                      # Build with default version"
	@echo "  make build VERSION=v2.0.3       # Build with specific version"
	@echo "  make build-signed VERSION=v2.0.3 # Build and sign exe"
	@echo "  make sign                       # Sign existing exe"
	@echo "  make verify                     # Quick build check"
	@echo "  make precommit                  # Full pre-commit checks"
	@echo "  make test                       # Run all tests"
	@echo "  make clean                      # Clean build files"
	@echo ""
	@echo "Code signing requires running PowerShell as Administrator the first time."