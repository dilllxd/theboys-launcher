#!/bin/bash

# TheBoys Launcher macOS Installer Builder
# Creates a professional macOS installer with GUI

set -e

# Configuration
APP_NAME="TheBoys Launcher"
BUNDLE_ID="com.theboys.launcher"
APP_VERSION="1.0.0"
APP_EXECUTABLE="theboys-launcher-macos"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}TheBoys Launcher - macOS GUI Installer Builder${NC}"
echo "=================================================="
echo "Version: $APP_VERSION"
echo ""

# Check if we're on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo -e "${RED}Error: This script must be run on macOS${NC}"
    echo "macOS is required to build macOS installers"
    exit 1
fi

# Check required tools
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        echo "Please install Xcode Command Line Tools: xcode-select --install"
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
    <string>AppIcon.icns</string>
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
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024 TheBoys. All rights reserved.</string>
    <key>CFBundleGetInfoString</key>
    <string>${APP_NAME} ${APP_VERSION}, Modern Minecraft Modpack Launcher</string>
</dict>
</plist>
EOF

# Copy executable
echo -e "${YELLOW}Copying executable...${NC}"
cp "$APP_EXECUTABLE" "$APP_MACOS/"
chmod +x "$APP_MACOS/$APP_EXECUTABLE"

# Create icon (if available)
if [ -f "AppIcon.icns" ]; then
    echo -e "${YELLOW}Adding application icon...${NC}"
    cp "AppIcon.icns" "$APP_RESOURCES/"
else
    echo -e "${YELLOW}Creating default icon...${NC}"
    # Create a simple icon
    if command -v sips &> /dev/null; then
        # Use system app icon as base
        cp "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/GenericApplicationIcon.icns" "$APP_RESOURCES/AppIcon.icns" 2>/dev/null || true
    fi
fi

# Create PkgInfo
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
    codesign --force --deep --sign "Developer ID Application" "$APP_BUNDLE" 2>/dev/null || echo "Warning: Code signing failed"
else
    echo "No developer certificate found, skipping code signing"
fi

# Create distribution scripts
echo -e "${YELLOW}Creating distribution scripts...${NC}"
mkdir -p build/scripts

# Post-install script
cat > build/scripts/postinstall << 'EOF'
#!/bin/bash

# Post-install script for TheBoys Launcher
echo "Setting up TheBoys Launcher..."

# Create user data directory if it doesn't exist
USER_DATA_DIR="$HOME/.theboys-launcher"
mkdir -p "$USER_DATA_DIR/instances"
mkdir -p "$USER_DATA_DIR/config"
mkdir -p "$USER_DATA_DIR/logs"
mkdir -p "$USER_DATA_DIR/prism"
mkdir -p "$USER_DATA_DIR/util"

# Set proper permissions
chmod -R 755 "$USER_DATA_DIR"

# Create welcome message
cat > "$USER_DATA_DIR/welcome.txt" << 'WELCOME'
Welcome to TheBoys Launcher!

Your user data is stored in this directory:
~/.theboys-launcher/

This includes:
- Minecraft instances
- Configuration settings
- Downloaded files
- Log files

Thank you for choosing TheBoys Launcher!
WELCOME

echo "Installation completed successfully!"
echo "TheBoys Launcher has been installed in your Applications folder."
echo "You can now launch it from Finder or your Applications folder."

exit 0
EOF

# Uninstall script
cat > build/scripts/uninstall << 'EOF'
#!/bin/bash

# Uninstall script for TheBoys Launcher
echo "Uninstalling TheBoys Launcher..."

# Remove the application from Applications
if [ -d "/Applications/TheBoys Launcher.app" ]; then
    rm -rf "/Applications/TheBoys Launcher.app"
    echo "Removed application from Applications folder"
fi

# Ask user if they want to remove user data
echo ""
echo "Do you want to remove all user data (instances, settings, etc.)?"
echo "This includes ~/.theboys-launcher/ and all its contents."
echo ""
read -p "Remove user data? (y/N): " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -d "$HOME/.theboys-launcher" ]; then
        rm -rf "$HOME/.theboys-launcher"
        echo "Removed user data directory"
    fi
else
    echo "User data preserved in ~/.theboys-launcher/"
fi

# Clean up launch services
/usr/bin/launchservicesutil unregister -domain local -file "/Applications/TheBoys Launcher.app" 2>/dev/null || true

echo "TheBoys Launcher has been uninstalled."
echo "Thank you for using TheBoys Launcher!"

exit 0
EOF

# Make scripts executable
chmod +x build/scripts/postinstall
chmod +x build/scripts/uninstall

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
    --install-location "/Applications" \
    "build/${APP_NAME}.pkg"

# Create distribution resources
echo -e "${YELLOW}Creating distribution resources...${NC}"
mkdir -p build/resources

# Create welcome text
cat > build/resources/Welcome.txt << EOF
TheBoys Launcher

Modern Minecraft Modpack Launcher

Version ${APP_VERSION}

This installer will install TheBoys Launcher on your Mac.

TheBoys Launcher features:
• Modern graphical user interface
• Automatic Java runtime management
• Support for multiple modpack sources
• Automatic updates and backups
• Integration with Prism Launcher

Your saved games and settings will be stored in:
~/.theboys-launcher/

Click Continue to proceed with the installation.
EOF

# Create conclusion text
cat > build/resources/Conclusion.txt << EOF
Installation Complete!

TheBoys Launcher has been successfully installed on your Mac.

You can find TheBoys Launcher in your Applications folder.

Your user data is stored in: ~/.theboys-launcher/

Thank you for choosing TheBoys Launcher!

For support and updates:
https://github.com/dilllxd/theboys-launcher

Click Close to finish the installation.
EOF

# Create license file
cat > build/resources/License.txt << EOF
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

# Create distribution definition
echo -e "${YELLOW}Creating distribution package...${NC}"
cat > build/distribution.xml << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="1">
    <title>${APP_NAME}</title>
    <organization>TheBoys</organization>
    <domains enable_anywhere="true" enable_currentUserHome="true" enable_localSystem="true"/>

    <options customize="never" allow-external-scripts="true" rootVolumeOnly="false"/>

    <welcome file="Welcome.txt" mime-type="text/plain"/>
    <license file="License.txt" mime-type="text/plain"/>
    <conclusion file="Conclusion.txt" mime-type="text/plain"/>

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

# Build final distribution package
echo -e "${YELLOW}Building final distribution package...${NC}"
productbuild \
    --distribution build/distribution.xml \
    --package-path build/ \
    --resources build/resources/ \
    --identifier "$BUNDLE_ID" \
    --version "$APP_VERSION" \
    --sign "Developer ID Installer: TheBoys (TEAM_ID)" 2>/dev/null || \
    productbuild \
        --distribution build/distribution.xml \
        --package-path build/ \
        --resources build/resources/ \
        --identifier "$BUNDLE_ID" \
        --version "$APP_VERSION" \
        "dist/${APP_NAME}-${APP_VERSION}.pkg"

# Clean up temporary files
echo -e "${YELLOW}Cleaning up temporary files...${NC}"
rm -rf build/

echo ""
echo -e "${GREEN}✅ macOS GUI installer created successfully!${NC}"
echo ""
echo "Package created: dist/${APP_NAME}-${APP_VERSION}.pkg"
echo ""
echo "Installation:"
echo "1. Double-click the .pkg file"
echo "2. Follow the installation wizard"
echo "3. Find TheBoys Launcher in Applications"
echo ""
echo "Features:"
echo "• Professional installation wizard"
echo "• User-friendly interface"
echo "• Automatic setup of user directories"
echo "• Easy uninstallation"
echo "• Proper macOS integration"
echo ""
echo "User data location: ~/.theboys-launcher/"
echo ""