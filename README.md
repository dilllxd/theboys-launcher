# TheBoys Launcher

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-repo/theboys-launcher)

A modern, cross-platform Minecraft modpack launcher built with Go and Fyne. TheBoys Launcher provides a beautiful, user-friendly interface for managing and launching Minecraft modpacks across Windows, macOS, and Linux.

## 🚀 Features

### Core Functionality
- **Cross-platform support**: Windows, macOS, and Linux
- **Modern GUI**: Beautiful native interface built with Fyne
- **Modpack management**: Download and launch multiple modpacks
- **Automatic dependency management**: Java runtime, Prism launcher, packwiz
- **Self-updating**: Automatic updates with rollback support

### User Experience
- **Intuitive interface**: Card-based modpack selection
- **Progress tracking**: Real-time download and installation progress
- **Configurable settings**: Memory allocation, themes, update channels
- **Dark/Light themes**: System theme integration
- **Responsive design**: Works on different screen sizes

### Technical Features
- **Robust error handling**: Graceful failure recovery
- **Comprehensive logging**: Debug information for troubleshooting
- **Cross-platform paths**: Proper file handling on all platforms
- **Concurrent downloads**: Faster installation with parallel downloads
- **Memory efficient**: Optimized for performance

## 📦 Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/your-repo/theboys-launcher/releases).

- **Windows**: `theboys-launcher-windows-amd64.exe`
- **Linux**: `theboys-launcher-linux-amd64`
- **macOS**: `theboys-launcher-darwin-amd64` (Intel) or `theboys-launcher-darwin-arm64` (Apple Silicon)

### Build from Source

#### Prerequisites
- Go 1.22 or later
- Make (optional, for build automation)

#### Quick Build
```bash
git clone https://github.com/your-repo/theboys-launcher.git
cd theboys-launcher
make build
```

#### Cross-platform Build
```bash
make build-all
```

This will create binaries for all supported platforms in the `build/` directory.

### Development Setup
```bash
# Clone the repository
git clone https://github.com/your-repo/theboys-launcher.git
cd theboys-launcher

# Install dependencies
make deps

# Run in development mode
make dev
```

## 🎮 Usage

### First Launch

1. **Launch the application**: Run the TheBoys Launcher executable
2. **Select a modpack**: Choose from available modpacks in the main interface
3. **Configure settings** (optional): Adjust memory allocation, Java settings, etc.
4. **Launch**: Click the launch button to start the selected modpack

### Configuration

TheBoys Launcher automatically handles:

- **Java Runtime Detection**: Downloads appropriate Java version if needed
- **Prism Launcher**: Downloads and configures Prism Launcher
- **Modpack Installation**: Uses packwiz for modpack management
- **Instance Management**: Creates and manages Minecraft instances

### Settings

Access the Settings window to configure:

- **Memory allocation**: 512MB - 32GB (default: 4GB)
- **Theme**: Light, Dark, or System
- **Update channel**: Stable, Beta, or Alpha
- **Network settings**: Download timeout, concurrent downloads
- **Advanced options**: Custom paths, debug logging

## 🔧 Development

### Project Structure

```
theboys-launcher/
├── cmd/theboys/             # Application entry point
├── internal/                # Internal packages
│   ├── app/                 # Core application logic
│   ├── config/              # Configuration management
│   ├── gui/                 # GUI components
│   ├── launcher/            # Minecraft launcher logic
│   ├── logging/             # Logging system
│   └── updater/             # Self-updater
├── pkg/                     # Public packages
│   ├── version/             # Version information
│   └── platform/            # Platform detection
├── assets/                  # Application assets
├── configs/                 # Configuration files
├── docs/                    # Documentation
└── legacy/                  # Original code for reference
```

### Build Commands

```bash
# Development
make dev                    # Run in development mode
make dev-debug             # Run with debug logging

# Building
make build                 # Build for current platform
make build-all             # Build for all platforms
make build-windows         # Build for Windows
make build-linux           # Build for Linux
make build-darwin          # Build for macOS

# Testing
make test                  # Run tests
make test-coverage         # Run tests with coverage

# Code Quality
make fmt                   # Format code
make vet                   # Run go vet
make lint                  # Run linter
make security              # Run security checks

# Dependencies
make deps                  # Install dependencies
make deps-update           # Update dependencies

# Release
make release               # Create release packages
make clean                 # Clean build artifacts
```

### Adding Features

1. **GUI Components**: Add new widgets in `internal/gui/widgets/`
2. **Launcher Logic**: Extend `internal/launcher/` for new Minecraft features
3. **Configuration**: Update `internal/config/` for new settings
4. **Platform Support**: Enhance `pkg/platform/` for platform-specific code

## 🐛 Troubleshooting

### Common Issues

#### Application Won't Start
- **Check logs**: Look at the log files in your system's temp directory
- **Verify dependencies**: Ensure you have a supported Go version if building from source
- **Platform support**: Verify your platform is supported

#### Java Issues
- **Automatic detection**: The launcher will download Java if needed
- **Custom path**: Set a custom Java path in Settings > Advanced
- **Version conflicts**: The launcher manages multiple Java versions automatically

#### Modpack Issues
- **Network connection**: Ensure you have an internet connection for downloads
- **Disk space**: Verify you have sufficient disk space for modpacks
- **Permissions**: Make sure the application has write permissions

### Getting Help

- **GitHub Issues**: [Report bugs](https://github.com/your-repo/theboys-launcher/issues)
- **Discussions**: [Feature requests and discussions](https://github.com/your-repo/theboys-launcher/discussions)
- **Logs**: Include log files when reporting issues

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Run tests: `make test`
5. Check code quality: `make fmt vet lint`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Fyne** - Beautiful cross-platform GUI framework for Go
- **Prism Launcher** - Excellent Minecraft launcher we integrate with
- **Packwiz** - Modpack management tool
- **Go Community** - For the amazing ecosystem and tools

## 📊 Version History

- **v2.0.0** - Complete rewrite with modern GUI and cross-platform support
- **v1.x.x** - Original Windows-only TUI version (see `legacy/` directory)

---

**TheBoys Launcher** - Making Minecraft modpacks accessible to everyone. 🎮👥