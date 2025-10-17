# TheBoys Launcher - Build Guide

## Single-File Deployment ✅

**YES!** TheBoys Launcher supports single-file deployment just like the legacy Winterpack Launcher. Each executable is self-contained and creates all required files in the same directory (Windows) or user home directory (macOS/Linux).

### Portable Operation

- **Windows**: All files created beside the `.exe` (fully portable)
- **macOS**: Files created in `~/.theboys-launcher/` (user home)
- **Linux**: Files created in `~/.theboys-launcher/` (user home)

No installation required! Just download and run.

## Quick Start

### Option 1: Build Script (Recommended)

#### Linux/macOS
```bash
# Build for all platforms
./build-all.sh v1.0.0

# Or use Makefile
make all VERSION=v1.0.0
```

#### Windows
```batch
# Build for all platforms
build-all.bat v1.0.0
```

### Option 2: Manual Build

```bash
# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Build frontend
cd frontend && npm run build && cd ..

# Build for current platform
go build -ldflags="-s -w -X main.version=v1.0.0" -o theboys-launcher ./cmd/launcher
```

## Build Requirements

### Required Tools
- **Go 1.23+** - Go programming language
- **Node.js 18+** - Frontend build system
- **NPM** - Package manager (comes with Node.js)

### Optional Tools
- **Wails CLI** - For development mode
- **Make** - For convenient build targets (Linux/macOS)
- **Git** - For version control

## Build Targets

### Available Executables

| Platform | Architecture | Output File | Description |
|----------|--------------|-------------|-------------|
| Windows | x64 | `TheBoysLauncher.exe` | Main Windows executable |
| Windows | ARM64 | `TheBoysLauncher-arm64.exe` | Windows ARM64 (Surface, etc.) |
| Linux | x64 | `theboys-launcher-linux-amd64` | Standard Linux builds |
| Linux | ARM64 | `theboys-launcher-linux-arm64` | ARM64 Linux (Pi, servers) |
| macOS | Intel | `theboys-launcher-macos-amd64` | Intel Macs |
| macOS | Apple Silicon | `theboys-launcher-macos-arm64` | M1/M2/M3 Macs |

### Build Commands

#### Using Makefile (Linux/macOS)
```bash
# Build everything
make all VERSION=v1.0.0

# Build specific platforms
make build-windows    # Windows (x64 + ARM64)
make build-linux      # Linux (x64 + ARM64)
make build-macos      # macOS (Intel + Apple Silicon)

# Quick build for current platform
make quick VERSION=v1.0.0

# Development mode
make dev

# Run tests
make test

# Clean build artifacts
make clean
```

#### Using Build Scripts
```bash
# Cross-platform build script
./build-all.sh v1.0.0    # Linux/macOS
build-all.bat v1.0.0      # Windows
```

## Development

### Prerequisites
```bash
# Install Wails CLI (optional, for dev mode)
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Development Mode
```bash
# Start development server with hot reload
make dev
# or
wails dev
```

### Project Structure
```
theboys-launcher/
├── cmd/launcher/          # Main application entry point
├── internal/              # Internal packages
│   ├── app/              # Main application logic
│   ├── config/           # Configuration management
│   ├── launcher/         # Core launcher functionality
│   ├── platform/         # Platform-specific code
│   ├── gui/              # GUI frontend
│   └── logging/          # Logging system
├── frontend/             # Web-based GUI
│   ├── src/              # Frontend source code
│   ├── dist/             # Built frontend assets
│   └── package.json      # Frontend dependencies
├── pkg/types/            # Shared types
├── wails.json           # Wails configuration
├── build-all.sh         # Cross-platform build script
├── build-all.bat        # Windows build script
├── Makefile             # Build targets
└── README.md            # This file
```

## Continuous Integration

### GitHub Actions
The project includes automated builds via GitHub Actions:

- **Testing**: Runs on all pull requests
- **Building**: Cross-platform builds on push/tags
- **Releases**: Automatic release creation for tags
- **Docker**: Optional container builds

### Local CI
```bash
# Run tests
make test

# Lint code
make lint

# Full CI pipeline
make ci
```

## Deployment

### Single-File Distribution
Each build produces a **single, self-contained executable**:

1. **No external dependencies** - Everything is embedded
2. **No installation required** - Just download and run
3. **Portable operation** - Works from any directory
4. **Self-updating** - Built-in update mechanism

### File Locations
The launcher creates files in these locations:

#### Windows (Portable)
```
TheBoysLauncher.exe    # Main executable
prism/                 # Prism Launcher portable
util/                  # Utilities (Java, packwiz)
instances/             # Minecraft instances
config/                # Configuration files
logs/                  # Log files
```

#### macOS/Linux (User Directory)
```
~/.theboys-launcher/
├── prism/              # Prism Launcher
├── util/               # Utilities
├── instances/          # Minecraft instances
├── config/             # Configuration
└── logs/               # Log files
```

## Build Optimization

### Size Optimization
The builds use several optimization techniques:

- **Strip symbols**: `-s -w` ldflags
- **UPX compression** (optional): Further reduces executable size
- **Embedded assets**: Frontend embedded in executable
- **Static linking**: No external dependencies

### Performance Optimization
- **Cross-platform**: Optimized for each target platform
- **Concurrent downloads**: Faster modpack downloads
- **Efficient memory management**: Better resource usage
- **Smart caching**: Reduced redundant downloads

## Troubleshooting

### Common Build Issues

#### "Go not found"
```bash
# Install Go from https://golang.org/dl/
# or use package manager:
brew install go        # macOS
sudo apt install golang  # Ubuntu/Debian
```

#### "Node.js not found"
```bash
# Install Node.js from https://nodejs.org/
# or use package manager:
brew install node       # macOS
sudo apt install nodejs npm  # Ubuntu/Debian
```

#### "Frontend build failed"
```bash
# Clear frontend cache and rebuild
cd frontend
rm -rf node_modules package-lock.json
npm install
npm run build
```

#### "Cross-compilation failed"
```bash
# Install cross-compilation tools
# For Windows targets on Linux:
sudo apt install gcc-mingw-w64
```

### Runtime Issues

#### "Permission denied" (Linux/macOS)
```bash
# Make executable
chmod +x theboys-launcher-linux-amd64
```

#### "Cannot open display" (Linux GUI)
```bash
# Install X11 libraries
sudo apt install libx11-dev libgtk-3-dev libwebkit2gtk-4.0-dev
```

## Version Management

### Version Injection
Build scripts automatically inject version information:

```bash
# Build with version
./build-all.sh v1.2.3

# Version will be available in:
# - GUI title bar
# - CLI --version flag
# - Help text
# - Update checks
```

### Release Process
1. Update version in code
2. Create git tag: `git tag v1.2.3`
3. Push tag: `git push origin v1.2.3`
4. GitHub Actions builds and creates release
5. Download executables from releases

## Support

### Getting Help
- **Issues**: Report bugs via GitHub Issues
- **Documentation**: Check this guide and code comments
- **Community**: Join discussions in GitHub Discussions

### Legacy Compatibility
This launcher maintains **100% compatibility** with the legacy Winterpack Launcher:

- All command-line arguments preserved
- Same configuration format support
- Compatible modpack formats
- Portable deployment model maintained

### Migration from Legacy
1. Download the new launcher for your platform
2. Place it in the same directory as the old launcher
3. Run the new launcher - it will detect existing configurations
4. All your instances and settings will be preserved