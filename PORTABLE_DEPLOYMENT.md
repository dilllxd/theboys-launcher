# Portable Deployment Guide

## Overview

TheBoys Launcher maintains the same **portable deployment model** as the legacy Winterpack Launcher. Each executable is completely self-contained and requires **no installation**.

## How It Works

### Single-File Architecture
- **Frontend Embedded**: All web assets are embedded in the executable
- **No External Dependencies**: Everything needed is included
- **Self-Updating**: Built-in update mechanism
- **Portable Operation**: Works from any directory

### File Creation Behavior

#### Windows - Fully Portable 🪟
```
LauncherDirectory/
├── TheBoysLauncher.exe          # Main executable (you provide this)
├── prism/                       # Created automatically
│   └── [Prism Launcher files]
├── util/                        # Created automatically
│   ├── jre17/                   # Java runtime
│   ├── packwiz-installer-bootstrap.exe
│   └── packwiz-installer-bootstrap.jar
├── instances/                   # Created automatically
│   └── [Minecraft instances]
├── config/                      # Created automatically
│   ├── settings.json
│   └── modpacks.json
└── logs/                        # Created automatically
    ├── latest.log
    └── previous.log
```

**Windows is fully portable** - everything is created beside the executable. You can:
- Move the entire folder anywhere
- Run from USB drives
- Copy between computers
- No registry entries or system files

#### macOS - User Home Based 🍎
```
~/.theboys-launcher/              # Created in user home directory
├── prism/                       # Prism Launcher
├── util/                        # Utilities (Java, packwiz)
├── instances/                   # Minecraft instances
├── config/                      # Configuration files
└── logs/                        # Log files
```

**macOS creates files in user home** due to macOS security requirements:
- Executable can be placed anywhere
- All data stored in `~/.theboys-launcher/`
- Follows macOS application conventions
- Sandbox-compatible

#### Linux - User Home Based 🐧
```
~/.theboys-launcher/              # Created in user home directory
├── prism/                       # Prism Launcher
├── util/                        # Utilities (Java, packwiz)
├── instances/                   # Minecraft instances
├── config/                      # Configuration files
└── logs/                        # Log files
```

**Linux creates files in user home** following Linux conventions:
- Executable can be placed anywhere
- All data stored in `~/.theboys-launcher/`
- Follows XDG Base Directory specification
- Proper Unix permissions

## Download and Run

### Windows
1. Download `TheBoysLauncher.exe`
2. Place in any writable folder
3. Double-click to run
4. **That's it!** Everything else is automatic

### macOS
1. Download `theboys-launcher-macos-amd64` (Intel) or `theboys-launcher-macos-arm64` (Apple Silicon)
2. Open Terminal, navigate to download location
3. Make executable: `chmod +x theboys-launcher-macos-*`
4. Run: `./theboys-launcher-macos-*`
5. **That's it!** Everything else is automatic

### Linux
1. Download `theboys-launcher-linux-amd64` (x64) or `theboys-launcher-linux-arm64` (ARM64)
2. Open Terminal, navigate to download location
3. Make executable: `chmod +x theboys-launcher-linux-*`
4. Run: `./theboys-launcher-linux-*`
5. **That's it!** Everything else is automatic

## Migration from Legacy Launcher

### From Winterpack Launcher (Windows)
1. **Backup** your current launcher folder (optional but recommended)
2. **Download** the new `TheBoysLauncher.exe`
3. **Replace** the old executable with the new one
4. **Run** - all your instances, settings, and configurations are preserved
5. **Delete** old `WinterpackLauncher.exe` if desired

### What Gets Preserved
- ✅ All Minecraft instances
- ✅ Configuration settings
- ✅ Downloaded Java runtimes
- ✅ Prism Launcher installation
- ✅ Modpack configurations
- ✅ Log files (moved to new location)

### What Changes
- 🔄 New GUI interface (web-based instead of TUI)
- 🔄 Enhanced features and better error handling
- 🔄 Cross-platform support
- 🔄 Better performance and reliability

## CLI Support (Legacy Compatible)

All legacy command-line arguments are preserved:

```bash
# List available modpacks
TheBoysLauncher.exe --list-modpacks

# Run in CLI mode (no GUI)
TheBoysLauncher.exe --cli

# Install specific modpack
TheBoysLauncher.exe --cli --modpack 123

# Show settings
TheBoysLauncher.exe --cli --settings

# Show help
TheBoysLauncher.exe --help
```

## Benefits of Portable Deployment

### Advantages
- **No Installation**: No admin rights required
- **No Registry**: Doesn't modify system settings
- **Portable**: Run from USB, network drive, any location
- **Isolated**: Doesn't interfere with other applications
- **Easy Backup**: Just copy the folder
- **Clean Uninstall**: Just delete the folder
- **Multiple Versions**: Run different versions side-by-side

### Security
- **Sandboxed**: Limited system interaction
- **No Hidden Files**: Everything is visible
- **User-Level**: No system-wide changes
- **Transparent**: You can see all created files

## Deployment Scenarios

### Personal Gaming PC
- Download to `C:\Games\TheBoys\`
- All instances in `C:\Games\TheBoys\instances\`
- Move entire folder when needed

### USB Drive Gaming
- Download to USB drive
- Play on any computer
- Settings and instances travel with you

### Network Share
- Place on network share
- Multiple users can run same copy
- Each user gets their own instances

### School/Work Computer
- No admin rights needed
- Runs from user directory
- Doesn't install system-wide

## Troubleshooting

### Permission Issues
```bash
# Linux/macOS: Make executable
chmod +x theboys-launcher-*

# Windows: Run as user (not admin required)
# If blocked by Windows Defender, click "More info" -> "Run anyway"
```

### Portable vs. User Directory Confusion
- **Windows**: Everything beside the executable
- **macOS/Linux**: Everything in `~/.theboys-launcher/`
- **Why**: macOS/Linux security policies require user directory storage

### Multiple Launchers
- You can run multiple launcher versions side-by-side
- Each version will use the same data directory
- Windows: Different folders = completely separate
- macOS/Linux: All versions share `~/.theboys-launcher/`

## Comparison with Legacy

| Feature | Legacy Winterpack | New TheBoys | Status |
|---------|-------------------|-------------|---------|
| Portable Deployment | ✅ Windows only | ✅ All platforms | Enhanced |
| Single Executable | ✅ | ✅ | Maintained |
| No Installation | ✅ | ✅ | Maintained |
| CLI Support | ✅ | ✅ | Maintained |
| Self-Updating | ✅ | ✅ | Enhanced |
| Cross-Platform | ❌ Windows only | ✅ All platforms | New |
| GUI | TUI (console) | Modern Web GUI | Enhanced |
| Error Handling | Basic | Comprehensive | Enhanced |
| Performance | Standard | Optimized | Enhanced |

## Conclusion

The new TheBoys Launcher **maintains the simple, portable deployment** that made the legacy launcher popular while adding significant enhancements:

- ✅ **Same portable model** - No installation required
- ✅ **Single executable** - Everything embedded
- ✅ **Cross-platform** - Windows, macOS, Linux
- ✅ **Enhanced features** - Better GUI, error handling, performance
- ✅ **Legacy compatible** - All CLI arguments preserved
- ✅ **Easy migration** - Seamless upgrade from legacy

The result is a **modern, cross-platform launcher** that maintains the simplicity and portability of the original while providing a significantly improved user experience.