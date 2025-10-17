#!/bin/bash

# TheBoys Launcher macOS Installer Script
# Creates a proper macOS application bundle and installer package

set -e

# Configuration
APP_NAME="TheBoys Launcher"
BUNDLE_ID="com.theboys.launcher"
APP_VERSION="1.0.0"
APP_EXECUTABLE="theboys-launcher-macos"
ICON_FILE="AppIcon.icns"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}TheBoys Launcher - macOS Installer Builder${NC}"
echo "=============================================="
echo "Version: $APP_VERSION"
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo -e "${RED}Error: This script must be run on macOS${NC}"
    exit 1
fi

# Check required tools
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        echo "Please install $1 and try again"
        exit 1
    fi
}

echo -e "${YELLOW}Checking required tools...${NC}"
check_tool "pkgbuild"
check_tool "productbuild"
check_tool "xcrun"

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf build/
rm -rf dist/
rm -f *.pkg

# Create build directories
mkdir -p build/
mkdir -p dist/

# Check if executable exists
if [ ! -f "$APP_EXECUTABLE" ]; then
    echo -e "${RED}Error: Executable '$APP_EXECUTABLE' not found${NC}"
    echo "Please build the application first using: make build-macos"
    exit 1
fi

# Create application bundle structure
echo -e "${YELLOW}Creating application bundle...${NC}"
APP_BUNDLE="build/${APP_NAME}.app"
APP_CONTENTS="$APP_BUNDLE/Contents"
APP_MACOS="$APP_CONTENTS/MacOS"
APP_RESOURCES="$APP_CONTENTS/Resources"

mkdir -p "$APP_MACOS"
mkdir -p "$APP_RESOURCES"
mkdir -p "$APP_CONTENTS/Frameworks"

# Create Info.plist
echo -e "${YELLOW}Creating Info.plist...${NC}"
cat > "$APP_CONTENTS/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDisplayName</key>
    <string>${APP_NAME}</string>
    <key>CFBundleExecutable</key>
    <string>${APP_EXECUTABLE}</string>
    <key>CFBundleIconFile</key>
    <string>${ICON_FILE}</string>
    <key>CFBundleIdentifier</key>
    <string>${BUNDLE_ID}</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${APP_VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${APP_VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>LSApplicationCategoryType</key>
    <string>public.app-category.games</string>
    <key>NSPrincipalClass</key>
    <string>NSApplication</string>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>NSAppTransportSecurity</key>
    <dict>
        <key>NSAllowsArbitraryLoads</key>
        <true/>
    </dict>
</dict>
</plist>
EOF

# Copy executable
echo -e "${YELLOW}Copying executable...${NC}"
cp "$APP_EXECUTABLE" "$APP_MACOS/"
chmod +x "$APP_MACOS/$APP_EXECUTABLE"

# Create icon (if available)
if [ -f "$ICON_FILE" ]; then
    echo -e "${YELLOW}Adding application icon...${NC}"
    cp "$ICON_FILE" "$APP_RESOURCES/"
else
    echo -e "${YELLOW}No icon file found, creating default icon...${NC}"
    # Create a simple icon using sips if available
    if command -v sips &> /dev/null; then
        # Create a simple 512x512 icon
        sips -s format png -z 512 512 /System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericApplicationIcon.icns --out "$APP_RESOURCES/${ICON_FILE}" 2>/dev/null || true
    fi
fi

# Create PkgInfo
echo -e "${YELLOW}Creating PkgInfo...${NC}"
echo "APPL????" > "$APP_CONTENTS/PkgInfo"

# Set proper permissions
echo -e "${YELLOW}Setting permissions...${NC}"
chmod -R 755 "$APP_BUNDLE"
find "$APP_BUNDLE" -type f -exec chmod 644 {} \;
chmod +x "$APP_MACOS/$APP_EXECUTABLE"

# Sign the application (if developer certificate is available)
echo -e "${YELLOW}Attempting to sign application...${NC}"
if security find-identity -v -p codesigning 2>/dev/null | grep -q "Developer ID Application"; then
    echo "Developer certificate found, signing application..."
    codesign --force --deep --sign "Developer ID Application" "$APP_BUNDLE" 2>/dev/null || echo "Signing failed, continuing without signature"
else
    echo "No developer certificate found, skipping code signing"
fi

# Create distribution scripts
echo -e "${YELLOW}Creating distribution scripts...${NC}"
mkdir -p build/scripts

# Pre-install script
cat > build/scripts/preinstall << 'EOF'
#!/bin/bash

# Pre-install script for TheBoys Launcher
echo "Installing TheBoys Launcher..."

# Check if previous version exists
if [ -d "/Applications/TheBoys Launcher.app" ]; then
    echo "Previous installation found, backing up..."
    cp -R "/Applications/TheBoys Launcher.app" "$HOME/TheBoysLauncher_backup_$(date +%Y%m%d_%H%M%S).app" 2>/dev/null || true
fi

# Close running application
osascript -e 'tell application "TheBoys Launcher" to quit' 2>/dev/null || true

exit 0
EOF

# Post-install script
cat > build/scripts/postinstall << 'EOF'
#!/bin/bash

# Post-install script for TheBoys Launcher
echo "Post-installation setup..."

# Create user data directory if it doesn't exist
USER_DATA_DIR="$HOME/.theboys-launcher"
mkdir -p "$USER_DATA_DIR/instances"
mkdir -p "$USER_DATA_DIR/config"
mkdir -p "$USER_DATA_DIR/logs"
mkdir -p "$USER_DATA_DIR/prism"
mkdir -p "$USER_DATA_DIR/util"

# Set proper permissions
chmod -R 755 "$USER_DATA_DIR"

echo "Installation completed successfully!"

exit 0
EOF

# Make scripts executable
chmod +x build/scripts/preinstall
chmod +x build/scripts/postinstall

# Build component package
echo -e "${YELLOW}Building component package...${NC}"
pkgbuild \
    --root build/ \
    --component "$APP_BUNDLE" \
    --install-location "/Applications" \
    --scripts build/scripts \
    --identifier "$BUNDLE_ID" \
    --version "$APP_VERSION" \
    --ownership recommended \
    "build/${APP_NAME}.pkg"

# Create distribution options
echo -e "${YELLOW}Creating distribution package...${NC}"

# Create distribution definition
cat > build/distribution.xml << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="1">
    <title>${APP_NAME}</title>
    <organization>TheBoys</organization>
    <domains enable_anywhere="true" enable_currentUserHome="true" enable_localSystem="true"/>

    <options customize="never" allow-external-scripts="true" rootVolumeOnly="false"/>

    <welcome file="Welcome.html" mime-type="text/html"/>
    <license file="LICENSE.txt" mime-type="text/plain"/>
    <conclusion file="Conclusion.html" mime-type="text/html"/>

    <pkg-ref id="${BUNDLE_ID}"/>
    <choices-outline>
        <line choice="default">
            <line choice="${BUNDLE_ID}"/>
        </line>
    </choices-outline>

    <choice id="default"/>
    <choice id="${BUNDLE_ID}" visible="false">
        <pkg-ref id="${BUNDLE_ID}"/>
    </choice>

    <pkg-ref id="${BUNDLE_ID}" version="${APP_VERSION}" onConclusion="none">${APP_NAME}.pkg</pkg-ref>
</installer-gui-script>
EOF

# Create welcome file
cat > build/Welcome.html << EOF
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome to ${APP_NAME}</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; }
        h1 { color: #333; }
        p { color: #666; line-height: 1.6; }
    </style>
</head>
<body>
    <h1>Welcome to ${APP_NAME}</h1>
    <p>This installer will install ${APP_NAME} version ${APP_VERSION} on your Mac.</p>
    <p>${APP_NAME} is a modern Minecraft modpack launcher with the following features:</p>
    <ul>
        <li>Cross-platform support</li>
        <li>Modern GUI interface</li>
        <li>Automatic Java management</li>
        <li>Modpack updates and backups</li>
        <li>Integration with Prism Launcher</li>
    </ul>
    <p>Click Continue to proceed with the installation.</p>
</body>
</html>
EOF

# Create conclusion file
cat > build/Conclusion.html << EOF
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Installation Complete</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; margin: 40px; }
        h1 { color: #333; }
        p { color: #666; line-height: 1.6; }
    </style>
</head>
<body>
    <h1>Installation Complete!</h1>
    <p>${APP_NAME} has been successfully installed on your Mac.</p>
    <p>You can find the application in your Applications folder.</p>
    <p>Thank you for choosing ${APP_NAME}!</p>
</body>
</html>
EOF

# Create LICENSE file if it doesn't exist
if [ ! -f "LICENSE.txt" ]; then
    cat > build/LICENSE.txt << EOF
TheBoys Launcher - License Agreement

Copyright (c) 2024 TheBoys

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF
fi

# Build final distribution package
echo -e "${YELLOW}Building distribution package...${NC}"
productbuild \
    --distribution build/distribution.xml \
    --package-path build/ \
    --resources build/ \
    --identifier "$BUNDLE_ID" \
    --version "$APP_VERSION" \
    "dist/${APP_NAME}-${APP_VERSION}.pkg"

# Clean up temporary files
echo -e "${YELLOW}Cleaning up temporary files...${NC}"
rm -rf build/

echo ""
echo -e "${GREEN}✅ macOS installer created successfully!${NC}"
echo ""
echo "Package created: dist/${APP_NAME}-${APP_VERSION}.pkg"
echo ""
echo "Installation options:"
echo "1. Double-click the .pkg file to install"
echo "2. Follow the installer instructions"
echo "3. Find ${APP_NAME} in Applications folder"
echo ""
echo "User data location: ~/.theboys-launcher"
echo ""