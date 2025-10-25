# Release Notes

## v3.0.1 - Cross-Platform Release 🎉

### 🚀 Major Features

#### Cross-Platform Support
- **✅ Windows** - Full Windows support with executable and installer
- **✅ macOS Intel** - Native support for Intel-based Macs
- **✅ macOS Apple Silicon** - Native support for M1/M2/M3 Macs
- **✅ macOS Universal** - Single binary that works on all Macs
- **✅ Linux** - Basic Linux support (from source)

#### Platform Abstraction
- **Memory Detection** - Automatic system memory detection on all platforms
- **Process Management** - Platform-appropriate process termination
- **Path Handling** - Cross-platform file path and environment variable management
- **Executable Naming** - Platform-specific executable names (.exe vs no extension)

### 🔧 Technical Improvements

#### Build System
- **Unified Makefile** - Single build system for all platforms
- **Cross-Compilation** - Build any platform from any other platform
- **GitHub Actions CI/CD** - Automated builds and testing
- **Optimized Caching** - Faster builds with intelligent caching

#### Application Architecture
- **Platform-Specific Files** - Clean separation using Go build tags
  - `platform_windows.go` - Windows-specific implementations
  - `platform_darwin.go` - macOS-specific implementations
  - `memory_windows.go` / `memory_darwin.go` - Platform memory detection
  - `process_windows.go` / `process_darwin.go` - Platform process management

#### Integration Updates
- **Java Runtime Management** - Cross-platform Adoptium API integration
- **Prism Launcher Integration** - Automatic detection and configuration
- **Update System** - Self-updating mechanism with macOS quarantine handling
- **Archive Extraction** - Support for both .zip and .tar.gz formats

### 📦 Packaging

#### macOS
- **App Bundles** - Proper .app bundles with Info.plist
- **Icon Conversion** - Automatic ICO to ICNS conversion
- **DMG Creation** - Professional DMG installers with proper layout
- **Universal Binaries** - Single binary for Intel and Apple Silicon

#### Windows
- **Executable** - Optimized Windows executable
- **Icon Integration** - Windows icon support maintained

### 🧪 Testing

#### Comprehensive Test Suite
- **Cross-Platform Tests** - 16 different test categories
- **Build Verification** - Automatic build validation
- **API Testing** - Adoptium and GitHub API integration testing
- **Platform Validation** - Platform-specific functionality testing

#### Test Coverage
- ✅ Basic compilation and platform detection
- ✅ Configuration system
- ✅ Java API integration
- ✅ Archive extraction capabilities
- ✅ File operations and permissions
- ✅ Memory detection
- ✅ Executable naming
- ✅ PATH environment handling
- ✅ Update system functions

### 🛠️ Development Tools

#### Scripts
- `create-app-bundle.sh` - macOS app bundle creation
- `convert-icon.sh` - Icon conversion between formats
- `create-dmg.sh` - DMG installer creation
- `test-cross-platform.sh` - Comprehensive testing

#### Documentation
- `README.md` - Complete user documentation
- `BUILD.md` - Comprehensive build guide
- `INSTALL_MACOS.md` - Detailed macOS installation guide
- `TESTING_REPORT.md` - Complete testing results
- `MACOS_DEVELOPMENT_PLAN.md` - Full development documentation

### 🔄 Breaking Changes

#### Windows Users
- **None** - Full backward compatibility maintained

#### macOS Users
- **New Application** - Previously Windows-only, now available on macOS
- **App Store Not Available** - Distributed as DMG due to code signing requirements

#### Developers
- **Build System** - New Makefile-based build system
- **Platform Files** - New platform-specific source files
- **Dependencies** - Go 1.22+ now required

### 🔒 Security

#### Code Signing
- **Windows** - No code signing (unsigned distribution)
- **macOS** - No code signing (requires user approval on first launch)
- **Recommendation** - Users may need to bypass Gatekeeper on macOS

#### Network Security
- **HTTPS Only** - All API calls use secure connections
- **Checksum Verification** - Automatic verification of downloaded files

### 🐛 Bug Fixes

#### Platform Issues
- **Fixed** - Windows hard block preventing non-Windows execution
- **Fixed** - Hardcoded paths and environment variables
- **Fixed** - Platform-specific executable naming issues
- **Fixed** - Memory detection on non-Windows platforms
- **Fixed** - Process management on macOS

#### Build Issues
- **Fixed** - Cross-platform compilation errors
- **Fixed** - CGO dependency issues
- **Fixed** - Build constraint conflicts
- **Fixed** - Icon path issues in builds

### 📊 Performance

#### Improvements
- **Faster Builds** - Optimized build system with caching
- **Smaller Binaries** - Improved build flags for size optimization
- **Better Memory Detection** - More accurate system memory detection
- **Optimized Downloads** - Parallel download capabilities

### 🔮 Future Roadmap

#### v3.1.0 (Planned)
- **Linux Packaging** - Native Linux packages (deb, rpm, AppImage)
- **Code Signing** - Optional code signing for Windows and macOS
- **Auto-Update UI** - Improved update interface
- **More Modpack Sources** - Additional modpack repository support

#### v3.2.0 (Future)
- **Plugin System** - Extensible plugin architecture
- **Theme Support** - Custom themes and UI customization
- **Advanced Settings** - More configuration options
- **Performance Monitoring** - Built-in performance metrics

### 🙏 Acknowledgments

#### Special Thanks
- **Go Community** - Excellent cross-platform development tools
- **Fyne Team** - Amazing cross-platform GUI framework
- **Adoptium** - Reliable Java runtime distribution
- **Prism Launcher** - Excellent Minecraft launcher foundation
- **Testers** - Everyone who helped test the cross-platform versions

### 📞 Support

#### Getting Help
- **GitHub Issues** - [Report bugs and request features](https://github.com/dilllxd/theboyslauncher/issues)
- **GitHub Discussions** - [Community discussions](https://github.com/dilllxd/theboyslauncher/discussions)
- **Documentation** - [Complete documentation](https://github.com/dilllxd/theboyslauncher/tree/macos-support)

#### Contributing
- **Pull Requests** - Welcome for bug fixes and features
- **Testing** - Help test on different platforms
- **Documentation** - Improve documentation and guides

---

## Download Links

### Latest Release (v3.0.1)

#### Windows
- [TheBoysLauncher-Setup.exe](https://github.com/dilllxd/theboyslauncher/releases/download/v3.0.1/TheBoysLauncher-Setup.exe)

#### macOS
- [TheBoysLauncher-Universal.dmg](https://github.com/dilllxd/theboyslauncher/releases/download/v3.0.1/TheBoysLauncher-Universal.dmg)
- [TheBoysLauncher-Intel.zip](https://github.com/dilllxd/theboyslauncher/releases/download/v3.0.1/TheBoysLauncher-Intel.zip)
- [TheBoysLauncher-AppleSilicon.zip](https://github.com/dilllxd/theboyslauncher/releases/download/v3.0.1/TheBoysLauncher-AppleSilicon.zip)

#### Source Code
- [Source Code (tar.gz)](https://github.com/dilllxd/theboyslauncher/archive/refs/tags/v3.0.1.tar.gz)
- [Source Code (zip)](https://github.com/dilllxd/theboyslauncher/archive/refs/tags/v3.0.1.zip)

### Build from Source
See [BUILD.md](./BUILD.md) for comprehensive build instructions.

---

**Thank you for using TheBoysLauncher!** 🚀

This release represents months of work to bring TheBoysLauncher to multiple platforms while maintaining the simplicity and reliability that users expect.