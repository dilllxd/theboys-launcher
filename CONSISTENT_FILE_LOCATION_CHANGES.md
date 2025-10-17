# Consistent File Location Changes

## Summary of Changes

This document outlines the major changes made to implement consistent file locations across all platforms and add professional installer support.

## 🔧 Changes Made

### 1. **Consistent File Location Implementation**

#### Before (Platform-Specific)
- **Windows**: Files created beside executable (portable)
- **macOS**: Files in `~/.theboys-launcher/`
- **Linux**: Files in `~/.theboys-launcher/`

#### After (Consistent Across All Platforms)
- **All Platforms**: Files in `~/.theboys-launcher/`
- **Windows**: No longer portable by default (user directory like other platforms)
- **Migration**: Automatic detection and migration from legacy portable installations

### 2. **Enhanced Platform Interface**

Updated `internal/platform/interface.go` with new methods:
- `GetCustomDataDir()` - Support for custom data directories
- `CanCreateShortcut()` - Check if platform supports shortcuts
- `CreateShortcut()` - Create desktop shortcuts
- `IsInstalled()` - Check if properly installed
- `GetInstallationPath()` - Get installation location
- `RegisterInstallation()` - Register installation
- `UnregisterInstallation()` - Unregister installation

### 3. **Windows Platform Enhancements**

Updated `internal/platform/windows.go`:
- ✅ Consistent user directory behavior
- ✅ Windows Registry integration
- ✅ Shortcut creation support
- ✅ Installation detection
- ✅ Professional installation support

### 4. **Migration System**

New `internal/launcher/migration.go`:
- ✅ Automatic detection of legacy portable installations
- ✅ Safe migration with backup creation
- ✅ Progress tracking and error handling
- ✅ Migration markers to prevent re-migration
- ✅ Detailed migration reports

## 📦 Professional Installer Support

### Windows Installer (NSIS)
- ✅ Professional installation wizard
- ✅ Desktop and Start Menu shortcuts
- ✅ File associations (.modpack files)
- ✅ Add/Remove Programs integration
- ✅ Automatic updates support
- ✅ Uninstall functionality
- ✅ Registry integration

### macOS Installer (pkgbuild)
- ✅ Professional macOS application bundle
- ✅ Applications folder integration
- ✅ Code signing support
- ✅ Launchpad and Dock integration
- ✅ Proper macOS conventions
- ✅ Uninstall script

### Linux Installers (Multiple Formats)
- ✅ DEB package (Ubuntu/Debian)
- ✅ RPM package (Fedora/RHEL)
- ✅ AppImage (Universal)
- ✅ Tarball with install script
- ✅ Desktop integration
- ✅ Package manager integration

## 🗂️ File Structure

### New Consistent Structure
```
~/.theboys-launcher/          # ALL PLATFORMS
├── instances/               # Minecraft instances
├── config/                  # Configuration files
│   ├── settings.json        # User settings
│   └── modpacks.json        # Modpack configurations
├── prism/                   # Prism Launcher
├── util/                    # Utilities
│   ├── jre17/              # Java runtime
│   ├── jre21/              # Java runtime
│   └── packwiz-*           # Packwiz bootstrap
├── logs/                   # Log files
└── .migration-completed     # Migration marker
```

### Installation Directories
- **Windows**: `C:\Program Files\TheBoysLauncher\`
- **macOS**: `/Applications/TheBoys Launcher.app/`
- **Linux**: `/opt/theboys-launcher/` (or package manager location)

## 🔄 Migration Process

### Automatic Migration
1. **Detection**: Launcher detects legacy portable installation
2. **Prompt**: User is prompted to migrate data
3. **Backup**: Automatic backup created in temp directory
4. **Migration**: Data copied to `~/.theboys-launcher/`
5. **Verification**: Migration success confirmed
6. **Cleanup**: Optional cleanup of portable directory

### What Gets Migrated
- ✅ All Minecraft instances
- ✅ Configuration settings
- ✅ Downloaded Java runtimes
- ✅ Prism Launcher installation
- ✅ Modpack configurations
- ✅ Log files

## 🛠️ Build System Updates

### New Make Targets
```bash
make installer-all        # Build all installers
make installer-windows    # Build Windows installer
make installer-macos      # Build macOS installer
make installer-linux      # Build Linux installers
```

### Installer Creation Scripts
- `installers/windows-installer.nsi` - NSIS script for Windows
- `installers/macos-installer.sh` - macOS package builder
- `installers/linux-installer.sh` - Linux package builder

## 📚 Documentation Updates

### New Documentation
- `INSTALLATION_GUIDE.md` - Comprehensive installation guide
- `CONSISTENT_FILE_LOCATION_CHANGES.md` - This document

### Updated Documentation
- `BUILD_GUIDE.md` - Updated with installer building
- `PORTABLE_DEPLOYMENT.md` - Updated with new file locations

## ✅ Benefits of These Changes

### 1. **Consistency**
- Same file location across all platforms
- Predictable behavior for users
- Easier support and troubleshooting
- Better documentation

### 2. **Professional Installation**
- Proper package manager integration
- System-wide installation
- Desktop environment integration
- Professional user experience

### 3. **Migration Safety**
- Automatic detection of legacy installations
- Safe migration with backup
- No data loss during upgrade
- Smooth transition experience

### 4. **Maintainability**
- Cleaner codebase
- Platform-specific implementations isolated
- Better testing capabilities
- Easier feature additions

## 🔄 Backward Compatibility

### Legacy Support
- ✅ Automatic migration from portable installations
- ✅ All legacy CLI arguments preserved
- ✅ Configuration file compatibility
- ✅ Instance format compatibility

### Portable Mode (Optional)
Users can still use portable mode by setting environment variable:
```bash
export THEBOYS_DATA_DIR=/custom/portable/path
./theboys-launcher
```

## 🚀 Getting Started

### For Users
1. **Download** the appropriate installer for your platform
2. **Run** the installer with default settings
3. **Launch** from applications menu
4. **Migrate** if prompted (for Windows legacy users)
5. **Enjoy** consistent experience across platforms

### For Developers
1. **Build** all platforms: `make all VERSION=v1.0.0`
2. **Create installers**: `make installer-all VERSION=v1.0.0`
3. **Test** migration: Run with legacy installation present
4. **Package** for distribution: Files in `installers/dist/`

## 📊 Impact Assessment

### User Experience
- **Improved**: Consistent behavior across platforms
- **Improved**: Professional installation experience
- **Maintained**: All existing functionality
- **Enhanced**: Better integration with operating systems

### Developer Experience
- **Improved**: Cleaner, more maintainable code
- **Improved**: Better platform abstraction
- **Improved**: Comprehensive build system
- **Enhanced**: Professional distribution packages

### Maintenance
- **Reduced complexity**: Single file location logic
- **Improved testing**: Consistent test environments
- **Better documentation**: Clear installation procedures
- **Easier support**: Predictable file locations

## 🎉 Conclusion

These changes transform TheBoys Launcher from a Windows-centric portable application into a professional, cross-platform launcher with consistent behavior and proper operating system integration. Users get a better experience, developers get cleaner code, and the launcher is ready for professional distribution across all major platforms.