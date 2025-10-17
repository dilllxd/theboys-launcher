# TheBoys Launcher - GUI Installer Guide

## Overview

TheBoys Launcher provides **professional GUI installers** for all platforms - no command line or scripts required! Each installer features a user-friendly wizard that guides users through the installation process with clear instructions and visual feedback.

## 🎯 Key Features

### **User-Friendly Interface**
- ✅ **Professional installation wizards** on all platforms
- ✅ **Directory selection** with clear explanations
- ✅ **Visual progress indicators** with detailed feedback
- ✅ **Component selection** (shortcuts, file associations, etc.)
- ✅ **One-click installation** - no technical knowledge required

### **Professional Experience**
- ✅ **Welcome screens** with feature highlights
- ✅ **License agreements** with clear acceptance
- ✅ **Customizable installation options**
- ✅ **Completion screens** with launch options
- ✅ **Easy uninstallation** with complete cleanup

## 📦 Platform-Specific Installers

### Windows Installer (Professional Wizard)

**File:** `TheBoysLauncher-Setup-1.0.0.exe`

**Features:**
- 🎨 **Modern installation wizard** with step-by-step guidance
- 📁 **Custom directory selection** - install anywhere you want
- 🎯 **Component selection**:
  - Desktop shortcut creation
  - Start Menu shortcuts
  - File association for .modpack files
- 🔄 **Automatic migration** from legacy portable installations
- 🏠 **User data separation** - app files separate from game data
- 🗑️ **Complete uninstallation** with Add/Remove Programs integration

**Installation Process:**
1. **Welcome** - Introduction to TheBoys Launcher
2. **License** - MIT License agreement
3. **Components** - Choose shortcuts and associations
4. **Directory** - Select installation location
5. **Install** - Progress with detailed feedback
6. **Finish** - Launch option and completion summary

### macOS Installer (Native Package)

**File:** `TheBoys Launcher-1.0.0.pkg`

**Features:**
- 🍎 **Native macOS installation** with familiar interface
- 📱 **Drag-and-drop installation** to Applications folder
- 📋 **License agreement** with clear terms
- ✅ **System integration** with Launchpad and Dock
- 🔐 **Code signing** for security and trust
- 🗑️ **Proper uninstallation** with cleanup script

**Installation Process:**
1. **Introduction** - Welcome and feature overview
2. **License** - License agreement acceptance
3. **Installation Type** - Standard installation
4. **Install** - Copy files with progress indication
5. **Summary** - Completion with launch option

### Linux Installers (Multiple GUI Options)

**Primary Installer:** `TheBoysLauncher-Linux-Installer-1.0.0.tar.gz`

**Features:**
- 🐧 **Qt-based GUI installer** with modern interface
- 📁 **Directory selection** with default suggestions
- ⚙️ **Installation options**:
  - Command-line symlink creation
  - Desktop shortcut creation
- 📊 **Real-time progress** with detailed logging
- 🚀 **Launch option** after installation
- 🔄 **Multiple fallback options** for different environments

**Alternative Installers:**
- **Zenity Installer** - Lightweight GTK-based installer
- **Dialog Installer** - Terminal-based GUI for minimal environments
- **AppImage** - Self-contained portable installer

## 🔄 Installation Process (All Platforms)

### Step 1: Download
1. Visit the [GitHub Releases](https://github.com/dilllxd/theboys-launcher/releases) page
2. Download the appropriate installer for your platform
3. Verify the download completed successfully

### Step 2: Launch Installer
- **Windows**: Double-click `TheBoysLauncher-Setup-1.0.0.exe`
- **macOS**: Double-click `TheBoys Launcher-1.0.0.pkg`
- **Linux**: Extract and run the installer from the GUI

### Step 3: Follow Wizard
1. **Welcome** - Read the introduction
2. **License** - Accept the license agreement
3. **Options** - Choose installation preferences
4. **Directory** - Select where to install (optional)
5. **Install** - Wait for installation to complete
6. **Finish** - Launch the application

### Step 4: Launch Application
- **Windows**: Start Menu → TheBoys Launcher
- **macOS**: Applications folder → TheBoys Launcher
- **Linux**: Applications menu or run `theboys-launcher`

## 🎨 User Interface Screenshots

### Windows Installer
```
┌─────────────────────────────────────────────────────────────┐
│ Welcome to TheBoys Launcher Setup                        │
│─────────────────────────────────────────────────────────────│
│ This wizard will guide you through the installation of     │
│ TheBoys Launcher, a modern Minecraft modpack launcher.     │
│                                                         │
│ Features:                                                │
│ • Modern graphical user interface                         │
│ • Automatic Java runtime management                       │
│ • Support for multiple modpack sources                    │
│ • Automatic updates and backups                           │
│                                                         │
│ Click Next to continue, or Cancel to exit Setup.         │
│                                                         │
│                  [  Back  ] [ Next >] [ Cancel ]           │
└─────────────────────────────────────────────────────────────┘
```

### Directory Selection
```
┌─────────────────────────────────────────────────────────────┐
│ Select Installation Folder                                 │
│─────────────────────────────────────────────────────────────│
│ Please select the folder where you would like to install    │
│ TheBoys Launcher.                                          │
│                                                         │
│ This will install the application files, but your saved    │
│ games and settings will be stored in your user profile    │
│ directory.                                                │
│                                                         │
│ Destination Folder:                                       │
│ [ C:\Program Files\TheBoys Launcher         ] [ Browse... ] │
│                                                         │
│ Space required: 45.2 MB                                  │
│ Space available: 156.8 GB                                 │
│                                                         │
│                  [  Back  ] [ Next >] [ Cancel ]           │
└─────────────────────────────────────────────────────────────┘
```

### Progress Screen
```
┌─────────────────────────────────────────────────────────────┐
│ Installing TheBoys Launcher                               │
│─────────────────────────────────────────────────────────────│
│ Please wait while TheBoys Launcher is being installed...   │
│                                                         │
│ Progress: ████████████████████████████████████░░ 85%       │
│                                                         │
│ Status: Creating desktop shortcuts...                     │
│                                                         │
│ Details:                                                 │
│ • Creating directories                                    │
│ ✓ Installing application files                            │
│ • Creating desktop entry                                   │
│ • Setting up user data directory                          │
│ • Updating desktop database                               │
│                                                         │
│                    [ Cancel ]                            │
└─────────────────────────────────────────────────────────────┘
```

## 🏁 Installation Completion

### Success Screen
All platforms show a completion screen with:
- ✅ **Success message** confirming installation
- 🚀 **Launch option** to start the application immediately
- 📂 **User data location** information
- 🔗 **Support links** for help and documentation
- 💡 **Next steps** guidance

### User Data Information
The installer clearly explains that:
- **Application files** are installed in the chosen directory
- **User data** (instances, settings, saves) is stored in:
  - `~/.theboys-launcher/` (Linux/macOS)
  - `%USERPROFILE%\.theboys-launcher\` (Windows)

## 🛠️ Advanced Options

### Custom Installation Directory
Users can choose any installation location:
- **Windows**: Any writable directory (default: `C:\Program Files\TheBoys Launcher\`)
- **macOS**: Applications folder or custom location
- **Linux**: `/opt/theboys-launcher` or user-defined path

### Component Selection
Choose which features to install:
- **Desktop shortcuts** for quick access
- **Start Menu/Applications menu** integration
- **File associations** for .modpack files
- **Command-line tools** for advanced users

### Migration Support
For Windows users upgrading from the legacy portable version:
- **Automatic detection** of portable installations
- **One-click migration** with backup creation
- **Data preservation** - all instances and settings maintained
- **Fallback option** - manual migration if needed

## 🔧 Troubleshooting

### Common Issues

#### "Installer won't run"
- **Windows**: Right-click → "Run as administrator"
- **macOS**: Allow in System Preferences → Security & Privacy
- **Linux**: Make executable: `chmod +x installer`

#### "Not enough disk space"
- Clear temporary files
- Choose a different installation directory
- Free up space on the target drive

#### "Installation failed"
- Check available disk space
- Ensure write permissions to installation directory
- Temporarily disable antivirus software
- Run installer as administrator (Windows)

#### "Can't launch after installation"
- Check if application is in quarantine (macOS)
- Verify executable permissions (Linux)
- Look for error messages in log files
- Reinstall if necessary

### Getting Help
- **Documentation**: [INSTALLATION_GUIDE.md](INSTALLATION_GUIDE.md)
- **Support**: [GitHub Issues](https://github.com/dilllxd/theboys-launcher/issues)
- **Community**: [GitHub Discussions](https://github.com/dilllxd/theboys-launcher/discussions)

## 🎉 Conclusion

TheBoys Launcher's GUI installers provide a **professional, user-friendly installation experience** that requires no technical knowledge. With clear visual feedback, customizable options, and automatic setup, users can get started with Minecraft modpacks in just a few clicks!

The installers prioritize:
- **Simplicity** - No command line or scripts required
- **Clarity** - Step-by-step guidance with clear explanations
- **Flexibility** - Customizable installation options
- **Reliability** - Error handling and recovery options
- **Professionalism** - Modern, polished interface

This approach ensures that even non-technical users can easily install and enjoy TheBoys Launcher on any platform!