# TheBoys Launcher - Development & Update Guide

## 🎯 Quick Answers to Your Questions

### **1. Development Commands for Easy Testing**

**YES!** You don't need to build installers every time. Here are the convenient development commands:

```bash
# Quick development and testing (RECOMMENDED)
make run                    # Quick build + run in GUI mode
make run-gui               # Same as above
make run-cli               # Build + run in CLI mode
make run-dev               # Development mode with debug features
make test-run              # Test with sample data

# Live development with hot reload
make dev                   # Wails development server (best for GUI changes)

# Quick builds without running
make quick                 # Fast build for current platform
make build-current         # Clean build for current platform
```

**Development vs Production:**
- **Development**: Use `make run-dev` or `make dev` - builds instantly
- **Testing**: Use `make run` - builds and launches immediately
- **Production**: Use installers - only when releasing

### **2. Auto-Updater Integration with Installers**

**EXCELLENT INTEGRATION!** The auto-updater works seamlessly with both installed and portable versions:

#### **Smart Detection Logic**
```go
// The updater automatically detects how it's installed
isInstalled := platform.IsInstalled()
installPath, _ := platform.GetInstallationPath()

if isInstalled {
    // Handle installed application updates
    // Checks permissions, requires admin on Windows, etc.
} else {
    // Handle portable application updates
    // Simpler process, no special permissions needed
}
```

#### **Update Scenarios Handled**
- ✅ **Installed → Installed**: Updates in-place with proper permissions
- ✅ **Portable → Portable**: Simple executable replacement
- ✅ **Cross-type**: Handles migrations between types
- ✅ **Permission Errors**: Clear error messages and guidance
- ✅ **Network Failures**: Graceful fallbacks and retry logic

#### **Platform-Specific Update Behavior**
- **Windows**: Admin privileges required for system installations
- **macOS**: Checks write permissions, provides guidance
- **Linux**: Handles both system and user installations

### **3. Installer Detection of Existing Installations**

**COMPREHENSIVE DETECTION!** All installers properly detect and handle existing installations:

#### **Windows Installer Detection**
- ✅ **Previous installed version**: Shows current vs new version, offers upgrade
- ✅ **Portable installation**: Detects and offers migration
- ✅ **Existing user data**: Warns and confirms preservation
- ✅ **Parallel installation**: Option to install alongside existing

#### **macOS Installer Detection**
- ✅ **Existing app**: Replaces in Applications folder
- ✅ **User data preservation**: Maintains `~/.theboys-launcher/`
- ✅ **System integration**: Proper package management

#### **Linux Installer Detection**
- ✅ **System installation**: Package manager integration
- ✅ **User installation**: Handles upgrade vs parallel install
- ✅ **Portable detection**: Offers migration from portable versions

### **4. Edge Cases and Error Handling**

**ROBUST ERROR HANDLING!** All edge cases are covered:

#### **Installation Edge Cases**
- ✅ **Application running**: Attempts to close, prompts user
- ✅ **Permission denied**: Clear instructions and solutions
- ✅ **Insufficient disk space**: Pre-installation validation
- ✅ **Network failures**: Retry logic and fallbacks
- ✅ **Corrupted downloads**: Hash verification and re-download
- ✅ **Missing dependencies**: Clear installation instructions

#### **Update Edge Cases**
- ✅ **Permission lost during update**: Detects and reports clearly
- ✅ **Update fails mid-process**: Automatic rollback from backup
- ✅ **New version incompatible**: Version validation and fallback
- ✅ **Network timeout**: Retry with exponential backoff
- ✅ **Disk full**: Cleanup and space reporting

#### **Platform-Specific Edge Cases**
- **Windows**: UAC elevation, registry issues, antivirus interference
- **macOS**: Gatekeeper, SIP restrictions, permissions
- **Linux**: Package manager conflicts, permissions, display issues

## 🚀 Recommended Development Workflow

### **For Everyday Development**
```bash
# 1. Quick development cycle
make run-dev               # Start with debug features

# 2. Test changes
make run-gui               # Test GUI
make run-cli               # Test CLI

# 3. Live GUI development
make dev                   # Hot reload for frontend changes
```

### **For Testing Before Release**
```bash
# 1. Build all platforms
make build-all

# 2. Create installers
make installer-all

# 3. Test installation scenarios
# 4. Test update scenarios
# 5. Test edge cases
```

### **For Release**
```bash
make clean && make all && make installer-all VERSION=v1.0.0
```

## 📋 Complete Feature Matrix

| Feature | Development | Installer | Updater | Edge Cases |
|---------|-------------|-----------|---------|------------|
| **Quick Testing** | ✅ `make run` | ✅ Built-in | ✅ Auto-check | ✅ Handled |
| **GUI Development** | ✅ `make dev` | ✅ N/A | ✅ N/A | ✅ N/A |
| **Cross-Platform Build** | ✅ `make build-all` | ✅ `make installer-all` | ✅ N/A | ✅ Validated |
| **Install Detection** | N/A | ✅ All platforms | ✅ Smart detection | ✅ Covered |
| **Update Integration** | N/A | ✅ Proper registration | ✅ Installer aware | ✅ Handled |
| **Error Handling** | ✅ Development mode | ✅ User-friendly | ✅ Recovery | ✅ Comprehensive |
| **Permission Handling** | N/A | ✅ Elevation prompts | ✅ Validation | ✅ Clear guidance |
| **Migration Support** | N/A | ✅ Auto-migration | ✅ Type detection | ✅ Smooth |
| **Rollback Support** | N/A | ✅ Clean uninstall | ✅ Backup/restore | ✅ Automatic |

## 🎯 Best Practices

### **Development**
1. **Use `make run-dev`** for quick iteration
2. **Test both GUI and CLI** modes regularly
3. **Use `make dev`** for frontend changes
4. **Build installers** only when testing installation scenarios

### **Testing**
1. **Test installer scenarios** on clean systems
2. **Test update process** from both installed and portable
3. **Test edge cases** (permissions, disk space, network)
4. **Test uninstallation** and data preservation

### **Release**
1. **Build all installers** for comprehensive testing
2. **Test on all target platforms**
3. **Validate update mechanism**
4. **Document any breaking changes**

## 🛠️ Troubleshooting Quick Reference

### **Development Issues**
```bash
# Build fails
make clean && make deps && make all

# Frontend not updating
cd frontend && npm run build

# Permission issues
sudo make install  # Linux/macOS if needed
```

### **Installer Issues**
```bash
# Windows: NSIS not found
# Download from https://nsis.sourceforge.io/

# macOS: Tools missing
xcode-select --install

# Linux: GUI installer not working
python3 -m pip install --user PySide6
```

### **Update Issues**
```bash
# Check permissions
ls -la ~/.theboys-launcher/

# Check logs
tail -f ~/.theboys-launcher/logs/latest.log

# Force re-check updates
rm -f ~/.theboys-launcher/config/update-info.json
```

## 🎉 Conclusion

**TheBoys Launcher has excellent development and update infrastructure:**

1. ✅ **Easy Development**: Quick build and run commands for rapid iteration
2. ✅ **Smart Installers**: Detect existing installations and handle migrations
3. ✅ **Robust Updater**: Works seamlessly with both installed and portable versions
4. ✅ **Comprehensive Edge Cases**: All failure scenarios handled gracefully
5. ✅ **Professional Experience**: Users get smooth installation and update experiences

You can confidently develop quickly with `make run` and only build installers when needed for testing or release. The auto-updater and installers work together perfectly to provide a seamless user experience!