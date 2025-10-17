#!/bin/bash

# TheBoys Launcher Linux Installer Script
# Creates a proper Linux installer package (AppImage, deb, and rpm)

set -e

# Configuration
APP_NAME="TheBoys Launcher"
APP_ID="theboys-launcher"
APP_VERSION="1.0.0"
APP_EXECUTABLE="theboys-launcher-linux"
ICON_FILE="AppIcon.png"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}TheBoys Launcher - Linux Installer Builder${NC}"
echo "============================================="
echo "Version: $APP_VERSION"
echo ""

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo -e "${YELLOW}Building for architecture: $ARCH${NC}"

# Check required tools
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        echo "Please install $1 and try again"
        echo "On Ubuntu/Debian: sudo apt install $1"
        echo "On Fedora: sudo dnf install $1"
        echo "On Arch: sudo pacman -S $1"
        exit 1
    fi
}

# Check for executable
if [ ! -f "$APP_EXECUTABLE-$ARCH" ]; then
    echo -e "${RED}Error: Executable '$APP_EXECUTABLE-$ARCH' not found${NC}"
    echo "Please build the application first using: make build-linux"
    exit 1
fi

# Clean previous builds
echo -e "${YELLOW}Cleaning previous builds...${NC}"
rm -rf build/
rm -rf dist/
mkdir -p build/
mkdir -p dist/

echo -e "${YELLOW}Building Linux packages...${NC}"
echo ""

# Function to create desktop file
create_desktop_file() {
    cat > "$1" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=${APP_NAME}
Comment=Modern Minecraft Modpack Launcher
Exec=$2
Icon=${APP_ID}
Terminal=false
Categories=Game;Utility;
StartupWMClass=${APP_NAME}
MimeType=x-scheme-handler/theboys;
EOF
}

# Function to create AppImage
create_appimage() {
    echo -e "${YELLOW}Creating AppImage...${NC}"

    APPDIR="build/${APP_NAME}.AppDir"
    mkdir -p "$APPDIR"

    # Create AppDir structure
    mkdir -p "$APPDIR/usr/bin"
    mkdir -p "$APPDIR/usr/share/applications"
    mkdir -p "$APPDIR/usr/share/icons/hicolor/256x256/apps"
    mkdir -p "$APPDIR/usr/lib"

    # Copy executable
    cp "$APP_EXECUTABLE-$ARCH" "$APPDIR/usr/bin/${APP_ID}"
    chmod +x "$APPDIR/usr/bin/${APP_ID}"

    # Create desktop file
    create_desktop_file "$APPDIR/usr/share/applications/${APP_ID}.desktop" "/usr/bin/${APP_ID}"

    # Copy icon (create a simple one if not available)
    if [ -f "$ICON_FILE" ]; then
        cp "$ICON_FILE" "$APPDIR/usr/share/icons/hicolor/256x256/apps/${APP_ID}.png"
    else
        # Create a simple icon
        convert -size 256x256 xc:blue -fill white -gravity center -pointsize 32 -annotate +0+0 "TBL" \
            "$APPDIR/usr/share/icons/hicolor/256x256/apps/${APP_ID}.png" 2>/dev/null || \
            echo "Warning: Could not create icon (ImageMagick not available)"
    fi

    # Create AppRun
    cat > "$APPDIR/AppRun" << EOF
#!/bin/bash
HERE="\$(dirname "\$(readlink -f "\${0}")")"
exec "\$HERE/usr/bin/${APP_ID}" "\$@"
EOF
    chmod +x "$APPDIR/AppRun"

    # Download appimagetool if not available
    if [ ! -f "appimagetool-x86_64.AppImage" ]; then
        echo -e "${YELLOW}Downloading appimagetool...${NC}"
        wget -q "https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage"
        chmod +x appimagetool-x86_64.AppImage
    fi

    # Create AppImage
    ./appimagetool-x86_64.AppImage "$APPDIR" "dist/${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage"

    echo -e "${GREEN}✓ AppImage created: dist/${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage${NC}"
}

# Function to create DEB package
create_deb() {
    echo -e "${YELLOW}Creating DEB package...${NC}"

    DEB_DIR="build/deb"
    mkdir -p "$DEB_DIR/DEBIAN"
    mkdir -p "$DEB_DIR/usr/bin"
    mkdir -p "$DEB_DIR/usr/share/applications"
    mkdir -p "$DEB_DIR/usr/share/icons/hicolor/256x256/apps"
    mkdir -p "$DEB_DIR/usr/share/doc/${APP_ID}"

    # Copy executable
    cp "$APP_EXECUTABLE-$ARCH" "$DEB_DIR/usr/bin/${APP_ID}"
    chmod +x "$DEB_DIR/usr/bin/${APP_ID}"

    # Create desktop file
    create_desktop_file "$DEB_DIR/usr/share/applications/${APP_ID}.desktop" "${APP_ID}"

    # Copy icon
    if [ -f "$ICON_FILE" ]; then
        cp "$ICON_FILE" "$DEB_DIR/usr/share/icons/hicolor/256x256/apps/${APP_ID}.png"
    fi

    # Create documentation
    cat > "$DEB_DIR/usr/share/doc/${APP_ID}/copyright" << EOF
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: ${APP_NAME}
Source: https://github.com/dilllxd/theboys-launcher

Files: *
Copyright: 2024 TheBoys
License: MIT

License: MIT
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:
 .
 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.
 .
 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
EOF

    # Create control file
    local size=$(du -s "$DEB_DIR" | cut -f1)
    cat > "$DEB_DIR/DEBIAN/control" << EOF
Package: ${APP_ID}
Version: ${APP_VERSION}
Architecture: $ARCH
Maintainer: TheBoys <contact@theboys.dev>
Depends: libc6, libx11-6, libxrandr2, libxinerama1, libxcursor1, libxi6, libgl1-mesa-glx
Section: games
Priority: optional
Homepage: https://github.com/dilllxd/theboys-launcher
Description: Modern Minecraft Modpack Launcher
 ${APP_NAME} is a modern, cross-platform Minecraft modpack launcher
 with a graphical interface, automatic Java management, and support for
 multiple modpack sources.
 .
 Features:
  * Modern GUI interface
  * Cross-platform support
  * Automatic Java runtime management
  * Modpack updates and backups
  * Integration with Prism Launcher
Installed-Size: $size
EOF

    # Create postinst script
    cat > "$DEB_DIR/DEBIAN/postinst" << 'EOF'
#!/bin/bash
set -e

# Create user data directory
mkdir -p "$HOME/.theboys-launcher"/{instances,config,logs,prism,util}
chmod -R 755 "$HOME/.theboys-launcher"

# Update desktop database
if command -v update-desktop-database &> /dev/null; then
    update-desktop-database -q /usr/share/applications || true
fi

# Update icon cache
if command -v gtk-update-icon-cache &> /dev/null; then
    gtk-update-icon-cache -q -t -f /usr/share/icons/hicolor || true
fi

exit 0
EOF
    chmod +x "$DEB_DIR/DEBIAN/postinst"

    # Build DEB package
    dpkg-deb --build "$DEB_DIR" "dist/${APP_ID}_${APP_VERSION}_${ARCH}.deb"

    echo -e "${GREEN}✓ DEB package created: dist/${APP_ID}_${APP_VERSION}_${ARCH}.deb${NC}"
}

# Function to create RPM package
create_rpm() {
    echo -e "${YELLOW}Creating RPM package...${NC}"

    if ! command -v rpmbuild &> /dev/null; then
        echo -e "${YELLOW}rpmbuild not found, skipping RPM package${NC}"
        return
    fi

    # Create rpmbuild directory structure
    mkdir -p "$HOME/rpmbuild/SOURCES"
    mkdir -p "$HOME/rpmbuild/SPECS"
    mkdir -p "$HOME/rpmbuild/BUILD"
    mkdir -p "$HOME/rpmbuild/RPMS"

    # Create source tarball
    TAR_NAME="${APP_ID}-${APP_VERSION}"
    mkdir -p "build/$TAR_NAME"
    cp "$APP_EXECUTABLE-$ARCH" "build/$TAR_NAME/${APP_ID}"

    if [ -f "$ICON_FILE" ]; then
        cp "$ICON_FILE" "build/$TAR_NAME/${APP_ID}.png"
    fi

    # Create spec file
    cat > "$HOME/rpmbuild/SPECS/${APP_ID}.spec" << EOF
Name:           ${APP_ID}
Version:        ${APP_VERSION}
Release:        1%{?dist}
Summary:        Modern Minecraft Modpack Launcher
License:        MIT
URL:            https://github.com/dilllxd/theboys-launcher
Source0:        ${TAR_NAME}.tar.gz

BuildRequires:  desktop-file-utils
Requires:       glibc, libX11, libXrandr, libXinerama, libXcursor, libXi, mesa-libGL

%description
${APP_NAME} is a modern, cross-platform Minecraft modpack launcher
with a graphical interface, automatic Java management, and support for
multiple modpack sources.

Features:
 * Modern GUI interface
 * Cross-platform support
 * Automatic Java runtime management
 * Modpack updates and backups
 * Integration with Prism Launcher

%prep
%autosetup

%build
# No build required, binary package

%install
rm -rf %{buildroot}
install -D -m 755 ${APP_ID} %{buildroot}%{_bindir}/${APP_ID}

# Install desktop file
mkdir -p %{buildroot}%{_datadir}/applications
cat > %{buildroot}%{_datadir}/applications/${APP_ID}.desktop << 'DESKTOP_EOF'
[Desktop Entry]
Version=1.0
Type=Application
Name=${APP_NAME}
Comment=Modern Minecraft Modpack Launcher
Exec=${APP_ID}
Icon=${APP_ID}
Terminal=false
Categories=Game;Utility;
StartupWMClass=${APP_NAME}
DESKTOP_EOF

# Install icon
if [ -f ${APP_ID}.png ]; then
    install -D -m 644 ${APP_ID}.png %{buildroot}%{_datadir}/icons/hicolor/256x256/apps/${APP_ID}.png
fi

%files
%{_bindir}/${APP_ID}
%{_datadir}/applications/${APP_ID}.desktop
%{_datadir}/icons/hicolor/256x256/apps/${APP_ID}.png

%post
# Create user data directory
mkdir -p %{_localstatedir}/lib/${APP_ID}/instances
mkdir -p %{_localstatedir}/lib/${APP_ID}/config
mkdir -p %{_localstatedir}/lib/${APP_ID}/logs
mkdir -p %{_localstatedir}/lib/${APP_ID}/prism
mkdir -p %{_localstatedir}/lib/${APP_ID}/util

# Update desktop database
update-desktop-database &> /dev/null || true

# Update icon cache
gtk-update-icon-cache &> /dev/null || true

%postun
if [ \$1 -eq 0 ]; then
    # Update desktop database
    update-desktop-database &> /dev/null || true

    # Update icon cache
    gtk-update-icon-cache &> /dev/null || true
fi

%changelog
* $(date +'%a %b %d %Y') TheBoys <contact@theboys.dev> - ${APP_VERSION}-1
- Initial RPM release
EOF

    # Create tarball
    cd build
    tar -czf "$HOME/rpmbuild/SOURCES/${TAR_NAME}.tar.gz" "$TAR_NAME"
    cd ..

    # Build RPM
    rpmbuild -ba "$HOME/rpmbuild/SPECS/${APP_ID}.spec"

    # Copy built RPM to dist
    find "$HOME/rpmbuild/RPMS" -name "*.rpm" -exec cp {} dist/ \;

    echo -e "${GREEN}✓ RPM package created${NC}"
}

# Function to create tar.gz package
create_tarball() {
    echo -e "${YELLOW}Creating tar.gz package...${NC}"

    TARBALL_DIR="build/${APP_NAME}-${APP_VERSION}-${ARCH}"
    mkdir -p "$TARBALL_DIR"

    # Copy executable
    cp "$APP_EXECUTABLE-$ARCH" "$TARBALL_DIR/${APP_ID}"
    chmod +x "$TARBALL_DIR/${APP_ID}"

    # Create installation script
    cat > "$TARBALL_DIR/install.sh" << EOF
#!/bin/bash

# TheBoys Launcher Installation Script

set -e

APP_NAME="${APP_NAME}"
APP_ID="${APP_ID}"
INSTALL_DIR="/opt/\${APP_ID}"

echo "Installing \${APP_NAME}..."

# Check if running as root
if [ "\$EUID" -ne 0 ]; then
    echo "This script must be run as root (sudo)"
    echo "Usage: sudo ./install.sh"
    exit 1
fi

# Create installation directory
mkdir -p "\$INSTALL_DIR"

# Copy executable
cp "\${APP_ID}" "\$INSTALL_DIR/"
chmod +x "\$INSTALL_DIR/\${APP_ID}"

# Create symlink
ln -sf "\$INSTALL_DIR/\${APP_ID}" "/usr/local/bin/\${APP_ID}"

# Create desktop entry
mkdir -p "/usr/share/applications"
cat > "/usr/share/applications/\${APP_ID}.desktop" << 'DESKTOP_EOF'
[Desktop Entry]
Version=1.0
Type=Application
Name=\${APP_NAME}
Comment=Modern Minecraft Modpack Launcher
Exec=\${APP_ID}
Icon=\${APP_ID}
Terminal=false
Categories=Game;Utility;
StartupWMClass=\${APP_NAME}
DESKTOP_EOF

# Copy icon if available
if [ -f "\${APP_ID}.png" ]; then
    mkdir -p "/usr/share/icons/hicolor/256x256/apps"
    cp "\${APP_ID}.png" "/usr/share/icons/hicolor/256x256/apps/\${APP_ID}.png"
fi

echo "Installation completed successfully!"
echo "You can now run \${APP_NAME} from your applications menu or by typing '\${APP_ID}'"

EOF
    chmod +x "$TARBALL_DIR/install.sh"

    # Create README
    cat > "$TARBALL_DIR/README.txt" << EOF
${APP_NAME} v${APP_VERSION}

Installation:
1. Run: sudo ./install.sh
2. Launch from applications menu or run '${APP_ID}' in terminal

Uninstallation:
1. Run: sudo rm -rf /opt/${APP_ID}
2. Run: sudo rm /usr/local/bin/${APP_ID}
3. Run: sudo rm /usr/share/applications/${APP_ID}.desktop

User data is stored in: ~/.theboys-launcher
EOF

    # Copy icon if available
    if [ -f "$ICON_FILE" ]; then
        cp "$ICON_FILE" "$TARBALL_DIR/${APP_ID}.png"
    fi

    # Create tarball
    cd build
    tar -czf "../dist/${APP_NAME}-${APP_VERSION}-${ARCH}.tar.gz" "$TARBALL_DIR"
    cd ..

    echo -e "${GREEN}✓ Tarball created: dist/${APP_NAME}-${APP_VERSION}-${ARCH}.tar.gz${NC}"
}

# Build packages based on available tools
create_appimage
create_deb
create_rpm
create_tarball

echo ""
echo -e "${GREEN}✅ All Linux packages created successfully!${NC}"
echo ""
echo "Available packages:"
echo "- AppImage: dist/${APP_NAME}-${APP_VERSION}-${ARCH}.AppImage"
echo "- DEB: dist/${APP_ID}_${APP_VERSION}_${ARCH}.deb"
echo "- Tarball: dist/${APP_NAME}-${APP_VERSION}-${ARCH}.tar.gz"
if [ -f "dist/${APP_ID}-${APP_VERSION}-1.*.rpm" ]; then
    echo "- RPM: dist/${APP_ID}-${APP_VERSION}-1.*.rpm"
fi
echo ""
echo "Installation instructions:"
echo ""
echo "AppImage (Recommended):"
echo "1. Download and make executable: chmod +x *.AppImage"
echo "2. Run directly: ./TheBoysLauncher-*.AppImage"
echo ""
echo "DEB (Ubuntu/Debian):"
echo "1. Install: sudo dpkg -i *.deb"
echo "2. Fix dependencies: sudo apt-get install -f"
echo ""
echo "RPM (Fedora/RHEL):"
echo "1. Install: sudo rpm -i *.rpm"
echo ""
echo "Tarball (Universal):"
echo "1. Extract: tar -xzf *.tar.gz"
echo "2. Install: cd extracted_folder && sudo ./install.sh"
echo ""
echo "User data location: ~/.theboys-launcher"
echo ""