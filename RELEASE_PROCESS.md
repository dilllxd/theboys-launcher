# TheBoys Launcher - Release Process Guide

This document describes the complete release process for TheBoys Launcher, including building, packaging, testing, and distribution.

## 📋 Prerequisites

### Development Environment
- Node.js 18+ and npm
- Rust 1.70+ with Cargo
- Tauri CLI: `npm install -g @tauri-apps/cli`

### Platform-Specific Requirements

#### Windows
- Visual Studio Build Tools 2019+ or Visual Studio 2019+
- Windows 10 SDK
- OpenSSL for Windows (if building from source)
- Wix Toolset (for MSI installer)

#### macOS
- Xcode Command Line Tools
- macOS 11+ (for building and testing)
- Apple Developer ID (for code signing, optional for development)

#### Linux
- Basic build tools: `build-essential`
- WebKitGTK development libraries: `libwebkit2gtk-4.1-dev`
- OpenSSL development libraries: `libssl-dev`

### Cross-Platform Build Support
For building packages for other platforms, you may need:
- Docker (for Linux builds on Windows/macOS)
- Cross-compilation toolchains

## 🏗️ Build Process

### 1. Development Build
```bash
# Start development server
npm run tauri:dev
```

### 2. Production Build (Single Platform)
```bash
# Build for current platform
npm run tauri:build

# Build for specific target (example: Windows x64)
npm run tauri build -- --target x86_64-pc-windows-msvc
```

### 3. Full Release Build (All Platforms)
```bash
# Automated build script for all platforms
npm run build:release
```

The build script will:
- Install dependencies
- Build the frontend
- Create packages for all supported platforms
- Generate release notes
- Create checksums for all packages

## 📦 Package Types

### Windows
- **MSI Installer** (`*.msi`): Professional installer with proper shortcuts and file associations
- **NSIS Installer** (`*.exe`): Alternative installer with customizable dialogs

### macOS
- **DMG Package** (`*.dmg`): Drag-and-drop installation with background image
- **App Bundle** (`*.app`): macOS application bundle

### Linux
- **AppImage** (`*.AppImage`): Portable, self-contained application
- **DEB Package** (`*.deb`): Debian/Ubuntu package with proper dependencies
- **RPM Package** (`*.rpm`): Fedora/RHEL/CentOS package

## 🔧 Configuration

### Build Configuration
All build settings are configured in `tauri.conf.json`:

- **Bundle settings**: Icons, file associations, dependencies
- **Platform-specific options**: Installers, signing, certificates
- **Updater configuration**: Auto-update endpoints and behavior

### Icon Requirements
Icons should be placed in the `icons/` directory:
- `32x32.png` - Small icon for Linux
- `128x128.png` - Medium icon
- `128x128@2x.png` - High DPI icon
- `icon.ico` - Windows icon
- `icon.icns` - macOS icon
- `256x256.png` - Linux desktop integration
- `512x512.png` - High resolution Linux icon

### Additional Assets
- `icons/banner.png` (493x58px) - Windows MSI installer banner
- `icons/dialog.png` (493x312px) - Windows MSI installer dialog image
- `icons/dmg-background.png` - macOS DMG background image

## 🧪 Testing Process

### Pre-Release Testing
1. **Unit Tests**: Run all backend and frontend tests
   ```bash
   npm test
   npm run test:coverage
   ```

2. **Integration Tests**: Test core functionality
   - Modpack installation and launching
   - Settings management
   - Update functionality

3. **Platform Testing**:
   - **Windows**: Test on Windows 10/11 (x64)
   - **macOS**: Test on macOS 11+ (Intel and Apple Silicon)
   - **Linux**: Test on Ubuntu/Debian and Fedora

### Installation Testing
1. **Clean Install**: Test installation on a fresh system
2. **Upgrade**: Test upgrading from previous version
3. **Uninstall**: Test complete removal and cleanup

### Functionality Testing
1. **Core Features**:
   - Modpack browsing and installation
   - Minecraft launching
   - Settings management
   - Update checking and installation

2. **Platform Integration**:
   - File associations (.tbmod, .tbprofile)
   - Desktop shortcuts
   - Auto-start functionality
   - System tray integration

## 🚀 Release Process

### 1. Preparation
```bash
# Update version numbers
# - package.json (version field)
# - tauri.conf.json (version field)
# - Cargo.toml (version field)

# Update changelog
# Add release notes to RELEASE_NOTES.md

# Run full test suite
npm test
npm run lint
```

### 2. Build Release
```bash
# Build all packages
npm run build:release

# Verify packages in releases/ directory
ls -la releases/
```

### 3. Package Verification
1. **Checksum Verification**: Verify generated checksums
   ```bash
   sha256sum -c releases/checksums.txt
   ```

2. **Package Integrity**: Test each package:
   - Install on clean system
   - Verify functionality
   - Check file associations
   - Test auto-updater

### 4. Distribution Setup

#### GitHub Releases
1. Create new release on GitHub
2. Upload all packages
3. Include release notes
4. Attach checksums.txt file

#### Auto-Update Server
1. Upload packages to update server
2. Configure update endpoints in `tauri.conf.json`
3. Ensure proper HTTPS and certificate setup

#### Package Repositories
- **Windows**: Consider Microsoft Store submission
- **macOS**: App Store submission (requires proper signing)
- **Linux**: Submit to distribution repositories

### 5. Post-Release
1. **Documentation Update**: Update website and documentation
2. **Announcement**: Announce release on relevant channels
3. **Monitoring**: Monitor download stats and user feedback
4. **Bug Tracking**: Track and prioritize bug reports

## 🔐 Code Signing (Optional but Recommended)

### Windows Code Signing
1. Obtain code signing certificate from CA
2. Configure certificate in `tauri.conf.json`
3. Sign packages during build process

### macOS Code Signing
1. Enroll in Apple Developer Program
2. Obtain Developer ID certificate
3. Configure signing in `tauri.conf.json`
4. Notarize packages for Gatekeeper compliance

### Linux Signing
1. Create GPG key for package signing
2. Sign DEB/RPM packages
3. Configure package repositories to trust signature

## 🔄 Auto-Update Configuration

### Update Server Setup
1. Configure update endpoints in `tauri.conf.json`
2. Set up HTTPS server for update packages
3. Implement version checking API
4. Configure rollback capability

### Update Process
1. **Background Check**: Periodically check for updates
2. **User Notification**: Inform user of available updates
3. **Download**: Download update in background
4. **Installation**: Install update and restart application
5. **Rollback**: Fallback to previous version if needed

## 🐛 Troubleshooting

### Common Build Issues
1. **Missing Dependencies**: Install platform-specific dependencies
2. **Permission Errors**: Run with appropriate permissions
3. **Network Issues**: Check internet connectivity and firewalls
4. **Disk Space**: Ensure sufficient disk space for builds

### Platform-Specific Issues
- **Windows**: Verify Visual Studio and Windows SDK installation
- **macOS**: Check Xcode command line tools installation
- **Linux**: Install missing -dev packages for dependencies

### Update Issues
1. **Server Configuration**: Verify update server is accessible
2. **Certificate Issues**: Check HTTPS certificate validity
3. **Version Conflicts**: Ensure proper version numbering
4. **Network Problems**: Test update server connectivity

## 📊 Release Checklist

### Pre-Release
- [ ] Version numbers updated in all configuration files
- [ ] All tests passing
- [ ] Code linting complete
- [ ] Documentation updated
- [ ] Release notes written

### Build Process
- [ ] Clean build environment
- [ ] All platform packages built successfully
- [ ] Checksums generated and verified
- [ ] Packages tested on clean systems

### Distribution
- [ ] GitHub release created
- [ ] All packages uploaded
- [ ] Update server configured
- [ ] Documentation published
- [ ] Community notified

### Post-Release
- [ ] Monitor download statistics
- [ ] Track user feedback
- [ ] Address critical issues promptly
- [ ] Plan next release cycle

## 📞 Support

For issues with the release process:
1. Check this documentation first
2. Review Tauri documentation: https://tauri.app/
3. Check GitHub issues: https://github.com/theboys/launcher/issues
4. Contact the development team

---

**Note**: This release process should be followed for all stable releases. For pre-releases and testing builds, some steps may be simplified or skipped.