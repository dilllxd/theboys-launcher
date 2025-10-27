# TheBoysLauncher

A cross-platform Minecraft bootstrapper and modpack manager that automatically downloads, configures, and launches Prism Launcher with modpacks.

![Platform Support](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue)
![Go Version](https://img.shields.io/badge/Go-1.22+-blue)
![License](https://img.shields.io/badge/license-MIT-green)

## 🚀 Features

- **Cross-Platform Support**: Windows, macOS (Intel & Apple Silicon), and Linux
- **Automatic Java Management**: Downloads and manages the correct Java runtime
- **Prism Launcher Integration**: Automatically fetches and configures Prism Launcher
- **Modpack Management**: Download and launch modpacks with one click
- **Self-Updating**: Built-in update mechanism for seamless updates
- **GUI Interface**: Clean and intuitive user interface built with Fyne

## 📦 Installation

### Windows
1. Download the latest `TheBoysLauncher-Setup.exe` from [Releases](https://github.com/dilllxd/theboyslauncher/releases)
2. Run the installer
3. Launch from desktop shortcut or Start menu

### macOS
1. Download the latest `TheBoysLauncher-Universal.zip` from [Releases](https://github.com/dilllxd/theboyslauncher/releases)
2. Extract the ZIP file
3. Drag `TheBoysLauncher.app` to your Applications folder
4. Launch from Applications (first launch may require bypassing Gatekeeper)

**Note for macOS users**: The launcher is not signed, so you may need to right-click → Open or go to System Preferences → Security & Privacy to allow it to run.

### Linux
1. Download the latest `TheBoysLauncher-linux` from [Releases](https://github.com/dilllxd/theboyslauncher/releases)
2. Make the binary executable: `chmod +x TheBoysLauncher-linux`
3. Run `./TheBoysLauncher-linux` from the downloaded directory

## 🛠️ Building from Source

### Prerequisites
- **Go 1.22+** - [Install Go](https://golang.org/dl/)
- **For macOS**: Xcode Command Line Tools (`xcode-select --install`)
- **For Windows**: PowerShell 5.1+ (included with Windows 10+)

### Quick Build
```bash
# Clone the repository
git clone https://github.com/dilllxd/theboyslauncher.git
cd theboyslauncher

# Build for current platform
make build

# Or build using go directly
go build -o TheBoysLauncher .
```

### Cross-Platform Builds

#### Using Makefile (Recommended)
```bash
# Build for all platforms
make build-all

# Build specific platforms
make build-windows          # Windows executable
make build-linux            # Linux executable
make build-macos-intel      # macOS Intel
make build-macos-arm64      # macOS Apple Silicon
make build-macos-universal  # macOS Universal binary

# Create packages
make package-macos-universal  # macOS app bundle
make dmg-macos-universal      # macOS DMG installer

# Full build and package for all platforms
make package-all
```

#### Manual Cross-Compilation
```bash
# Windows
export GOOS=windows GOARCH=amd64 CGO_ENABLED=0
go build -ldflags="-s -w -H=windowsgui" -o TheBoysLauncher.exe .

# Linux
export GOOS=linux GOARCH=amd64 CGO_ENABLED=0
go build -ldflags="-s -w" -o TheBoysLauncher-linux .

# macOS Intel
export GOOS=darwin GOARCH=amd64 CGO_ENABLED=1
go build -ldflags="-s -w" -o TheBoysLauncher .

# macOS Apple Silicon
export GOOS=darwin GOARCH=arm64 CGO_ENABLED=1
go build -ldflags="-s -w" -o TheBoysLauncher-arm64 .

# Create universal binary (on macOS)
lipo -create TheBoysLauncher TheBoysLauncher-arm64 -output TheBoysLauncher-universal
```

## 📁 Project Structure

```
theboyslauncher/
├── main.go                    # Application entry point
├── platform.go                # Platform detection and abstractions
├── platform_windows.go        # Windows-specific implementations
├── platform_darwin.go         # macOS-specific implementations
├── memory_windows.go          # Windows memory detection
├── memory_darwin.go           # macOS memory detection
├── process_windows.go         # Windows process management
├── process_darwin.go          # macOS process management
├── java.go                    # Java runtime management
├── prism.go                   # Prism Launcher integration
├── download.go                # Download and extraction utilities
├── launcher.go                # Main launcher logic
├── config.go                  # Configuration management
├── utils.go                   # Utility functions
├── update_windows.go          # Windows update handling
├── update_darwin.go           # macOS update handling
├── gui.go                     # GUI interface
├── console_windows.go         # Windows console handling
├── console_darwin.go          # macOS console handling
├── multimc_windows.go         # Windows MultiMC handling
├── multimc_darwin.go          # macOS MultiMC handling
├── packwiz.go                 # Packwiz integration
├── Makefile                   # Build system
├── go.mod                     # Go modules
├── version.env                # Centralized version configuration
├── icon.ico                   # Application icon
├── LICENSE.txt                # License file
├── README.md                  # This file
├── scripts/                   # Build and utility scripts
│   ├── get-version.sh         # Version extraction (Unix)
│   ├── get-version.ps1        # Version extraction (Windows)
│   ├── set-version.sh         # Version updating (Unix)
│   ├── set-version.ps1        # Version updating (Windows)
│   ├── validate-version.sh    # Version validation
│   ├── update-wix-version.ps1 # WiX version sync
│   ├── create-app-bundle.sh   # macOS app bundle creation
│   ├── convert-icon.sh        # Icon conversion for macOS
│   └── create-dmg.sh          # macOS DMG creation
├── tools/                     # Development and build tools
│   ├── build.bat              # Windows build script
│   ├── build.ps1              # PowerShell build script
│   ├── verify-build.bat       # Windows build verification
│   ├── verify-build.sh        # Unix build verification
│   └── update-version.ps1     # Legacy version script
├── config/                    # Configuration files
│   ├── modpacks.json          # Modpack configurations
│   └── openssl.cnf            # OpenSSL configuration
├── wix/                       # WiX installer configuration
│   ├── TheBoysLauncher.wxs    # WiX installer script
│   └── Product.wxs            # Product configuration
├── docs/                      # Documentation
│   ├── BUILD.md               # Build instructions
│   ├── INSTALL_MACOS.md       # macOS installation guide
│   ├── MACOS_DEVELOPMENT_PLAN.md # macOS development notes
│   ├── RELEASE_NOTES.md       # Release notes
│   ├── TESTING_REPORT.md      # Testing reports
│   └── ICON_README.md         # Icon documentation
├── archive/                   # Archived files
├── build/                     # Build output directory
├── resources/                 # Platform-specific resources
│   ├── windows/
│   ├── darwin/
│   └── common/
└── .github/                   # GitHub workflows and templates
    └── workflows/
        └── build.yml          # CI/CD pipeline
```

## 🔧 Configuration

The launcher stores configuration in platform-specific locations:

- **Windows**: `%LOCALAPPDATA%\TheBoysLauncher`
- **macOS**: `~/Library/Application Support/TheBoysLauncher`
- **Linux**: `~/.theboyslauncher`

### Configuration Options
- **Memory Allocation**: Automatic detection with manual override
- **Java Version**: Automatically downloads compatible Java runtime
- **Update Settings**: Configure automatic update behavior
- **Modpack Sources**: Add custom modpack repositories

## 🐛 Troubleshooting

### Windows Issues
- **Antivirus False Positives**: Some antivirus software may flag the launcher. Add an exception if needed.
- **Missing .NET Framework**: Ensure .NET Framework 4.7.2+ is installed.

### macOS Issues
- **"App is damaged" error**: The app isn't signed. Run `xattr -cr /Applications/TheBoysLauncher.app` in Terminal.
- **Permission denied**: Ensure the app has execute permissions: `chmod +x /Applications/TheBoysLauncher.app/Contents/MacOS/TheBoysLauncher`

### Linux Issues
- **Missing dependencies**: Install required system packages:
  ```bash
  # Ubuntu/Debian
  sudo apt-get install libc6 libgl1-mesa-glx libxcomposite1 libxcursor1 libxrandr1 libxtst6 libxi6

  # Fedora
  sudo dnf install mesa-libGL libXcomposite libXcursor libXrandr libXtst libXi
  ```

### General Issues
- **Java not found**: The launcher automatically downloads Java, but you can specify a custom Java path in settings.
- **Launcher fails to start**: Check the logs in the launcher's data directory for detailed error information.

## 🧪 Testing

Run the comprehensive test suite:
```bash
# Run cross-platform tests
./scripts/test-cross-platform.sh

# Run specific tests
make verify     # Build verification
make lint       # Code formatting and vetting
make test       # Full test suite
```

## 📊 Development Status

This project has been converted from Windows-only to full cross-platform support. See [MACOS_DEVELOPMENT_PLAN.md](./MACOS_DEVELOPMENT_PLAN.md) for detailed development information and [TESTING_REPORT.md](./TESTING_REPORT.md) for comprehensive testing results.

### ✅ Completed Features
- [x] Cross-platform builds (Windows, macOS Intel/ARM64, Linux)
- [x] Platform-specific memory detection
- [x] Cross-platform process management
- [x] Automatic Java runtime management
- [x] Prism Launcher integration
- [x] Self-updating mechanism
- [x] macOS app bundles and DMG creation
- [x] Comprehensive testing framework

### 🔄 In Development
- [ ] Linux packaging and distribution
- [ ] Code signing for distribution
- [ ] Additional modpack sources

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Follow Go best practices and formatting (`go fmt ./...`)
- Ensure cross-platform compatibility
- Add tests for new functionality
- Update documentation as needed

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Prism Launcher](https://prismlauncher.org/)** - Excellent Minecraft launcher
- **[Fyne](https://fyne.io/)** - Cross-platform GUI framework
- **[Adoptium](https://adoptium.net/)** - Java runtime distribution
- **[Go](https://golang.org/)** - The Go programming language

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/dilllxd/theboyslauncher/issues)
- **Discussions**: [GitHub Discussions](https://github.com/dilllxd/theboyslauncher/discussions)
- **Releases**: [GitHub Releases](https://github.com/dilllxd/theboyslauncher/releases)

---

**TheBoysLauncher** - Simplifying Minecraft modpack management across all platforms. 🚀