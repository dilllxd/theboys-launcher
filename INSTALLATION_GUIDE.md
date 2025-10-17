# TheBoys Launcher - Installation Guide

## Overview

TheBoys Launcher provides multiple installation methods for optimal user experience across all platforms. All installations use **consistent file locations** in the user's home directory.

## File Locations (Consistent Across All Platforms)

### User Data Directory
```
~/.theboys-launcher/          # All platforms use this location
├── instances/               # Minecraft instances
├── config/                  # Configuration files
│   ├── settings.json        # User settings
│   └── modpacks.json        # Modpack configurations
├── prism/                   # Prism Launcher
├── util/                    # Utilities
│   ├── jre17/              # Java runtime
│   ├── jre21/              # Java runtime
│   ├── packwiz-installer-bootstrap
│   └── packwiz-installer-bootstrap.jar
└── logs/                   # Log files
    ├── latest.log
    └── previous.log
```

### Installation Directory
- **Windows**: `C:\Program Files\TheBoysLauncher\` (or custom location)
- **macOS**: `/Applications/TheBoys Launcher.app/`
- **Linux**: `/opt/theboys-launcher/` (or package manager location)

## Installation Methods

### 1. Installer Packages (Recommended)

#### Windows
```bash
# Download and run the installer
TheBoysLauncher-Setup-1.0.0.exe

# Or install via package manager (if available)
winget install TheBoysLauncher
```

**Features:**
- ✅ Professional installation wizard
- ✅ Desktop and Start Menu shortcuts
- ✅ File associations (.modpack files)
- ✅ Add/Remove Programs integration
- ✅ Automatic updates
- ✅ Proper Windows Registry integration

#### macOS
```bash
# Download and mount the DMG
TheBoysLauncher-1.0.0.dmg

# Drag to Applications folder
# Or run the installer package
TheBoysLauncher-1.0.0.pkg
```

**Features:**
- ✅ Professional macOS installation
- ✅ Applications folder integration
- ✅ Launchpad and Dock support
- ✅ Proper code signing
- ✅ Automatic updates
- ✅ Spotlight integration

#### Linux
```bash
# Ubuntu/Debian (DEB package)
sudo dpkg -i theboys-launcher_1.0.0_amd64.deb
sudo apt-get install -f  # Fix dependencies

# Fedora/RHEL (RPM package)
sudo rpm -i theboys-launcher-1.0.0-1.x86_64.rpm

# Universal (AppImage)
chmod +x TheBoysLauncher-1.0.0.AppImage
./TheBoysLauncher-1.0.0.AppImage

# Universal (Tarball)
tar -xzf TheBoysLauncher-1.0.0.tar.gz
cd theboys-launcher-1.0.0
sudo ./install.sh
```

**Features:**
- ✅ Package manager integration
- ✅ Desktop environment integration
- ✅ Menu shortcuts and icons
- ✅ Automatic dependency resolution
- ✅ System-wide installation

### 2. Portable Installation

For users who prefer portable operation:

```bash
# Download the appropriate executable
# Windows: TheBoysLauncher.exe
# macOS: theboys-launcher-macos
# Linux: theboys-launcher-linux

# Place in desired directory
# Run directly - no installation required
```

**Note:** Even in portable mode, user data is stored in `~/.theboys-launcher/` for consistency.

## Migration from Legacy Portable Installation

If you're upgrading from the legacy Winterpack Launcher on Windows:

### Automatic Migration
1. **Install** the new launcher using the installer
2. **Launch** - automatic migration will be detected
3. **Confirm** migration when prompted
4. **Enjoy** - all your instances and settings are preserved

### Manual Migration
If automatic migration doesn't work:

1. **Backup** your current launcher folder
2. **Install** new launcher
3. **Copy** these folders from old to new:
   - `instances/` → `~/.theboys-launcher/instances/`
   - `config/` → `~/.theboys-launcher/config/`
   - `prism/` → `~/.theboys-launcher/prism/`
   - `util/` → `~/.theboys-launcher/util/`

### What Gets Migrated
- ✅ All Minecraft instances
- ✅ Configuration settings
- ✅ Downloaded Java runtimes
- ✅ Prism Launcher installation
- ✅ Modpack configurations
- ✅ Log files

## Custom Installation Directory

### Windows
```bash
# Run installer with custom directory
TheBoysLauncher-Setup-1.0.0.exe /D=D:\Games\TheBoysLauncher
```

### macOS
```bash
# DMG installation allows custom location
# Drag to any folder (e.g., ~/Applications/)

# Package installer (command line)
sudo installer -pkg TheBoysLauncher-1.0.0.pkg -target /
```

### Linux
```bash
# Tarball installation with custom directory
tar -xzf TheBoysLauncher-1.0.0.tar.gz
cd theboys-launcher-1.0.0
sudo ./install.sh /opt/custom-location

# Set custom data directory
export THEBOYS_DATA_DIR=/custom/data/path
./theboys-launcher-linux
```

## Uninstallation

### Windows
1. **Control Panel** → Programs and Features
2. **Select** "TheBoys Launcher"
3. **Click** "Uninstall"
4. **Optional**: Remove `~/.theboys-launcher/` (user data)

### macOS
```bash
# Drag to Trash from Applications folder
# Or run uninstaller if available
sudo rm -rf "/Applications/TheBoys Launcher.app"

# Optional: Remove user data
rm -rf ~/.theboys-launcher
```

### Linux
```bash
# Ubuntu/Debian
sudo apt remove theboys-launcher

# Fedora/RHEL
sudo dnf remove theboys-launcher

# Manual installation
sudo rm -rf /opt/theboys-launcher
sudo rm /usr/local/bin/theboys-launcher
sudo rm /usr/share/applications/theboys-launcher.desktop

# Optional: Remove user data
rm -rf ~/.theboys-launcher
```

## System Requirements

### Minimum Requirements
- **OS**: Windows 10+, macOS 10.15+, or Linux (Ubuntu 18.04+)
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 500MB for launcher, additional for Minecraft instances
- **Network**: Internet connection for modpack downloads

### Recommended Requirements
- **OS**: Windows 11, macOS 12+, or modern Linux distribution
- **RAM**: 8GB or more
- **Storage**: 2GB+ for multiple instances
- **Network**: Broadband connection

### Java Requirements
- **Java 17+**: Automatically downloaded and managed
- **Java 21+**: Supported for newer Minecraft versions
- **Manual Java**: Can be specified in settings

## Troubleshooting

### Common Issues

#### "Cannot create user data directory"
```bash
# Check permissions
ls -la ~/

# Manual creation
mkdir -p ~/.theboys-launcher
chmod 755 ~/.theboys-launcher
```

#### "Permission denied" on Linux
```bash
# Fix executable permissions
chmod +x theboys-launcher-linux

# Install system-wide
sudo ./install.sh
```

#### "Application cannot be opened" on macOS
```bash
# Allow untrusted apps
System Preferences → Security & Privacy → General
# Click "Open Anyway" for TheBoys Launcher

# Or remove quarantine
xattr -d com.apple.quarantine TheBoysLauncher.app
```

#### "Windows Defender blocked"
```bash
# Click "More info" → "Run anyway"
# Or add to Windows Defender exclusions
```

### Migration Issues

#### Migration failed
1. Check the log file in `~/.theboys-launcher/logs/latest.log`
2. Ensure adequate disk space
3. Close any running Minecraft instances
4. Run as administrator on Windows

#### Missing instances after migration
1. Check backup location (shown in migration log)
2. Manually copy instances from backup to `~/.theboys-launcher/instances/`
3. Verify instance configuration files

### Performance Issues

#### Slow startup
1. Check for Java updates in settings
2. Clear log files: `rm ~/.theboys-launcher/logs/*`
3. Verify sufficient RAM allocation
4. Check disk space

#### Download failures
1. Verify internet connection
2. Check firewall settings
3. Try different download server in settings
4. Clear download cache

## Getting Help

### Support Resources
- **Documentation**: [README.md](README.md)
- **Build Guide**: [BUILD_GUIDE.md](BUILD_GUIDE.md)
- **Issues**: [GitHub Issues](https://github.com/dilllxd/theboys-launcher/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dilllxd/theboys-launcher/discussions)

### Reporting Issues
When reporting issues, please include:
- Operating system and version
- Launcher version
- Error messages from log files
- Steps to reproduce the issue
- System specifications

### Community
- Join our Discord community
- Follow development on GitHub
- Share feedback and suggestions

## Installation Verification

After installation, verify everything is working:

1. **Launch** the application
2. **Check** that all sections load correctly
3. **Verify** user data directory exists: `ls ~/.theboys-launcher/`
4. **Test** basic functionality (browse modpacks, check settings)
5. **Confirm** automatic updates are enabled

Congratulations! You now have TheBoys Launcher installed with consistent file management across all platforms.