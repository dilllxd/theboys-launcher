# TheBoys Launcher - Development Workflow

## Overview

This guide covers the complete development workflow for TheBoys Launcher, from quick testing to building professional installers and handling updates.

## 🚀 Quick Development Testing

### Development Commands

For rapid development and testing, use these convenient commands:

```bash
# Quick build and run (current platform only)
make run                    # Build and run in GUI mode
make run-gui               # Same as above (GUI mode)
make run-cli               # Build and run in CLI mode
make run-dev               # Build and run in development mode
make test-run              # Run with test data

# Live development with hot reload
make dev                   # Start Wails development server

# Quick build without running
make quick                 # Fast build for current platform
make build-current         # Clean build for current platform
```

### Development Modes

#### GUI Mode (Default)
```bash
make run-gui
# or
./theboys-launcher
```

#### CLI Mode
```bash
make run-cli
# or
./theboys-launcher --cli
```

#### Development Mode
```bash
make run-dev
# Enables:
# - Debug logging
# - Development features
# - Test data support
# - Skip update checks
```

#### Test Mode
```bash
make test-run
# Runs with test configuration
# Lists available modpacks
# Useful for testing CLI functionality
```

## 🏗️ Building for All Platforms

### Full Build Pipeline
```bash
# Build all platform executables
make all VERSION=v1.0.0

# Build all platforms and create installers
make installer-all VERSION=v1.0.0

# Clean + build all + installers
make clean && make all && make installer-all
```

### Platform-Specific Builds
```bash
# Windows
make build-windows VERSION=v1.0.0
make installer-windows VERSION=v1.0.0

# macOS
make build-macos VERSION=v1.0.0
make installer-macos VERSION=v1.0.0

# Linux
make build-linux VERSION=v1.0.0
make installer-linux VERSION=v1.0.0
```

## 📦 Professional Installer Creation

### Windows Installer
```bash
make installer-windows VERSION=v1.0.0
```
**Creates:** `installers/dist/TheBoysLauncher-Setup-1.0.0.exe`

**Features:**
- Professional installation wizard
- Directory selection with clear explanations
- Component selection (shortcuts, file associations)
- Automatic detection of existing installations
- Migration from portable installations
- Add/Remove Programs integration

### macOS Installer
```bash
make installer-macos VERSION=v1.0.0
```
**Creates:** `installers/dist/TheBoys Launcher-1.0.0.pkg`

**Features:**
- Native macOS package installer
- Drag-and-drop to Applications folder
- Welcome and completion screens
- License agreement
- System integration

### Linux Installers
```bash
make installer-linux VERSION=v1.0.0
```
**Creates:** Multiple installer options in `installers/dist/`

**Primary:**
- Qt-based GUI installer with professional wizard
- Directory selection with validation
- Installation options (symlinks, shortcuts)
- Real-time progress with logging

**Fallbacks:**
- Zenity-based installer (GTK)
- Dialog-based terminal installer
- Self-contained AppImage

## 🔄 Update System Integration

### Auto-Updater Features

The auto-updater works seamlessly with both installed and portable versions:

#### **Detection Logic**
```go
// Platform detection
isInstalled := platform.IsInstalled()
installPath, _ := platform.GetInstallationPath()

// Update strategy based on installation type
if isInstalled {
    // Update installed application (requires permissions)
    createUpdateScript(currentExe, updatePath, true, installPath)
} else {
    // Update portable application (simpler process)
    createUpdateScript(currentExe, updatePath, false, "")
}
```

#### **Windows Update Scenarios**
1. **Installed Application**:
   - Requires administrator privileges
   - Updates in `C:\Program Files\TheBoys Launcher\`
   - Validates permissions before updating
   - Creates backup and rollback on failure

2. **Portable Application**:
   - Updates executable in-place
   - No special permissions required
   - Simpler update process

#### **macOS/Linux Update Scenarios**
1. **Installed Application**:
   - Checks write permissions
   - Updates in installation directory
   - Handles permission errors gracefully
   - Provides clear error messages

2. **Portable Application**:
   - Updates executable directly
   - Preserves permissions
   - Works in user directories

### Update Edge Cases Handled

#### **Permission Issues**
```bash
# Linux: No write permissions to system directory
echo "Error: No write permissions to installation directory"
echo "Please run with appropriate privileges (sudo on system installations)"
echo "Or reinstall using your package manager"
```

#### **Windows: No Administrator**
```batch
echo No administrator privileges detected
echo Please run TheBoys Launcher as administrator to update
pause
```

#### **Backup and Recovery**
- Always creates backup before updating
- Automatic rollback on failure
- Preserves user configuration
- Clean up of temporary files

## 🛠️ Installer Edge Cases

### **Existing Installation Detection**

#### Windows Installer
1. **Previous Version Found**:
   - Shows current and new version
   - Option to upgrade or install alongside
   - Automatic uninstallation on upgrade

2. **Portable Installation Detected**:
   - Offers to migrate to proper installation
   - Creates migration script
   - Preserves all user data

3. **User Directory Exists**:
   - Warns about existing data
   - Confirms preservation
   - Continues with installation

#### macOS Installer
1. **Existing App in Applications**:
   - Replaces existing application
   - Preserves user data in `~/.theboys-launcher/`
   - Updates registry entries

#### Linux Installer
1. **System Installation Exists**:
   - Package manager integration handles updates
   - Manual installation prompts for overwrite

2. **User Installation Exists**:
   - Offers to upgrade or install alongside
   - Preserves user data

### **Error Handling**

#### **Windows Installer**
- **Application Running**: Attempts to close, prompts user
- **Permission Denied**: Prompts for administrator rights
- **Uninstall Failed**: Continues with parallel installation
- **Disk Space**: Checks before installation

#### **macOS Installer**
- **Gatekeeper Issues**: Provides guidance for security settings
- **Permissions**: Handles read-only file systems
- **Disk Space**: Validates available space

#### **Linux Installer**
- **Dependencies**: Checks for required packages
- **Permissions**: Validates write access
- **Display Server**: Ensures GUI environment available

## 📊 Development Testing Workflow

### **Local Testing**
```bash
# 1. Quick development cycle
make run-dev               # Test with development features

# 2. CLI testing
make run-cli               # Test CLI functionality
./theboys-launcher --cli --list-modpacks

# 3. GUI testing
make run-gui               # Test GUI functionality

# 4. Cross-platform testing
make build-all             # Build all platforms
# Test executables in respective environments
```

### **Installer Testing**
```bash
# 1. Build installers
make installer-all

# 2. Test installation scenarios:
#    - Fresh install
#    - Upgrade from previous version
#    - Parallel installation
#    - Migration from portable

# 3. Test update scenarios:
#    - Auto-update from installed version
#    - Manual update from portable version
#    - Permission denied scenarios
#    - Network failure scenarios
```

### **Integration Testing**
```bash
# 1. Test complete workflow
make clean && make all && make installer-all

# 2. Test installer in clean environment
# 3. Test application functionality
# 4. Test update mechanism
# 5. Test uninstallation process
```

## 🧪 Debugging and Troubleshooting

### **Development Debugging**
```bash
# Enable debug logging
export THEBOYS_DEV=1
export THEBOYS_LOG_LEVEL=debug
make run-dev

# Test with specific configuration
export THEBOYS_DATA_DIR=/tmp/test-launcher
make run-dev
```

### **Common Issues**

#### **Build Failures**
```bash
# Clean and rebuild
make clean && make all

# Check dependencies
make deps

# Verify Wails CLI
make install-wails
```

#### **Installer Issues**
```bash
# Check tools availability
which makensis          # Windows
which pkgbuild          # macOS
which python3           # Linux (for GUI installer)
```

#### **Update Issues**
- Check permissions on installation directory
- Verify internet connectivity
- Examine log files in `~/.theboys-launcher/logs/`

## 📝 Development Best Practices

### **Code Changes**
1. Test with `make run-dev` for quick feedback
2. Use `make test-run` for CLI functionality
3. Test GUI with `make run-gui`
4. Build for all platforms before committing

### **Installer Changes**
1. Test on clean virtual machines
2. Verify upgrade scenarios
3. Test edge cases (permissions, disk space, etc.)
4. Validate uninstallation process

### **Release Process**
1. Update version numbers
2. Build all installers: `make installer-all VERSION=v1.0.0`
3. Test installers on all platforms
4. Create GitHub release with installers
5. Update documentation

## 🔧 Development Tools

### **Required Tools**
```bash
# Go and build tools
go version
make install-wails

# Frontend development
node --version && npm --version

# Windows installer creation
makensis /VERSION

# macOS (on macOS)
pkgbuild --version
productbuild --version

# Linux installer creation
python3 --version
pip3 install --user PySide6  # For GUI installer
```

### **Optional Tools**
```bash
# Code quality
make tools    # Installs golangci-lint

# Cross-compilation (optional)
# Linux: sudo apt install gcc-mingw-w64
```

## 🎯 Quick Reference

### **Daily Development**
```bash
make run-dev       # Start development
make quick         # Quick build
make test          # Run tests
```

### **Pre-Release**
```bash
make clean && make all && make installer-all VERSION=v1.0.0
```

### **Platform-Specific Testing**
```bash
# Development on current platform
make run

# Cross-platform build
make build-all

# Installer creation
make installer-all
```

This workflow ensures efficient development, comprehensive testing, and reliable releases of TheBoys Launcher across all platforms.