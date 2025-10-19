# TheBoys Launcher - macOS Support Implementation Plan

## Executive Summary

This document outlines a comprehensive plan to convert TheBoys Launcher from Windows-only to full macOS support while maintaining 100% feature parity. The project is currently entirely Windows-dependent with hardcoded paths, Windows API calls, and Windows-specific build systems.

## Current Windows Features Requiring macOS Alternatives

### 1. Core System Integration
- **Windows API Calls**: Memory detection, process management, console hiding
- **Executable Names**: `.exe` extensions, Windows-specific binaries
- **File Paths**: Windows path separators, hardcoded directories
- **Process Management**: `taskkill` command, Windows process trees

### 2. User Interface
- **Fyne GUI Framework**: ✅ Already cross-platform compatible
- **Window Management**: Windows-specific centering and behavior
- **System Tray**: Windows-only implementation
- **Icon Handling**: Windows `.ico` format

### 3. Build System
- **Build Scripts**: PowerShell (.ps1) and Batch (.bat) files
- **Resource Compilation**: Windows resource files (.rc)
- **Code Signing**: Windows-specific certificates
- **Installer Creation**: Inno Setup (.iss)

### 4. Runtime Dependencies
- **Java Downloads**: Windows-specific Adoptium API calls
- **Prism Launcher**: Windows executable names and download patterns
- **Packwiz Bootstrap**: Windows executable handling

### 5. Self-Update System
- **Executable Replacement**: Windows file locking behavior
- **Process Management**: Windows-specific restart mechanisms

## Implementation Phases

### Phase 1: Foundation and Platform Abstraction (Week 1-2)

#### 1.1 Platform Detection and Constants
**Files to Create:**
- `platform_darwin.go` - macOS-specific implementations
- `platform_windows.go` - Windows-specific implementations (refactor existing)
- `platform.go` - Common platform interface

**Platform Constants:**
```go
// Platform-specific executable names
var (
    launcherExeName = "TheBoysLauncher.exe" // Windows
    prismExeName    = "PrismLauncher.exe"  // Windows
    javaBinName     = "java.exe"           // Windows
    javawBinName    = "javaw.exe"          // Windows
)

// macOS equivalents:
var (
    launcherExeName = "TheBoysLauncher"    // macOS
    prismExeName    = "PrismLauncher"     // macOS
    javaBinName     = "java"              // macOS
    javawBinName    = "java"              // macOS (no javaw equivalent)
)
```

#### 1.2 Path Management Abstraction
**Files to Modify:**
- `utils.go` - Add platform-specific path handling
- All files using `filepath.Join()` - Review for platform assumptions

**Implementation:**
```go
// Platform-specific directory structures
func getLauncherHome() string {
    switch runtime.GOOS {
    case "darwin":
        return filepath.Join(os.Getenv("HOME"), "Library", "Application Support", "TheBoysLauncher")
    case "windows":
        return filepath.Join(os.Getenv("USERPROFILE"), ".theboys-launcher")
    default:
        return filepath.Join(os.Getenv("HOME"), ".theboys-launcher")
    }
}
```

#### 1.3 Process Management Abstraction
**Files to Create:**
- `process_windows.go` - Windows process management
- `process_darwin.go` - macOS process management

**Implementation:**
```go
// macOS process management using sys/unix
func killProcessTree(pid int) error {
    // Use pgrep and pkill on macOS
    cmd := exec.Command("pkill", "-P", fmt.Sprintf("%d", pid))
    return cmd.Run()
}

func hideConsoleWindow() {
    // macOS doesn't need console hiding for GUI apps
}
```

### Phase 2: System Integration (Week 3-4)

#### 2.1 Memory Detection
**Files to Modify:**
- `config.go` - Replace Windows memory detection

**Implementation:**
```go
// macOS memory detection using sysctl
func totalRAMMB() int {
    if runtime.GOOS == "darwin" {
        // Use sysctl to get physical memory on macOS
        var totalMemory uint64
        size := uint64(8)
        if err := sysctlbyname("hw.memsize", unsafe.Pointer(&totalMemory), &size, nil, 0); err != nil {
            return 8192 // Fallback to 8GB
        }
        return int(totalMemory / (1024 * 1024))
    }
    // Existing Windows implementation...
}
```

#### 2.2 Java Runtime Management
**Files to Modify:**
- `java.go` - Update Adoptium API calls for macOS
- `launcher.go` - Java binary detection

**Implementation:**
```go
// Update Java download URLs for macOS
func fetchJREURL(javaVersion string) (string, error) {
    var osName string
    switch runtime.GOOS {
    case "darwin":
        osName = "mac"
    case "windows":
        osName = "windows"
    default:
        osName = "linux"
    }

    adoptium := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=x64&image_type=%s&os=%s",
        javaVersion, imageType, osName)
    // ... rest of implementation
}

// Java binary detection for macOS
func getJavaBinaries(jreDir string) (javaBin, javawBin string) {
    if runtime.GOOS == "darwin" {
        javaBin = filepath.Join(jreDir, "bin", "java")
        javawBin = javaBin // No javaw equivalent on macOS
        return
    }
    // Windows implementation...
}
```

#### 2.3 Prism Launcher Integration
**Files to Modify:**
- `prism.go` - Update download patterns for macOS
- `launcher.go` - Prism executable detection

**Implementation:**
```go
// macOS Prism download patterns with architecture support
func getPrismDownloadPatterns(latestTag string) []string {
    switch runtime.GOOS {
    case "darwin":
        // Check architecture for macOS
        arch := "arm64"
        if runtime.GOARCH == "amd64" {
            arch = "x86_64"
        }
        return []string{
            fmt.Sprintf("PrismLauncher-macOS-%s-%s.tar.gz", latestTag, arch),
            fmt.Sprintf("PrismLauncher-macos-%s-%s.tar.gz", latestTag, arch),
            fmt.Sprintf("PrismLauncher-darwin-%s-%s.tar.gz", latestTag, arch),
            fmt.Sprintf("PrismLauncher-macOS-%s.tar.gz", latestTag), // Fallback
        }
    case "windows":
        // Existing Windows patterns
    }
}

func ensurePrism(dir string) (bool, error) {
    var prismExe string
    switch runtime.GOOS {
    case "darwin":
        prismExe = filepath.Join(dir, "PrismLauncher.app", "Contents", "MacOS", "PrismLauncher")
    case "windows":
        prismExe = filepath.Join(dir, "PrismLauncher.exe")
    }

    if exists(prismExe) {
        return false, nil
    }
    // ... download and extract logic
}
```

### Phase 3: Build System Overhaul (Week 5-6)

#### 3.1 Cross-Platform Build Scripts
**Files to Create:**
- `build.sh` - macOS/Linux build script
- `build-macos.sh` - macOS-specific build
- `Makefile` - Update with cross-platform targets

**Implementation:**
```makefile
# Makefile updates
.PHONY: build-macos
build-macos:
	@echo "Building TheBoys Launcher for macOS Intel..."
	@mkdir -p build/amd64
	@export GOOS=darwin GOARCH=amd64 CGO_ENABLED=1
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/amd64/TheBoysLauncher .
	@echo "Build successful! Created TheBoysLauncher for macOS Intel"

.PHONY: build-macos-arm64
build-macos-arm64:
	@echo "Building TheBoys Launcher for macOS Apple Silicon..."
	@mkdir -p build/arm64
	@export GOOS=darwin GOARCH=arm64 CGO_ENABLED=1
	@go build -ldflags="-s -w -X main.version=$(VERSION)" -o build/arm64/TheBoysLauncher .
	@echo "Build successful! Created TheBoysLauncher for macOS Apple Silicon"

.PHONY: build-macos-all
build-macos-all: build-macos build-macos-arm64
	@echo "Built both Intel and Apple Silicon versions"

# Package both architectures
.PHONY: package-macos-all
package-macos-all: build-macos-all
	@echo "Creating packages for both architectures..."
	@./scripts/create-app-bundle.sh amd64
	@./scripts/create-app-bundle.sh arm64
	@./scripts/create-dmg.sh amd64
	@./scripts/create-dmg.sh arm64
	@echo "Created Intel and Apple Silicon packages"
```

#### 3.2 Icon and Resource Management
**Files to Create:**
- `resources/darwin/icon.icns` - macOS icon format
- `build/resources.go` - Resource embedding logic

**Implementation:**
```go
//go:generate go run build/embed-resources.go

// Platform-specific icon embedding
func embedIcon() error {
    switch runtime.GOOS {
    case "darwin":
        return embedDarwinIcon("resources/darwin/icon.icns")
    case "windows":
        return embedWindowsIcon("icon.ico")
    }
    return nil
}
```

#### 3.3 Dual Architecture Support
**Implementation:**
```makefile
# Build both architectures
.PHONY: build-macos-all
build-macos-all: build-macos build-macos-arm64
	@echo "Built both Intel and Apple Silicon versions"

# Create universal binary
.PHONY: build-macos-universal
build-macos-universal:
	@echo "Creating universal binary..."
	@mkdir -p build/universal
	@lipo -create build/amd64/TheBoysLauncher -output build/universal/TheBoysLauncher
	@echo "Universal binary created: build/universal/TheBoysLauncher"
```

### Phase 4: Installer and Distribution (Week 7-8)

#### 4.1 macOS Application Bundle
**Files to Create:**
- `build/Info.plist` - macOS app bundle metadata
- `build/entitlements.plist` - macOS app entitlements
- `scripts/create-app-bundle.sh` - App bundle creation script

**Implementation:**
```xml
<!-- Info.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDisplayName</key>
    <string>TheBoys Launcher</string>
    <key>CFBundleExecutable</key>
    <string>TheBoysLauncher</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
    <key>CFBundleIdentifier</key>
    <string>com.theboys.launcher</string>
    <key>CFBundleVersion</key>
    <string>3.0.0</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSSupportsAutomaticGraphicsSwitching</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
```

#### 4.2 DMG Creation
**Files to Create:**
- `scripts/create-dmg.sh` - DMG creation script
- `scripts/create-universal-dmg.sh` - Universal DMG creation

**Implementation:**
```bash
#!/bin/bash
# create-dmg.sh
echo "Creating macOS DMG for $ARCH..."

# Create app bundle
./scripts/create-app-bundle.sh $ARCH

# Create DMG
create-dmg \
  --volname "TheBoys Launcher ($ARCH)" \
  --volicon "build/$ARCH/TheBoysLauncher.app/Contents/Resources/AppIcon.icns" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --icon "TheBoysLauncher.app" 175 120 \
  --hide-extension "TheBoysLauncher.app" \
  --app-drop-link 425 120 \
  "TheBoys-Launcher-$VERSION-$ARCH.dmg" \
  "build/$ARCH/" \
  || exit 1

echo "DMG created successfully: TheBoys-Launcher-$VERSION-$ARCH.dmg"
```

```bash
#!/bin/bash
# create-universal-dmg.sh
echo "Creating Universal macOS DMG..."

# Build universal binary first
make build-macos-universal

# Create universal app bundle
./scripts/create-app-bundle.sh universal

# Create universal DMG
create-dmg \
  --volname "TheBoys Launcher" \
  --volicon "build/universal/TheBoysLauncher.app/Contents/Resources/AppIcon.icns" \
  --window-pos 200 120 \
  --window-size 600 400 \
  --icon-size 100 \
  --icon "TheBoysLauncher.app" 175 120 \
  --hide-extension "TheBoysLauncher.app" \
  --app-drop-link 425 120 \
  "TheBoys-Launcher-$VERSION-Universal.dmg" \
  "build/universal/" \
  || exit 1

echo "Universal DMG created successfully: TheBoys-Launcher-$VERSION-Universal.dmg"
```

### Phase 5: Testing and Polish (Week 9-10)

#### 5.1 Cross-Platform Testing
**Testing Areas:**
- ✅ GUI rendering and functionality
- ✅ Modpack installation and launching
- ✅ Java runtime management
- ✅ Update system functionality
- ✅ Settings persistence
- ✅ Log file management

#### 5.2 Performance Optimization
**Areas:**
- App bundle startup time
- Memory usage optimization
- File I/O performance
- Network request optimization

#### 5.3 Documentation Updates
**Files to Update:**
- `README.md` - Add macOS installation instructions
- `BUILD.md` - Add macOS build instructions
- User documentation - macOS-specific notes

## Technical Requirements

### Development Environment
- **Go 1.21+**: Latest stable Go release
- **Xcode Command Line Tools**: For macOS compilation
- **Create DMG tool**: For DMG creation (`brew install create-dmg`)
- **Optional**: macOS development machine (Intel and Apple Silicon for testing)

### Dependencies
- **Fyne v2.7+**: GUI framework (already cross-platform)
- **golang.org/x/sys**: System calls (already included)
- ** Adoptium Temurin JREs**: Java runtime (cross-platform)

### Build Targets
- **macOS AMD64**: Intel-based Macs (REQUIRED)
- **macOS ARM64**: Apple Silicon Macs (REQUIRED)
- **Universal Binary**: Combined Intel + ARM (recommended for distribution)

## Success Criteria

### Phase 1 Success
- [ ] Platform detection working correctly
- [ ] All Windows-specific code isolated to platform files
- [ ] Basic compilation on macOS without errors

### Phase 2 Success
- [ ] Memory detection working on macOS
- [ ] Java downloads working for macOS
- [ ] Prism Launcher integration functional on macOS

### Phase 3 Success
- [ ] Build scripts working on macOS for both Intel and Apple Silicon
- [ ] Dual architecture building working
- [ ] Resource embedding working

### Phase 4 Success
- [ ] App bundle created successfully for both architectures
- [ ] DMG installer created for both architectures
- [ ] Universal binary and DMG created
- [ ] Installation and launch working on all Mac variants

### Phase 5 Success
- [ ] All Windows features working on macOS
- [ ] Performance comparable to Windows version
- [ ] User acceptance testing passed

## Risk Mitigation

### Technical Risks
1. **Java Compatibility**: Adoptium provides excellent macOS support for both Intel and Apple Silicon
2. **Prism Launcher**: Official macOS builds available for both architectures
3. **Architecture Detection**: Need proper detection for Intel vs Apple Silicon
4. **Universal Binary Creation**: lipo tool dependency and correct usage

### Timeline Risks
1. **Dual Architecture Testing**: Need access to both Intel and Apple Silicon Macs
2. **Universal Binary Issues**: Complex debugging if universal binary fails
3. **Build Environment**: Cross-compilation setup complexity

### Contingency Plans
1. **Separate Distribution**: Distribute Intel and Apple Silicon versions separately if universal fails
2. **Alternative Install Methods**: Use tar.gz if DMG creation fails
3. **Manual Testing**: Extensive manual testing on both architectures
4. **Fallback to Single Architecture**: Focus on one architecture if dual proves problematic

## Post-Implementation (Phase 6)

### Universal Binary Optimization
- Verify universal binary works correctly on both architectures
- Optimize performance for Apple Silicon
- Update distribution strategy for universal binaries

### Linux Support (Optional)
- Apply platform abstractions for Linux
- Create Linux packaging (AppImage, DEB, RPM)
- Test with various Linux distributions

### Maintenance and Updates
- Establish CI/CD for dual architecture builds
- Create testing matrix for all supported platforms
- Document cross-platform build procedures

## Implementation Notes

### No Placeholders - Real Implementation Only
This plan emphasizes that all implementations must be complete and functional. No placeholder functions, TODO comments, or mock implementations should be committed to the main branch.

### Testing Requirements
Each phase must include comprehensive testing before proceeding to the next phase. All Windows features must work identically on macOS.

### Documentation Requirements
All new platform-specific code must be thoroughly documented. Build instructions must be clear and repeatable.

### Code Quality Standards
- Follow Go best practices
- Maintain existing code style
- Add comprehensive error handling
- Include platform-specific unit tests

## Conclusion

This 8-week implementation plan will convert TheBoys Launcher from Windows-only to full macOS support while maintaining 100% feature parity for both Intel and Apple Silicon Macs. The phased approach ensures manageable development cycles with clear success criteria and risk mitigation strategies.

The project is technically feasible with the current codebase architecture, and the Go ecosystem provides excellent cross-platform support. The main challenges are platform-specific integrations (dual architecture support, app bundles) rather than core functionality changes.

**Simplified Timeline Benefits:**
- **No code signing requirements** reduces complexity by 1-2 weeks
- **No App Store distribution** eliminates sandboxing and review processes
- **Focus on dual architecture support** ensures compatibility with all Macs
- **Direct distribution** via DMG files provides immediate user access

The plan prioritizes robust dual-architecture support and proper macOS integration while maintaining the simplicity of direct distribution without Apple's developer program requirements.