#!/bin/bash

# TheBoys Launcher Linux GUI Installer Wrapper
# Creates a user-friendly GUI installer for Linux

set -e

# Configuration
APP_NAME="TheBoys Launcher"
APP_VERSION="1.0.0"
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
esac

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}TheBoys Launcher - Linux GUI Installer${NC}"
echo "============================================"
echo "Version: $APP_VERSION"
echo "Architecture: $ARCH"
echo ""

# Check executable
EXECUTABLE="theboys-launcher-linux-${ARCH}"
if [ ! -f "$EXECUTABLE" ]; then
    echo -e "${RED}Error: Executable '$EXECUTABLE' not found${NC}"
    echo "Please build the application first or provide the correct executable."
    exit 1
fi

# Check for GUI environment
check_gui() {
    if [ -z "$DISPLAY" ]; then
        echo -e "${RED}Error: No display detected. This installer requires a graphical environment.${NC}"
        echo "Please run this installer from a graphical desktop environment."
        exit 1
    fi
}

# Check dependencies
check_dependencies() {
    echo -e "${YELLOW}Checking dependencies...${NC}"

    # Check for Python 3
    if ! command -v python3 &> /dev/null; then
        echo -e "${RED}Error: Python 3 is required but not installed${NC}"
        echo "Please install Python 3:"
        echo "  Ubuntu/Debian: sudo apt install python3 python3-pip"
        echo "  Fedora: sudo dnf install python3 python3-pip"
        echo "  Arch: sudo pacman -S python python-pip"
        exit 1
    fi

    # Check for PySide6
    if ! python3 -c "import PySide6" &> /dev/null; then
        echo -e "${YELLOW}PySide6 not found. Installing...${NC}"
        if command -v pip3 &> /dev/null; then
            pip3 install --user PySide6
        elif command -v pip &> /dev/null; then
            pip install --user PySide6
        else
            echo -e "${RED}Error: pip not found. Please install pip3 or pip${NC}"
            echo "  Ubuntu/Debian: sudo apt install python3-pip"
            echo "  Fedora: sudo dnf install python3-pip"
            echo "  Arch: sudo pacman -S python-pip"
            exit 1
        fi
    fi
}

# Create temporary installer directory
create_installer() {
    echo -e "${YELLOW}Creating installer package...${NC}"

    TEMP_DIR=$(mktemp -d)
    INSTALLER_DIR="$TEMP_DIR/theboys-launcher-installer"

    mkdir -p "$INSTALLER_DIR"

    # Copy installer script
    cp linux-setup.py "$INSTALLER_DIR/"
    chmod +x "$INSTALLER_DIR/linux-setup.py"

    # Copy executable
    cp "$EXECUTABLE" "$INSTALLER_DIR/"
    chmod +x "$INSTALLER_DIR/$EXECUTABLE"

    # Create launcher script
    cat > "$INSTALLER_DIR/install" << 'EOF'
#!/bin/bash
cd "$(dirname "$0")"
python3 linux-setup.py "./$(basename theboys-launcher-linux-*)"
EOF
    chmod +x "$INSTALLER_DIR/install"

    # Create README
    cat > "$INSTALLER_DIR/README.txt" << EOF
TheBoys Launcher Linux Installer

To start the installation, run:
./install

Or run directly with:
python3 linux-setup.py ./theboys-launcher-linux-$ARCH

System Requirements:
- Linux with graphical desktop environment
- Python 3.6 or higher
- 500MB disk space

For support and updates:
https://github.com/dilllxd/theboys-launcher
EOF

    # Create package
    PACKAGE_NAME="TheBoysLauncher-Linux-Installer-$APP_VERSION-$ARCH.tar.gz"
    tar -czf "$PACKAGE_NAME" -C "$TEMP_DIR" theboys-launcher-installer

    # Cleanup
    rm -rf "$TEMP_DIR"

    echo -e "${GREEN}✓ GUI installer package created: $PACKAGE_NAME${NC}"
}

# Create alternative installers for different environments
create_alternatives() {
    echo -e "${YELLOW}Creating alternative installers...${NC}"

    # Create Zenity-based installer (for systems without Python/Qt)
    create_zenity_installer

    # Create dialog-based installer (terminal GUI)
    create_dialog_installer
}

create_zenity_installer() {
    if command -v zenity &> /dev/null; then
        echo -e "${YELLOW}Creating Zenity-based installer...${NC}"

        cat > "theboys-launcher-zenity-installer.sh" << 'EOF'
#!/bin/bash

# TheBoys Launcher Zenity Installer
# Fallback installer using Zenity for basic GUI

set -e

APP_NAME="TheBoys Launcher"
DEFAULT_INSTALL_DIR="/opt/theboys-launcher"

# Welcome dialog
zenity --question \
    --title="$APP_NAME Installer" \
    --text="Welcome to $APP_NAME!\n\nThis installer will guide you through installing $APP_NAME on your system.\n\nDo you want to continue?" \
    --width=400 \
    --height=200 || exit 0

# License dialog
zenity --text-info \
    --title="License Agreement" \
    --filename=<(echo "TheBoys Launcher - License Agreement

Copyright (c) 2024 TheBoys

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the 'Software'), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.") \
    --width=600 \
    --height=400 \
    --checkbox="I accept the terms of the license agreement" || exit 0

# Installation directory
INSTALL_DIR=$(zenity --file-selection \
    --directory \
    --title="Select Installation Directory" \
    --filename="$DEFAULT_INSTALL_DIR")

if [ -z "$INSTALL_DIR" ]; then
    exit 0
fi

# Installation options
CREATE_SYMLINK=$(zenity --question \
    --title="Installation Options" \
    --text="Do you want to create a command-line symlink?" \
    --width=400 \
    --height=200; echo $?)

CREATE_DESKTOP=$(zenity --question \
    --title="Installation Options" \
    --text="Do you want to create a desktop shortcut?" \
    --width=400 \
    --height=200; echo $?)

# Progress dialog
(
    echo "10"
    echo "# Creating directories..."
    mkdir -p "$INSTALL_DIR"

    echo "30"
    echo "# Installing application files..."
    cp theboys-launcher-linux-* "$INSTALL_DIR/theboys-launcher"
    chmod +x "$INSTALL_DIR/theboys-launcher"

    echo "50"
    echo "# Creating desktop entry..."
    mkdir -p "$HOME/.local/share/applications"
    cat > "$HOME/.local/share/applications/theboys-launcher.desktop" << EOL
[Desktop Entry]
Version=1.0
Type=Application
Name=TheBoys Launcher
Comment=Modern Minecraft Modpack Launcher
Exec=$INSTALL_DIR/theboys-launcher
Icon=theboys-launcher
Terminal=false
Categories=Game;Utility;
EOL

    echo "70"
    echo "# Setting up user data directory..."
    mkdir -p "$HOME/.theboys-launcher"/{instances,config,logs,prism,util}

    if [ "$CREATE_SYMLINK" = "0" ]; then
        echo "# Creating command-line symlink..."
        mkdir -p "$HOME/.local/bin"
        ln -sf "$INSTALL_DIR/theboys-launcher" "$HOME/.local/bin/theboys-launcher"
    fi

    echo "90"
    echo "# Updating desktop database..."
    update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true

    echo "100"
    echo "# Installation completed!"
) | zenity --progress \
    --title="Installing $APP_NAME" \
    --text="Installing..." \
    --width=400 \
    --height=200 \
    --auto-close

# Completion dialog
if [ $? -eq 0 ]; then
    zenity --info \
        --title="Installation Complete" \
        --text="$APP_NAME has been successfully installed!\n\nYou can now launch it from your applications menu." \
        --width=400 \
        --height=200

    # Ask if user wants to launch
    if zenity --question \
        --title="Launch Application" \
        --text="Do you want to launch $APP_NAME now?" \
        --width=400 \
        --height=200; then
        "$INSTALL_DIR/theboys-launcher" &
    fi
else
    zenity --error \
        --title="Installation Failed" \
        --text="The installation was cancelled or failed." \
        --width=400 \
        --height=200
fi
EOF
        chmod +x "theboys-launcher-zenity-installer.sh"
        echo -e "${GREEN}✓ Zenity installer created: theboys-launcher-zenity-installer.sh${NC}"
    fi
}

create_dialog_installer() {
    if command -v dialog &> /dev/null; then
        echo -e "${YELLOW}Creating dialog-based installer...${NC}"

        cat > "theboys-launcher-dialog-installer.sh" << 'EOF'
#!/bin/bash

# TheBoys Launcher Dialog Installer
# Terminal GUI installer using dialog

set -e

APP_NAME="TheBoys Launcher"
DEFAULT_INSTALL_DIR="/opt/theboys-launcher"

# Check if dialog is available
if ! command -v dialog &> /dev/null; then
    echo "Error: dialog is not installed"
    echo "Install it with: sudo apt install dialog"
    exit 1
fi

# Welcome dialog
dialog --title "Welcome" \
    --msgbox "Welcome to $APP_NAME!\n\nThis installer will guide you through installing $APP_NAME on your system.\n\nPress OK to continue." \
    10 50

# License dialog
dialog --title "License Agreement" \
    --msgbox "TheBoys Launcher - License Agreement\n\nCopyright (c) 2024 TheBoys\n\nPermission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the 'Software'), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:\n\nThe above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.\n\nTHE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE." \
    20 70

# Installation directory
INSTALL_DIR=$(dialog --title "Installation Directory" \
    --inputbox "Enter the installation directory:" \
    10 50 "$DEFAULT_INSTALL_DIR" \
    3>&1 1>&2 2>&3 3>&-)

if [ $? -ne 0 ] || [ -z "$INSTALL_DIR" ]; then
    exit 0
fi

# Installation options
dialog --title "Installation Options" \
    --checklist "Select additional options:" \
    15 50 2 \
    1 "Create command-line symlink" ON \
    2 "Create desktop shortcut" ON \
    3>&1 1>&2 2>&3 3>&-

OPTIONS=$?

# Installation
{
    echo "XXX"
    echo "10"
    echo "Creating directories..."
    echo "XXX"

    mkdir -p "$INSTALL_DIR"

    echo "XXX"
    echo "30"
    echo "Installing application files..."
    echo "XXX"

    cp theboys-launcher-linux-* "$INSTALL_DIR/theboys-launcher"
    chmod +x "$INSTALL_DIR/theboys-launcher"

    echo "XXX"
    echo "50"
    echo "Creating desktop entry..."
    echo "XXX"

    mkdir -p "$HOME/.local/share/applications"
    cat > "$HOME/.local/share/applications/theboys-launcher.desktop" << EOL
[Desktop Entry]
Version=1.0
Type=Application
Name=TheBoys Launcher
Comment=Modern Minecraft Modpack Launcher
Exec=$INSTALL_DIR/theboys-launcher
Icon=theboys-launcher
Terminal=false
Categories=Game;Utility;
EOL

    echo "XXX"
    echo "70"
    echo "Setting up user data directory..."
    echo "XXX"

    mkdir -p "$HOME/.theboys-launcher"/{instances,config,logs,prism,util}

    echo "XXX"
    echo "90"
    echo "Finalizing installation..."
    echo "XXX"

    update-desktop-database "$HOME/.local/share/applications" 2>/dev/null || true

    echo "XXX"
    echo "100"
    echo "Installation completed!"
    echo "XXX"
} | dialog --title "Installing" \
    --gauge "Installing $APP_NAME..." \
    8 50

# Completion dialog
dialog --title "Installation Complete" \
    --msgbox "$APP_NAME has been successfully installed!\n\nYou can now launch it from your applications menu by running '$INSTALL_DIR/theboys-launcher'." \
    10 50

clear
echo "Installation completed successfully!"
echo "TheBoys Launcher is installed in: $INSTALL_DIR"
echo "Run it with: $INSTALL_DIR/theboys-launcher"
EOF
        chmod +x "theboys-launcher-dialog-installer.sh"
        echo -e "${GREEN}✓ Dialog installer created: theboys-launcher-dialog-installer.sh${NC}"
    fi
}

# Create AppImage (self-contained)
create_appimage() {
    echo -e "${YELLOW}Creating AppImage installer...${NC}"

    if command -v appimagetool &> /dev/null; then
        APPDIR="TheBoysLauncher.AppDir"
        mkdir -p "$APPDIR/usr/bin"
        mkdir -p "$APPDIR/usr/share/applications"
        mkdir -p "$APPDIR/usr/share/icons/hicolor/256x256/apps"

        # Copy executable
        cp "$EXECUTABLE" "$APPDIR/usr/bin/theboys-launcher"
        chmod +x "$APPDIR/usr/bin/theboys-launcher"

        # Create desktop file
        cat > "$APPDIR/usr/share/applications/theboys-launcher.desktop" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=TheBoys Launcher
Comment=Modern Minecraft Modpack Launcher
Exec=theboys-launcher
Icon=theboys-launcher
Terminal=false
Categories=Game;Utility;
EOF

        # Create AppRun
        cat > "$APPDIR/AppRun" << 'EOF'
#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")")"
exec "$HERE/usr/bin/theboys-launcher" "$@"
EOF
        chmod +x "$APPDIR/AppRun"

        # Create AppImage
        ./appimagetool "$APPDIR" "TheBoysLauncher-$APP_VERSION-$ARCH.AppImage"
        rm -rf "$APPDIR"

        echo -e "${GREEN}✓ AppImage created: TheBoysLauncher-$APP_VERSION-$ARCH.AppImage${NC}"
    else
        echo -e "${YELLOW}AppImageTool not found, skipping AppImage creation${NC}"
    fi
}

# Main execution
main() {
    echo -e "${BLUE}Starting Linux GUI installer creation...${NC}"

    check_gui
    check_dependencies

    # Create main GUI installer
    create_installer

    # Create alternatives
    create_alternatives

    # Create AppImage
    create_appimage

    echo ""
    echo -e "${GREEN}✅ All Linux installers created successfully!${NC}"
    echo ""
    echo "Available installers:"
    echo "1. GUI Installer (Qt): $PACKAGE_NAME"
    echo "2. Zenity Installer: theboys-launcher-zenity-installer.sh"
    echo "3. Dialog Installer: theboys-launcher-dialog-installer.sh"
    if [ -f "TheBoysLauncher-$APP_VERSION-$ARCH.AppImage" ]; then
        echo "4. AppImage: TheBoysLauncher-$APP_VERSION-$ARCH.AppImage"
    fi
    echo ""
    echo "Installation instructions:"
    echo "GUI Installer (Recommended):"
    echo "1. Extract: tar -xzf $PACKAGE_NAME"
    echo "2. Run: cd theboys-launcher-installer && ./install"
    echo ""
    echo "Alternative installers:"
    echo "- Zenity: ./theboys-launcher-zenity-installer.sh"
    echo "- Dialog: ./theboys-launcher-dialog-installer.sh"
    echo "- AppImage: ./TheBoysLauncher-$APP_VERSION-$ARCH.AppImage"
    echo ""
}

# Run main function
main "$@"