#!/bin/bash

# TheBoys Launcher - Cross-Platform Build Script
# This script builds single-file executables for Windows, macOS, and Linux
# Each executable is self-contained and drops all files in the same directory

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build configuration
APP_NAME="theboys-launcher"
VERSION=${1:-"dev"}
BUILD_DIR="build"
DIST_DIR="dist"

echo -e "${BLUE}TheBoys Launcher - Cross-Platform Build Script${NC}"
echo "=================================================="
echo "Version: $VERSION"
echo ""

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf "$BUILD_DIR"
rm -rf "$DIST_DIR"
rm -f "$APP_NAME"-*

# Install dependencies
echo -e "${YELLOW}Installing dependencies...${NC}"

# Install Go dependencies
echo "Installing Go dependencies..."
go mod download
go mod tidy

# Install frontend dependencies
echo "Installing frontend dependencies..."
cd frontend
npm install --silent
cd ..

# Check if Wails CLI is available, install if needed
if ! command -v wails &> /dev/null; then
    echo -e "${YELLOW}Wails CLI not found, installing...${NC}"
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    export PATH="$PATH:$(go env GOPATH)/bin"
fi

# Build frontend assets
echo -e "${YELLOW}Building frontend assets...${NC}"
cd frontend
npm run build
cd ..

echo ""
echo -e "${GREEN}Starting cross-platform builds...${NC}"
echo ""

# Function to build for a specific platform
build_platform() {
    local os=$1
    local arch=$2
    local output_name=$3
    local ldflags=$4

    echo -e "${BLUE}Building for $os/$arch...${NC}"

    # Set environment variables for cross-compilation
    export GOOS=$os
    export GOARCH=$arch
    export CGO_ENABLED=0

    # Create build directory
    local platform_dir="$BUILD_DIR/$os-$arch"
    mkdir -p "$platform_dir"

    # Build the application
    if [ "$os" = "windows" ]; then
        # Windows build with .exe extension
        local exe_name="$output_name.exe"
        go build -ldflags="$ldflags" -o "$platform_dir/$exe_name" ./cmd/launcher

        # Verify the executable
        if [ -f "$platform_dir/$exe_name" ]; then
            local size=$(du -h "$platform_dir/$exe_name" | cut -f1)
            echo -e "${GREEN}✓ Windows build successful: $exe_name ($size)${NC}"
        else
            echo -e "${RED}✗ Windows build failed${NC}"
            return 1
        fi
    else
        # Unix-like build (Linux, macOS)
        go build -ldflags="$ldflags" -o "$platform_dir/$output_name" ./cmd/launcher

        # Verify the executable
        if [ -f "$platform_dir/$output_name" ]; then
            # Make executable
            chmod +x "$platform_dir/$output_name"
            local size=$(du -h "$platform_dir/$output_name" | cut -f1)
            echo -e "${GREEN}✓ $os build successful: $output_name ($size)${NC}"
        else
            echo -e "${RED}✗ $os build failed${NC}"
            return 1
        fi
    fi

    # Unset environment variables
    unset GOOS
    unset GOARCH
    unset CGO_ENABLED

    return 0
}

# Build flags for version injection
LDFLAGS="-s -w -X main.version=$VERSION"

# Build for Windows (amd64)
build_platform "windows" "amd64" "$APP_NAME-windows-amd64" "$LDFLAGS"

# Build for Windows (arm64)
build_platform "windows" "arm64" "$APP_NAME-windows-arm64" "$LDFLAGS"

# Build for Linux (amd64)
build_platform "linux" "amd64" "$APP_NAME-linux-amd64" "$LDFLAGS"

# Build for Linux (arm64)
build_platform "linux" "arm64" "$APP_NAME-linux-arm64" "$LDFLAGS"

# Build for macOS (amd64)
build_platform "darwin" "amd64" "$APP_NAME-macos-amd64" "$LDFLAGS"

# Build for macOS (arm64 - Apple Silicon)
build_platform "darwin" "arm64" "$APP_NAME-macos-arm64" "$LDFLAGS"

echo ""
echo -e "${GREEN}All builds completed successfully!${NC}"

# Create distribution directory with proper naming
echo -e "${YELLOW}Creating distribution packages...${NC}"
mkdir -p "$DIST_DIR"

# Copy builds to dist directory with proper names
cd "$BUILD_DIR"

# Windows builds
if [ -f "windows-amd64/$APP_NAME-windows-amd64.exe" ]; then
    cp "windows-amd64/$APP_NAME-windows-amd64.exe" "../$DIST_DIR/TheBoysLauncher.exe"
    echo -e "${GREEN}✓ Created: TheBoysLauncher.exe (Windows x64)${NC}"
fi

if [ -f "windows-arm64/$APP_NAME-windows-arm64.exe" ]; then
    cp "windows-arm64/$APP_NAME-windows-arm64.exe" "../$DIST_DIR/TheBoysLauncher-arm64.exe"
    echo -e "${GREEN}✓ Created: TheBoysLauncher-arm64.exe (Windows ARM64)${NC}"
fi

# Linux builds
if [ -f "linux-amd64/$APP_NAME-linux-amd64" ]; then
    cp "linux-amd64/$APP_NAME-linux-amd64" "../$DIST_DIR/theboys-launcher-linux-amd64"
    chmod +x "../$DIST_DIR/theboys-launcher-linux-amd64"
    echo -e "${GREEN}✓ Created: theboys-launcher-linux-amd64 (Linux x64)${NC}"
fi

if [ -f "linux-arm64/$APP_NAME-linux-arm64" ]; then
    cp "linux-arm64/$APP_NAME-linux-arm64" "../$DIST_DIR/theboys-launcher-linux-arm64"
    chmod +x "../$DIST_DIR/theboys-launcher-linux-arm64"
    echo -e "${GREEN}✓ Created: theboys-launcher-linux-arm64 (Linux ARM64)${NC}"
fi

# macOS builds
if [ -f "darwin-amd64/$APP_NAME-macos-amd64" ]; then
    cp "darwin-amd64/$APP_NAME-macos-amd64" "../$DIST_DIR/theboys-launcher-macos-amd64"
    chmod +x "../$DIST_DIR/theboys-launcher-macos-amd64"
    echo -e "${GREEN}✓ Created: theboys-launcher-macos-amd64 (macOS Intel)${NC}"
fi

if [ -f "darwin-arm64/$APP_NAME-macos-arm64" ]; then
    cp "darwin-arm64/$APP_NAME-macos-arm64" "../$DIST_DIR/theboys-launcher-macos-arm64"
    chmod +x "../$DIST_DIR/theboys-launcher-macos-arm64"
    echo -e "${GREEN}✓ Created: theboys-launcher-macos-arm64 (macOS Apple Silicon)${NC}"
fi

cd ..

echo ""
echo -e "${BLUE}Build Summary${NC}"
echo "=============="
echo "All executables are self-contained and portable:"
echo ""
echo -e "${YELLOW}Windows:${NC}"
echo "  - TheBoysLauncher.exe (x64)"
echo "  - TheBoysLauncher-arm64.exe (ARM64)"
echo ""
echo -e "${YELLOW}Linux:${NC}"
echo "  - theboys-launcher-linux-amd64 (x64)"
echo "  - theboys-launcher-linux-arm64 (ARM64)"
echo ""
echo -e "${YELLOW}macOS:${NC}"
echo "  - theboys-launcher-macos-amd64 (Intel)"
echo "  - theboys-launcher-macos-arm64 (Apple Silicon)"
echo ""
echo -e "${GREEN}Single-File Deployment: ✅${NC}"
echo "Each executable drops all files in the same directory as the executable"
echo "just like the legacy launcher. No installation required!"
echo ""
echo -e "${GREEN}Portable Operation: ✅${NC}"
echo "- Windows: Creates files beside the .exe"
echo "- macOS: Creates ~/.theboys-launcher/ in user home"
echo "- Linux: Creates ~/.theboys-launcher/ in user home"
echo ""
echo -e "${GREEN}Build artifacts are available in: $DIST_DIR${NC}"