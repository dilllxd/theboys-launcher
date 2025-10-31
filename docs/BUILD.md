# Build Guide

This guide provides comprehensive instructions for building TheBoysLauncher on all supported platforms.

## 🛠️ Prerequisites

### General Requirements
- **Go 1.23+** - [Download Go](https://golang.org/dl/)
- **Git** - For cloning the repository
- **Make** - For using the Makefile (recommended)

### Platform-Specific Requirements

#### Windows
- **PowerShell 5.1+** (included with Windows 10+)
- **MinGW-w64** (optional, for certain CGO dependencies)
- **Git for Windows** - [Download](https://git-scm.com/download/win)
- **Inno Setup** - [Download](https://jrsoftware.org/isdl.php) (for creating Windows installer)

#### macOS
- **Xcode Command Line Tools**
  ```bash
  xcode-select --install
  ```
- **Homebrew** (recommended for tools)
  ```bash
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  ```

#### Linux
- **GCC** or compatible C compiler
  ```bash
  # Ubuntu/Debian
  sudo apt-get install build-essential

  # Fedora/RHEL
  sudo dnf groupinstall "Development Tools"

  # Arch Linux
  sudo pacman -S base-devel
  ```

## 🚀 Quick Start

### Clone and Build
```bash
# Clone the repository
git clone https://github.com/dilllxd/theboyslauncher.git
cd theboyslauncher

# Build for current platform
make build

# Or use go directly
go build -o TheBoysLauncher .
```

### Run the Application
```bash
# Windows
.\TheBoysLauncher.exe

# macOS/Linux
./TheBoysLauncher
```

## 🔨 Build System

### Makefile Targets

#### Build Commands
```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build specific platforms
make build-windows          # Windows executable
make build-macos            # macOS Intel
make build-macos-arm64      # macOS Apple Silicon
make build-macos-universal  # macOS Universal binary

# Quick verification (doesn't create final executable)
make verify
```

#### Package Commands
```bash
# Create app bundles
make package-macos-intel      # macOS Intel app bundle
make package-macos-arm64      # macOS Apple Silicon app bundle
make package-macos-universal  # macOS Universal app bundle
make package-macos            # All macOS variants

# Create DMG installers
make dmg-macos-intel      # macOS Intel DMG
make dmg-macos-arm64      # macOS Apple Silicon DMG
make dmg-macos-universal  # macOS Universal DMG
make dmg-macos            # All macOS DMGs

# Package for all platforms
make package-all
```

#### Development Commands
```bash
# Code formatting and linting
make lint

# Run tests
make test

# Quick build verification
make verify

# Full pre-commit checks
make precommit

# Clean build artifacts
make clean

# Show help
make help
```

## 🏗️ Cross-Platform Builds

### Using Makefile (Recommended)

#### Build All Platforms
```bash
make build-all
```

This creates:
- `build/windows/TheBoysLauncher.exe`
- `build/amd64/TheBoysLauncher` (macOS Intel)
- `build/arm64/TheBoysLauncher` (macOS ARM64)
- `build/linux/TheBoysLauncher` (Linux)

#### Create Universal macOS Binary
```bash
make build-macos-universal
```
This creates `build/universal/TheBoysLauncher` that works on both Intel and Apple Silicon Macs.

### Manual Cross-Compilation

#### Windows (from any platform)
```bash
export GOOS=windows GOARCH=amd64 CGO_ENABLED=0
go build -ldflags="-s -w -H=windowsgui -X main.version=v3.0.1" -o TheBoysLauncher.exe .
```

#### macOS Intel
```bash
export GOOS=darwin GOARCH=amd64 CGO_ENABLED=1
go build -ldflags="-s -w -X main.version=v3.0.1" -o TheBoysLauncher .
```

#### macOS Apple Silicon
```bash
export GOOS=darwin GOARCH=arm64 CGO_ENABLED=1
go build -ldflags="-s -w -X main.version=v3.0.1" -o TheBoysLauncher-arm64 .
```

#### Linux
```bash
export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
go build -ldflags="-s -w -X main.version=v3.0.1" -o TheBoysLauncher-linux .
```

## 📦 Packaging

### macOS App Bundles

#### Create App Bundle
```bash
# Intel app bundle
./scripts/create-app-bundle.sh amd64 v3.0.1

# Apple Silicon app bundle
./scripts/create-app-bundle.sh arm64 v3.0.1

# Universal app bundle
./scripts/create-app-bundle.sh universal v3.0.1
```

#### Create DMG Installer
```bash
# Install create-dmg (macOS only)
brew install create-dmg

# Create DMG
./scripts/create-dmg.sh universal v3.0.1
```

### Windows Installer

#### Create Inno Setup Installer
```bash
# Build the Windows installer using PowerShell
./tools/build-windows-msi.ps1

# Or use the build-installer script with a specific version
./scripts/build-installer.ps1 v3.2.68
```

The installer will be created in the `installer/` directory with the name `TheBoysLauncher-Setup-{version}.exe`.

#### Inno Setup Features
- Custom installation directory selection
- Optional desktop and start menu shortcuts
- Launch application after installation (checkbox)
- File association for .theboys files
- Registry entries for launcher to read installation path
- Proper uninstall with shortcut removal

### Icon Conversion

#### Convert Windows ICO to macOS ICNS
```bash
# Requires macOS with sips or iconutil
./scripts/convert-icon.sh
```

## 🧪 Testing

### Cross-Platform Test Suite
```bash
# Run comprehensive tests
./scripts/test-cross-platform.sh

# Test includes:
# - Basic compilation
# - Platform detection
# - Configuration loading
# - Java API integration
# - Archive extraction
# - File operations
# - Memory detection
# - Executable naming
# - PATH handling
# - Update system
# - Script availability
# - Dependencies
```

### Unit Tests
```bash
go test -v ./...
```

### Build Verification
```bash
make verify
```

## 🔧 Build Configuration

### Version Information
Set version during build:
```bash
go build -ldflags="-X main.version=v3.0.1" -o TheBoysLauncher .
```

### Build Flags
Common useful flags:
```bash
# Strip debug information (smaller binary)
-ldflags="-s -w"

# Windows GUI mode (no console)
-ldflags="-H=windowsgui"

# Set version
-ldflags="-X main.version=v3.0.1"

# Combined
-ldflags="-s -w -H=windowsgui -X main.version=v3.0.1"
```

### Environment Variables
```bash
# Custom build directory
export BUILD_DIR="/path/to/build"

# Custom version
export VERSION="v3.0.1-custom"

# Skip icon check (for builds without icon)
export SKIP_ICON_CHECK="1"
```

## 🐛 Troubleshooting

### Build Issues

#### "go: build constraints exclude all Go files"
- Check platform-specific file build tags
- Ensure `//go:build darwin` and `//go:build windows` are correct
- Verify file names match build constraints

#### "CGO_ENABLED=1 required but not available"
- Install C compiler for your platform
- Use `CGO_ENABLED=0` for pure Go builds
- macOS: Install Xcode Command Line Tools

#### "command not found: make"
- Install Make:
  - macOS: `brew install make`
  - Ubuntu: `sudo apt-get install build-essential`
  - Windows: Use WSL or use `go build` directly

#### "icon.ico not found"
- Place `icon.ico` in project root
- Or use `SKIP_ICON_CHECK=1` to skip icon requirement

### Runtime Issues

#### "Fyne requires CGO"
- Use `CGO_ENABLED=1` for builds
- Install C compiler toolchain

#### "Missing dylib on macOS"
- Use `CGO_ENABLED=1` for proper linking
- Check Xcode Command Line Tools installation

## 📁 Build Artifacts

### Directory Structure
```
build/
├── windows/
│   └── TheBoysLauncher.exe          # Windows executable
├── amd64/
│   ├── TheBoysLauncher              # macOS Intel executable
│   └── TheBoysLauncher.app/         # macOS Intel app bundle
├── arm64/
│   ├── TheBoysLauncher              # macOS ARM64 executable
│   └── TheBoysLauncher.app/         # macOS ARM64 app bundle
├── universal/
│   ├── TheBoysLauncher              # macOS Universal executable
│   ├── TheBoysLauncher.app/         # macOS Universal app bundle
│   └── TheBoysLauncher-Universal.dmg # macOS Universal DMG
└── linux/
    └── TheBoysLauncher-linux        # Linux executable
```

### File Sizes (Approximate)
- Windows executable: ~25 MB
- macOS executable: ~30 MB
- macOS app bundle: ~35 MB
- macOS DMG: ~45 MB
- Linux executable: ~25 MB

## 🔄 CI/CD Integration

### GitHub Actions
The project includes a comprehensive GitHub Actions workflow:

```yaml
# Triggers:
# - Push to macos-support branch
# - Pull requests to macos-support branch
# - Manual workflow dispatch

# Builds:
# - Windows (amd64)
# - macOS Intel (amd64)
# - macOS Apple Silicon (arm64)
# - macOS Universal (combined)

# Tests:
# - Unit tests
# - Linting
# - Cross-platform validation
```

### Local CI Testing
```bash
# Test the same way as CI
make test        # Run all tests
make lint        # Code formatting
make verify      # Build verification
```

## 🚀 Release Process

### Create Release Build
```bash
# Set version
export VERSION="v3.0.1"

# Build all platforms
make package-all

# Create release directory
mkdir -p release

# Copy artifacts
cp build/windows/TheBoysLauncher.exe release/TheBoysLauncher-${VERSION}.exe
cp build/linux/TheBoysLauncher-linux release/TheBoysLauncher-${VERSION}-linux
cp -r build/universal/TheBoysLauncher.app release/TheBoysLauncher.app
cp build/universal/TheBoysLauncher-Universal.dmg release/TheBoysLauncher-${VERSION}-Universal.dmg
```

### Verify Release
```bash
# Test each platform's artifact
# - Windows: Run .exe on Windows
# - macOS: Open .app and .dmg on macOS
# - Run cross-platform test suite
./scripts/test-cross-platform.sh
```

---

For platform-specific installation instructions, see [INSTALL_MACOS.md](./INSTALL_MACOS.md). For general information, see the main [README.md](./README.md).