#!/bin/bash

# Create macOS app bundle for TheBoys Launcher
# Usage: ./create-app-bundle.sh <arch> <version>
# arch: amd64, arm64, or universal
# version: version string (e.g., v3.0.1)

set -e

# Check arguments
if [ $# -ne 2 ]; then
    echo "Usage: $0 <arch> <version>"
    echo "arch: amd64, arm64, or universal"
    echo "version: version string (e.g., v3.0.1)"
    exit 1
fi

ARCH=$1
VERSION=$2

# Validate architecture
case $ARCH in
    amd64|arm64|universal)
        ;;
    *)
        echo "Error: Invalid architecture '$ARCH'. Must be amd64, arm64, or universal."
        exit 1
        ;;
esac

# Determine source directory
case $ARCH in
    amd64)
        SOURCE_DIR="build/amd64"
        ;;
    arm64)
        SOURCE_DIR="build/arm64"
        ;;
    universal)
        SOURCE_DIR="build/universal"
        ;;
esac

APP_NAME="TheBoys Launcher"
BUNDLE_NAME="TheBoysLauncher.app"
BUILD_DIR="build/$ARCH"
APP_DIR="$BUILD_DIR/$BUNDLE_NAME"

echo "Creating macOS app bundle for $ARCH..."
echo "Version: $VERSION"
echo "Source: $SOURCE_DIR/TheBoysLauncher"
echo "Target: $APP_DIR"

# Check if source binary exists
if [ ! -f "$SOURCE_DIR/TheBoysLauncher" ]; then
    echo "Error: Source binary not found: $SOURCE_DIR/TheBoysLauncher"
    echo "Please build the binary first with: make build-$ARCH"
    exit 1
fi

# Create app bundle structure
echo "Creating app bundle structure..."
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# Create Info.plist
echo "Creating Info.plist..."
cat > "$APP_DIR/Contents/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDisplayName</key>
    <string>$APP_NAME</string>
    <key>CFBundleExecutable</key>
    <string>TheBoysLauncher</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.theboys.launcher</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>$VERSION</string>
    <key>CFBundleVersion</key>
    <string>$VERSION</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>CFBundleDocumentTypes</key>
    <array>
        <dict>
            <key>CFBundleTypeExtensions</key>
            <array>
                <string>toml</string>
            </array>
            <key>CFBundleTypeName</key>
            <string>Modpack Configuration</string>
            <key>CFBundleTypeRole</key>
            <string>Editor</string>
        </dict>
    </array>
</dict>
</plist>
EOF

# Copy executable
echo "Copying executable..."
cp "$SOURCE_DIR/TheBoysLauncher" "$APP_DIR/Contents/MacOS/"
chmod +x "$APP_DIR/Contents/MacOS/TheBoysLauncher"

# Create icon placeholder (if icon doesn't exist)
ICON_PATH="$APP_DIR/Contents/Resources/AppIcon.icns"
if [ ! -f "resources/darwin/AppIcon.icns" ]; then
    echo "Creating placeholder icon..."
    # Create a simple placeholder icon using sips if available
    if command -v sips &> /dev/null; then
        # Create a simple 512x512 placeholder
        mkdir -p "resources/darwin/TheBoysLauncher.iconset"
        # This would require an actual image file, so we'll just create the directory structure
        touch "resources/darwin/TheBoysLauncher.iconset/icon_512x512.png"
    fi
else
    echo "Copying icon..."
    cp "resources/darwin/AppIcon.icns" "$ICON_PATH"
fi

# Set proper permissions
echo "Setting permissions..."
chmod -R 755 "$APP_DIR"
chmod +x "$APP_DIR/Contents/MacOS/TheBoysLauncher"

echo "App bundle created successfully!"
echo "Location: $APP_DIR"
echo "Bundle size: $(du -sh "$APP_DIR" | cut -f1)"

# Verify bundle structure
echo ""
echo "Bundle structure:"
ls -la "$APP_DIR/Contents/"
echo ""
echo "Executable:"
ls -la "$APP_DIR/Contents/MacOS/"