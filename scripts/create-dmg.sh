#!/bin/bash

# Create macOS DMG installer for TheBoys Launcher
# Usage: ./create-dmg.sh <arch> <version>
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

BUILD_DIR="build/$ARCH"
APP_NAME="TheBoysLauncher.app"
DMG_NAME="TheBoys-Launcher-$VERSION-$ARCH"
DMG_PATH="$DMG_NAME.dmg"

echo "Creating macOS DMG for $ARCH..."
echo "Version: $VERSION"
echo "DMG name: $DMG_PATH"

# Check if app bundle exists
if [ ! -d "$BUILD_DIR/$APP_NAME" ]; then
    echo "Error: App bundle not found: $BUILD_DIR/$APP_NAME"
    echo "Please create the app bundle first with: make package-macos-$ARCH"
    exit 1
fi

# Create temporary DMG directory
DMG_DIR="dmg_temp_$ARCH"
rm -rf "$DMG_DIR"
mkdir "$DMG_DIR"

# Copy app bundle to DMG directory
echo "Copying app bundle..."
cp -R "$BUILD_DIR/$APP_NAME" "$DMG_DIR/"

# Create Applications symbolic link
echo "Creating Applications symlink..."
ln -s /Applications "$DMG_DIR/Applications"

# Create DMG background and setup (if create-dmg is available)
if command -v create-dmg &> /dev/null; then
    echo "Creating styled DMG with create-dmg..."
    create-dmg \
        --volname "TheBoys Launcher ($ARCH)" \
        --volicon "$BUILD_DIR/$APP_NAME/Contents/Resources/AppIcon.icns" \
        --window-pos 200 120 \
        --window-size 600 400 \
        --icon-size 100 \
        --icon "$APP_NAME" 175 120 \
        --hide-extension "$APP_NAME" \
        --app-drop-link 425 120 \
        --disk-image-size 100 \
        "$DMG_PATH" \
        "$DMG_DIR" \
        || echo "create-dmg failed, falling back to hdiutil..."
else
    echo "Creating basic DMG with hdiutil..."

    # Create DMG
    hdiutil create -srcfolder "$DMG_DIR" -volname "TheBoys Launcher ($ARCH)" -fs HFS+ -fsargs "-c c=64,a=16,e=16" -format UDRW -size 100m "$DMG_PATH.temp.dmg"

    # Mount DMG
    DEVICE=$(hdiutil attach -readwrite -noverify -noautoopen "$DMG_PATH.temp.dmg" | egrep '^/dev/' | sed 1q | awk '{print $1}')

    # Get mount point
    MOUNT_POINT=$(hdiutil attach "$DMG_PATH.temp.dmg" | grepVolumes | sed 's/.*Volumes\///')

    # Set up appearance (basic)
    echo '
    tell application "Finder"
        tell disk "'TheBoys Launcher ($ARCH)'"
            open
            set current view of container window to icon view
            set toolbar visible of container window to false
            set statusbar visible of container window to false
            set the bounds of container window to {400, 100, 1000, 500}
            set theViewOptions to the icon view options of container window
            set arrangement of theViewOptions to not arranged
            set icon size of theViewOptions to 100
            set position of item "'$APP_NAME'" of container window to {175, 120}
            set position of item "Applications" of container window to {425, 120}
            update without registering applications
            close
            open
        end tell
    end tell
    ' | osascript || echo "Could not set DMG appearance (this is normal on non-macOS systems)"

    # Unmount and convert to compressed format
    hdiutil detach "$DEVICE"
    hdiutil convert "$DMG_PATH.temp.dmg" -format UDZO -imagekey zlib-level=9 -o "$DMG_PATH"
    rm -f "$DMG_PATH.temp.dmg"
fi

# Clean up temporary directory
rm -rf "$DMG_DIR"

# Check if DMG was created successfully
if [ -f "$DMG_PATH" ]; then
    echo "DMG created successfully!"
    echo "File: $DMG_PATH"
    echo "Size: $(du -h "$DMG_PATH" | cut -f1)"
else
    echo "Error: DMG creation failed!"
    exit 1
fi