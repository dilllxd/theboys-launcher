# Platform-Specific Changes Required for macOS Support

## Current Windows-Only Code Analysis

### 1. Hard Windows Checks
**Location: `main.go:55-57`**
```go
if runtime.GOOS != "windows" {
    fail(errors.New("Windows only"))
}
```
**Change Required:** Remove or replace with platform support check

### 2. Windows-Specific Imports
**Locations:**
- `config.go:10` - `"golang.org/x/sys/windows"`
- `launcher.go:15` - `"golang.org/x/sys/windows"`
- `multimc.go:11` - `"golang.org/x/sys/windows"`
- `update.go:15` - `"golang.org/x/sys/windows"`

**Changes Required:** Create platform-specific implementations

### 3. Windows API Calls
**Location: `config.go:183-201`**
```go
// Use Windows GlobalMemoryStatusEx API to get total physical memory
// Define the structure ourselves since it's not in the basic windows package
type MEMORYSTATUSEX struct {
    // ... Windows-specific struct
}
kernel32 := windows.NewLazyDLL("kernel32.dll")
```
**Change Required:** Implement macOS equivalent using `sysctlbyname`

### 4. Windows Process Management
**Location: `main.go:101-112`**
```go
cmd := exec.Command("taskkill", "/F", "/IM", "PrismLauncher.exe")
javaCmd := exec.Command("taskkill", "/F", "/IM", "java.exe")
killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", prismProcess.Pid))
```
**Change Required:** Replace with Unix `pkill`/`kill` commands

### 5. Windows Console Hiding
**Location: `console_windows.go` (entire file)**
```go
func hideConsoleWindow() {
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    freeConsole := kernel32.NewProc("FreeConsole")
    // ... Windows-specific console hiding
}
```
**Change Required:** No-op on macOS (GUI apps don't show console)

### 6. Windows Executable Names and Paths
**Locations throughout codebase:**
- `"TheBoysLauncher.exe"`
- `"PrismLauncher.exe"`
- `"java.exe"`
- `"javaw.exe"`
- `"packwiz-installer-bootstrap.exe"`

**Changes Required:** Platform-specific executable names

### 7. Windows Process Creation Flags
**Locations:**
- `launcher.go:254-258`
- `multimc.go:265-268`
- `update.go:93-97`

```go
cmd.SysProcAttr = &windows.SysProcAttr{
    HideWindow:    true,
    CreationFlags: windows.CREATE_NO_WINDOW,
}
```
**Change Required:** Platform-specific process attributes

### 8. Windows-Specific Java Download URLs
**Location: `java.go:156,221`**
```go
adoptium := fmt.Sprintf("https://api.adoptium.net/v3/assets/latest/%s/hotspot?architecture=x64&image_type=%s&os=windows", javaVersion, imageType)
assetName := fmt.Sprintf("OpenJDK%sU-%s_x64_windows_hotspot_%s.zip", javaVersion, imageType, tagWithoutHyphens)
```
**Change Required:** Dynamic OS parameter in URLs

### 9. Windows Prism Launcher Downloads
**Location: `prism.go:129-141`**
```go
patterns = append(patterns, fmt.Sprintf("PrismLauncher-Windows-MinGW-w64-Portable-%s.zip", latestTag))
patterns = append(patterns, fmt.Sprintf("PrismLauncher-Windows-MSVC-Portable-%s.zip", latestTag))
```
**Change Required:** Add macOS download patterns

### 10. Windows Build Scripts
**Files:**
- `build.ps1` - PowerShell build script
- `build.bat` - Batch build script
- `Makefile` - Windows-only make targets
- `TheBoysLauncher.iss` - Inno Setup installer script

**Changes Required:** Create macOS build equivalents

### 11. Windows Resource Compilation
**Location: `main.go:1`**
```go
//go:generate goversioninfo -64 -icon=icon.ico
```
**Changes Required:** Platform-specific resource embedding

### 12. Windows User Environment
**Location: `gui.go:371`**
```go
widget.NewLabel(fmt.Sprintf("Signed in as: %s", os.Getenv("USERNAME")))
```
**Change Required:** Use `$USER` on macOS instead of `$USERNAME`

## Required New Files

### Platform Abstraction Files
1. **`platform.go`** - Common platform interface
2. **`platform_windows.go`** - Windows implementations (refactor existing)
3. **`platform_darwin.go`** - macOS implementations
4. **`process_windows.go`** - Windows process management
5. **`process_darwin.go`** - macOS process management
6. **`memory_windows.go`** - Windows memory detection
7. **`memory_darwin.go`** - macOS memory detection
8. **`console_darwin.go`** - macOS console handling (empty)

### Build System Files
1. **`build.sh`** - Cross-platform build script
2. **`build-macos.sh`** - macOS-specific build
3. **`scripts/create-app-bundle.sh`** - App bundle creation
4. **`scripts/create-dmg.sh`** - DMG creation
5. **`scripts/sign-macos.sh`** - macOS code signing
6. **`build/Info.plist`** - macOS app metadata
7. **`build/entitlements.plist`** - macOS app entitlements

### Resource Files
1. **`resources/darwin/icon.icns`** - macOS icon format
2. **`resources/common/`** - Cross-platform resources

## File-by-File Changes Required

### `main.go`
- Remove hardcoded Windows check
- Update executable name handling
- Fix user environment variable

### `config.go`
- Replace Windows memory detection
- Add platform-specific constants
- Remove Windows-specific imports

### `launcher.go`
- Update Java binary paths
- Replace process creation flags
- Add platform-specific executable handling

### `java.go`
- Add OS parameter to Adoptium API calls
- Update asset naming patterns
- Add macOS Java installation logic

### `prism.go`
- Add macOS download patterns
- Update Prism executable detection
- Handle macOS app bundle structure

### `utils.go`
- Add platform-specific directory handling
- Update file permissions for macOS
- Add macOS-specific utilities

### `gui.go`
- Fix user environment variable
- Add platform-specific window handling
- Update status displays for macOS

### `update.go`
- Replace Windows process attributes
- Add macOS update mechanisms
- Handle macOS file locking

### `multimc.go`
- Replace Windows process creation
- Add platform-specific installer handling
- Update command-line arguments

### `console_windows.go` → `console_darwin.go`
- Create empty implementation for macOS
- Keep Windows implementation in separate file

## Build System Changes

### Makefile Updates
```makefile
# New targets needed
.PHONY: build-macos build-macos-arm64 package-macos clean-macos

build-macos:
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(VERSION)" -o TheBoysLauncher .

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build -ldflags="-s -w -X main.version=$(VERSION)" -o TheBoysLauncher-arm64 .

package-macos: build-macos
	./scripts/create-app-bundle.sh
	./scripts/create-dmg.sh
```

### New Build Scripts
1. **Cross-platform shell script** replacing PowerShell/Batch
2. **App bundle creation** for proper macOS integration
3. **Code signing** for distribution
4. **DMG creation** for user-friendly installation

## Testing Requirements

### Unit Tests
- Memory detection on macOS
- Process management functions
- Path handling functions
- Platform detection

### Integration Tests
- Java download and installation
- Prism Launcher integration
- Modpack installation and launch
- Update mechanism
- GUI functionality

### Manual Testing
- Complete user workflows
- Performance comparison
- Error handling
- Edge cases

## Dependencies Changes

### Go Modules
- Add `golang.org/x/sys/unix` for macOS system calls
- Keep existing `golang.org/x/sys/windows` for Windows support
- No major GUI changes needed (Fyne is cross-platform)

### External Dependencies
- **Java Runtime**: Adoptium supports macOS
- **Prism Launcher**: Official macOS builds available
- **Build Tools**: Xcode Command Line Tools required

## Security Considerations

### Code Signing
- Developer ID required for distribution
- Notarization required for Gatekeeper compliance
- Proper entitlements for app functionality

### File Permissions
- macOS permission model differences
- App sandboxing considerations
- Library directory access

## Performance Considerations

### macOS Optimizations
- Grand Central Dispatch integration (optional)
- Metal API considerations (future)
- Apple Silicon optimization
- Memory management differences

### Build Optimizations
- Stripped binaries for smaller size
- Proper linking flags
- Architecture-specific optimizations

## Migration Strategy

### Development Workflow
1. Create feature branch for macOS support
2. Implement platform abstractions first
3. Add macOS-specific implementations
4. Test extensively on macOS
5. Maintain Windows compatibility throughout

### Testing Workflow
1. Set up macOS development environment
2. Create automated build pipeline
3. Test both platforms in CI/CD
4. User acceptance testing
5. Performance benchmarking

### Release Strategy
1. Beta release for macOS users
2. Gather feedback and fix issues
3. Full release with documentation
4. Ongoing maintenance for both platforms

This comprehensive analysis provides the foundation for implementing full macOS support while maintaining all existing Windows functionality.