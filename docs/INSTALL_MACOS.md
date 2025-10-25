# macOS Installation Guide

This guide provides detailed instructions for installing TheBoysLauncher on macOS, including both Intel and Apple Silicon Macs.

## 🔧 System Requirements

### Minimum Requirements
- **macOS 10.15 (Catalina) or newer**
- **4 GB RAM** (8 GB recommended)
- **2 GB free disk space**
- **Active internet connection** for initial setup

### Supported Architecture
- ✅ **Intel Macs** (x86_64)
- ✅ **Apple Silicon Macs** (M1, M2, M3, etc.)
- ✅ **Universal Binary** (works on both)

## 📥 Installation Methods

### Method 1: DMG Installer (Recommended)

1. **Download the DMG**
   - Go to [Releases](https://github.com/dilllxd/theboyslauncher/releases)
   - Download `TheBoysLauncher-Universal.dmg`

2. **Install the Application**
   - Double-click the downloaded DMG file
      - A window will open showing TheBoysLauncher icon
      - Drag TheBoysLauncher to your Applications folder

3. **First Launch**
   - Open Applications folder
      - Double-click TheBoysLauncher
   - If you see a security warning, see "Security Instructions" below

### Method 2: Manual App Bundle

1. **Download the App Bundle**
   - Go to [Releases](https://github.com/dilllxd/theboyslauncher/releases)
   - Download `TheBoysLauncher-Universal.zip`

2. **Extract and Install**
   - Double-click the ZIP file to extract it
   - Drag `TheBoysLauncher.app` to your Applications folder

3. **Set Permissions**
   - Open Terminal
   - Run: `chmod +x "/Applications/TheBoysLauncher.app/Contents/MacOS/TheBoysLauncher"`

## 🔒 Security Instructions

macOS may show security warnings because the app is not signed with an Apple Developer certificate. This is normal for open-source software.

### If you see "TheBoysLauncher.app can't be opened because Apple cannot check it for malicious software":

1. **Right-click Method**
   - Right-click (or Control-click) on TheBoysLauncher.app
   - Select "Open" from the context menu
   - Click "Open" in the confirmation dialog

2. **System Preferences Method**
   - Open System Preferences → Security & Privacy
   - Click the "General" tab
   - Look for "TheBoysLauncher.app was blocked from use"
   - Click "Allow Anyway"

3. **Terminal Method (Advanced)**
   ```bash
   # Remove quarantine attribute
   xattr -cr "/Applications/TheBoysLauncher.app"
   ```

### If you see "TheBoysLauncher.app is damaged and can't be opened":

```bash
# Remove quarantine attribute completely
xattr -d com.apple.quarantine "/Applications/TheBoysLauncher.app"

# Or use the comprehensive removal command
xattr -cr "/Applications/TheBoysLauncher.app"
```

## 🚀 First Launch Setup

1. **Launch the Application**
      - Open TheBoysLauncher from Applications
   - The launcher will detect your macOS version and architecture

2. **Initial Configuration**
   - The launcher will create necessary directories in:
     `~/Library/Application Support/TheBoysLauncher`
   - It will automatically detect your system memory
   - Java runtime will be downloaded automatically if needed

3. **Modpack Selection**
   - Browse available modpacks
   - Click "Download" to install a modpack
   - The launcher will handle Java and Prism Launcher setup

## 🛠️ Troubleshooting

### Common Issues

#### "App won't open" or "bounces in Dock then closes"
```bash
# Check console logs
log show --predicate 'process == "TheBoysLauncher"' --last 1m

# Remove and reinstall
rm -rf "/Applications/TheBoysLauncher.app"
# Then reinstall using the DMG method
```

#### "Out of disk space" error
```bash
# Clear launcher cache
rm -rf ~/Library/Application\ Support/TheBoysLauncher/cache

# Clear Java cache (if exists)
rm -rf ~/Library/Application\ Support/TheBoysLauncher/java
```

#### "Java not found" error
```bash
# Let launcher download Java automatically
# Or specify custom Java path in launcher settings
```

#### "Network connection failed"
```bash
# Check firewall settings
# Ensure the launcher can access github.com and adoptium.net
```

### Reset to Factory Settings

If you need to completely reset the launcher:

```bash
# Remove all launcher data
rm -rf ~/Library/Application\ Support/TheBoysLauncher

# Remove app and reinstall
rm -rf "/Applications/TheBoysLauncher.app"
```

## 🧪 Verification

### Check the Application
```bash
# Verify app bundle structure
ls -la "/Applications/TheBoysLauncher.app/Contents/"

# Check executable permissions
ls -la "/Applications/TheBoysLauncher.app/Contents/MacOS/TheBoysLauncher"

# Verify universal binary (if using universal version)
file "/Applications/TheBoysLauncher.app/Contents/MacOS/TheBoysLauncher"
lipo -info "/Applications/TheBoysLauncher.app/Contents/MacOS/TheBoysLauncher"
```

### Test Functionality
1. Launch the application
2. Verify it detects your macOS version correctly
3. Check that modpacks load properly
4. Ensure Prism Launcher integration works

## 📁 File Locations

### Application Files
- **App Bundle**: `/Applications/TheBoysLauncher.app/`
- **Executable**: `/Applications/TheBoysLauncher.app/Contents/MacOS/TheBoysLauncher`

### Data Files
- **Configuration**: `~/Library/Application Support/TheBoysLauncher/`
- **Logs**: `~/Library/Application Support/TheBoysLauncher/logs/`
- **Cache**: `~/Library/Application Support/TheBoysLauncher/cache/`
- **Java**: `~/Library/Application Support/TheBoysLauncher/java/`
- **Prism Launcher**: `~/Library/Application Support/TheBoysLauncher/prism/`

### Modpack Files
- **Downloads**: `~/Library/Application Support/TheBoysLauncher/modpacks/`
- **Instances**: `~/Library/Application Support/TheBoysLauncher/instances/`

## 🔧 Advanced Configuration

### Environment Variables
You can customize behavior with environment variables:

```bash
# Set custom data directory
export THEBOYS_LAUNCHER_HOME="/path/to/custom/directory"

# Disable automatic updates
export THEBOYS_LAUNCHER_NO_UPDATE="1"

# Enable debug logging
export THEBOYS_LAUNCHER_DEBUG="1"
```

### Manual Java Configuration
If you want to use a specific Java installation:

1. Go to launcher settings
2. Navigate to "Java" section
3. Set "Java Path" to your Java installation
4. Typical Java paths:
   - `/Library/Java/JavaVirtualMachines/`
   - `/usr/local/opt/openjdk/`
   - `~/.jenv/versions/`

## 📞 Support

If you encounter issues not covered in this guide:

1. **Check Logs**: Look in `~/Library/Application Support/TheBoysLauncher/logs/`
2. **Search Issues**: [GitHub Issues](https://github.com/dilllxd/theboyslauncher/issues)
3. **Create New Issue**: Include your macOS version, Mac model, and error logs
4. **Community**: [GitHub Discussions](https://github.com/dilllxd/theboyslauncher/discussions)

## 🔄 Updates

The launcher includes automatic update functionality:

- **Automatic**: Checks for updates on launch
- **Manual**: Use "Check for Updates" in settings
- **Update Files**: Downloaded and applied automatically
- **Rollback**: Previous versions kept in launcher directory

---

**Enjoy using TheBoysLauncher on macOS!** 🎉

For more information, see the main [README.md](./README.md) file.