# TheBoys Launcher

A modern, portable Minecraft modpack launcher built with Rust and Tauri.

## 🚀 Features

- **Cross-Platform Support**: Windows, macOS, and Linux
- **Modern UI**: Clean, responsive interface built with React and TypeScript
- **Auto-Updater**: Background update checking and silent installation
- **Modpack Management**: Browse, install, and manage Minecraft modpacks
- **Java Integration**: Automatic Java detection and installation
- **Prism Launcher Integration**: Seamless integration with Prism Launcher
- **Performance Monitoring**: Built-in performance metrics and monitoring

## 📋 System Requirements

### Windows
- Windows 10/11 (x64)
- 4GB RAM minimum
- 2GB disk space

### macOS
- macOS 11+ (Intel or Apple Silicon)
- 4GB RAM minimum
- 2GB disk space

### Linux
- x64 Linux distribution
- 4GB RAM minimum
- 2GB disk space
- libwebkit2gtk-4.1-0 or equivalent

## 🔧 Installation

### From Release Packages

1. Download the appropriate package for your platform from [Releases](https://github.com/theboys/launcher/releases)
2. Run the installer and follow the prompts
3. Launch TheBoys Launcher from your applications menu

### Development Build

```bash
# Clone the repository
git clone https://github.com/theboys/launcher.git
cd launcher

# Install dependencies
npm install

# Run in development mode
npm run tauri:dev

# Build for production
npm run tauri:build
```

## 🎯 Quick Start

1. **Configure Settings**: Set your preferred memory allocation and paths
2. **Install Dependencies**: The launcher will guide you through Java and Prism installation
3. **Browse Modpacks**: Explore available modpacks in the library
4. **Install & Play**: Select a modpack and click "Install & Play"

## 📦 Package Formats

- **Windows**: MSI installer (recommended) or NSIS installer
- **macOS**: DMG package with drag-and-drop installation
- **Linux**: AppImage (portable), DEB package (Debian/Ubuntu), RPM package (Fedora/RHEL)

## 🔄 Updates

The launcher automatically checks for updates on startup. You can configure:
- Automatic download and installation
- Update notifications
- Prerelease updates
- Update channel (stable/beta)

## 🛠️ Development

### Prerequisites
- Node.js 18+
- Rust 1.70+
- Tauri CLI: `npm install -g @tauri-apps/cli`

### Build Commands

```bash
# Development server
npm run tauri:dev

# Production build (current platform)
npm run tauri:build

# Full release build (all platforms)
npm run build:release

# Run tests
npm test

# Code linting
npm run lint
```

### Project Structure

```
src/                 # Frontend React application
src-tauri/           # Rust backend
├── src/
│   ├── commands/    # Tauri commands
│   ├── models/      # Data structures
│   ├── utils/       # Utility modules
│   └── main.rs      # Application entry point
├── Cargo.toml       # Rust dependencies
└── tauri.conf.json  # Tauri configuration
assets/              # Platform integration files
icons/               # Application icons
```

## 📚 Documentation

- [Release Process](./RELEASE_PROCESS.md) - Complete release and distribution guide
- [AI Implementation](./AI_IMPLEMENTATION_PROMPT.md) - Development guidelines
- [Code Quality](./CODE_QUALITY_AND_DOCUMENTATION.md) - Code standards and practices

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 🐛 Bug Reports

Please report bugs through [GitHub Issues](https://github.com/theboys/launcher/issues) with:
- Operating system and version
- Launcher version
- Steps to reproduce
- Expected vs actual behavior
- Any relevant logs or screenshots

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## 🙏 Acknowledgments

- [Tauri](https://tauri.app/) - Application framework
- [React](https://reactjs.org/) - UI framework
- [Prism Launcher](https://prismlauncher.org/) - Minecraft launcher integration
- [Minecraft](https://www.minecraft.net/) - Game platform

## 📞 Support

- [Documentation](https://docs.theboys-launcher.com)
- [Discord](https://discord.gg/theboys)
- [GitHub Issues](https://github.com/theboys/launcher/issues)

---

**TheBoys Launcher** - Making Minecraft modpack management simple and elegant.