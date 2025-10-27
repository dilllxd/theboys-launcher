# Qt Library Bundling Solutions Analysis for Prism Launcher on Linux

## Problem Analysis

### Current Issue
TheBoysLauncher is experiencing Qt library bundling issues when launching Prism Launcher on Linux (specifically Ubuntu 24.04). The error message indicates missing Qt plugins and system libraries:

```
Error: The launcher is missing the following libraries that it needs to work correctly:
- /home/server/.theboyslauncher/prism/plugins/iconengines/libqsvgicon.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqgif.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqicns.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqico.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqjpeg.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqsvg.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqwbmp.so
- /home/server/.theboyslauncher/prism/plugins/imageformats/libqwebp.so
- /home/server/.theboyslauncher/prism/plugins/platforms/libqeglfs.so
- /home/server/.theboyslauncher/prism/plugins/platforms/libqvkkhrdisplay.so
- /home/server/.theboyslauncher/prism/plugins/platforms/libqvnc.so
- /home/server/.theboyslauncher/prism/plugins/platforms/libqwayland-egl.so
- /home/server/.theboyslauncher/prism/plugins/platforms/libqwayland-generic.so
- /home/server/.theboyslauncher/prism/plugins/platforms/libqxcb.so
- /home/server/.theboyslauncher/prism/plugins/tls/libqopensslbackend.so
- /home/server/.theboyslauncher/prism/plugins/wayland-decoration-client/libbradient.so
- /home/server/.theboyslauncher/prism/plugins/wayland-graphics-integration-client/libdmabuf-server.so
- /home/server/.theboyslauncher/prism/plugins/wayland-graphics-integration-client/libdrm-egl-server.so
- /home/server/.theboyslauncher/prism/plugins/wayland-graphics-integration-client/libqt-plugin-wayland-egl.so
- /home/server/.theboyslauncher/prism/plugins/wayland-graphics-integration-client/libshm-emulation-server.so
- /home/server/.theboyslauncher/prism/plugins/wayland-graphics-integration-client/libvulkan-server.so
- /home/server/.theboyslauncher/prism/plugins/wayland-shell-integration/libfullscreen-shell-v1.so
- /home/server/.theboyslauncher/prism/plugins/wayland-shell-integration/libivi-shell.so
- /home/server/.theboyslauncher/prism/plugins/wayland-shell-integration/libqt-shell.so
- /home/server/.theboyslauncher/prism/plugins/wayland-shell-integration/libwl-shell-plugin.so
- /home/server/.theboyslauncher/prism/plugins/wayland-shell-integration/libxdg-shell.so
- libxcb-cursor.so.0
```

### Root Cause
The issue occurs because Prism Launcher's Qt runtime cannot locate its bundled plugins and system libraries. This happens when:
1. `QT_PLUGIN_PATH` is not set to point to the bundled plugins directory
2. `LD_LIBRARY_PATH` is not set to include the bundled Qt libraries
3. The launcher doesn't properly configure the environment before starting Prism

### Current Implementation Analysis
From the codebase analysis, the launcher currently sets:
- `JAVA_HOME` and `PATH` for Java runtime
- Basic process attributes for GUI execution
- Working directory to the prism directory

However, it doesn't set Qt-specific environment variables needed for proper Qt runtime initialization.

---

## Solution 1: Environment Variable Solution

### Description
Modify the launcher to set proper Qt environment variables before launching Prism Launcher.

### Implementation Details
```go
// In launcher.go, modify the Prism launch section
launch := exec.Command(prismExe, "--dir", ".", "--launch", modpack.InstanceName)
launch.Dir = prismDir

// Set Qt environment variables
qtEnv := []string{
    "JAVA_HOME=" + jreDir,
    "PATH=" + BuildPathEnv(filepath.Join(jreDir, "bin")),
    "QT_PLUGIN_PATH=" + filepath.Join(prismDir, "plugins"),
    "LD_LIBRARY_PATH=" + BuildPathEnv(filepath.Join(prismDir, "lib")),
}

// Merge with existing environment
launch.Env = append(os.Environ(), qtEnv...)
```

### Implementation Complexity: **Low**
- Requires minimal code changes
- Only needs modification to the Prism launch section
- No changes to build process

### Maintenance Overhead: **Low**
- Simple to maintain
- No additional dependencies
- Clear and straightforward implementation

### User Experience Impact: **Positive**
- Transparent to users
- No additional steps required
- Works with existing distribution model

### Compatibility: **High**
- Works across all Linux distributions
- Compatible with both Qt5 and Qt6
- Doesn't interfere with system Qt installations

### Build Process Changes: **None**
- No changes to Makefile or build scripts
- No additional tools required

### Pros
- Simple implementation
- Minimal code changes
- Maintains current distribution model
- Works with existing portable Prism builds
- No user-side configuration required

### Cons
- Relies on Prism's portable build having proper library structure
- May not work if Prism's directory structure changes
- Doesn't solve system library dependencies (like libxcb-cursor.so.0)

---

## Solution 2: Wrapper Script Solution

### Description
Create a wrapper script that sets up the proper environment before launching Prism Launcher.

### Implementation Details
Create a wrapper script `launch-prism.sh`:
```bash
#!/bin/bash
set -e

# Get directory where script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRISM_DIR="$SCRIPT_DIR"

# Set Qt environment variables
export QT_PLUGIN_PATH="$PRISM_DIR/plugins"
export LD_LIBRARY_PATH="$PRISM_DIR/lib:$LD_LIBRARY_PATH"
export JAVA_HOME="$PRISM_DIR/java/jre*"
export PATH="$JAVA_HOME/bin:$PATH"

# Launch Prism with provided arguments
exec "$PRISM_DIR/PrismLauncher" "$@"
```

Modify the launcher to use the wrapper:
```go
// Create wrapper script
wrapperPath := filepath.Join(prismDir, "launch-prism.sh")
wrapperContent := fmt.Sprintf(`#!/bin/bash
set -e
PRISM_DIR="%s"
export QT_PLUGIN_PATH="$PRISM_DIR/plugins"
export LD_LIBRARY_PATH="$PRISM_DIR/lib:$LD_LIBRARY_PATH"
export JAVA_HOME="%s"
export PATH="$JAVA_HOME/bin:$PATH"
exec "$PRISM_DIR/PrismLauncher" "$@"`, prismDir, jreDir)

if err := os.WriteFile(wrapperPath, []byte(wrapperContent), 0755); err != nil {
    return err
}

// Launch using wrapper
launch := exec.Command(wrapperPath, "--dir", ".", "--launch", modpack.InstanceName)
```

### Implementation Complexity: **Medium**
- Requires creating and managing wrapper script
- Need to handle script permissions
- More complex than direct environment variable setting

### Maintenance Overhead: **Medium**
- Additional script to maintain
- Need to ensure script stays in sync with launcher logic
- Script debugging can be more challenging

### User Experience Impact: **Neutral**
- Transparent to users
- Slight overhead from script execution
- May create additional debugging complexity

### Compatibility: **High**
- Works across all Linux distributions
- Compatible with different shell environments
- Can be easily modified for different scenarios

### Build Process Changes: **Minimal**
- No changes to compilation
- May need to include wrapper script in distribution

### Pros
- Separates environment setup from launcher code
- Easy to debug and modify environment
- Can be extended with additional checks
- Provides clear separation of concerns

### Cons
- Additional file to manage
- Script execution overhead
- More complex error handling
- Potential permission issues

---

## Solution 3: AppImage Solution

### Description
Convert the launcher to an AppImage that bundles all dependencies including Qt libraries.

### Implementation Details
1. Create an AppImage structure:
```
TheBoysLauncher.AppDir/
├── TheBoysLauncher          # Main executable
├── prism/                   # Prism Launcher with bundled Qt
├── java/                    # Bundled Java runtime
├── lib/                     # Additional system libraries
├── usr/
│   ├── bin/
│   ├── lib/
│   └── share/
├── AppRun                   # Launcher script
└── TheBoysLauncher.desktop   # Desktop entry
```

2. Create AppRun script:
```bash
#!/bin/bash
HERE="$(dirname "$(readlink -f "${0}")")"
export LD_LIBRARY_PATH="${HERE}/usr/lib:${HERE}/lib:${LD_LIBRARY_PATH}"
export QT_PLUGIN_PATH="${HERE}/usr/plugins:${HERE}/prism/plugins"
export PATH="${HERE}/usr/bin:${PATH}"
exec "${HERE}/TheBoysLauncher" "$@"
```

3. Modify build process to create AppImage:
```makefile
build-appimage: build-linux
	@echo "Creating AppImage..."
	@mkdir -p build/appimage/TheBoysLauncher.AppDir
	@cp build/linux/TheBoysLauncher-linux build/appimage/TheBoysLauncher.AppDir/TheBoysLauncher
	@# Copy dependencies and create AppImage structure
	@appimagetool build/appimage/TheBoysLauncher.AppDir
```

### Implementation Complexity: **High**
- Requires significant restructuring
- Need to learn AppImage creation process
- Complex dependency management

### Maintenance Overhead: **High**
- AppImage build process to maintain
- Dependency updates require AppImage rebuilds
- More complex release process

### User Experience Impact: **Mixed**
- Single file distribution is convenient
- No installation required
- Larger download size
- May have integration issues with system

### Compatibility: **Very High**
- Works across most Linux distributions
- Self-contained with no external dependencies
- Consistent behavior across systems

### Build Process Changes: **Major**
- Requires AppImage tools and dependencies
- Significant changes to build pipeline
- New packaging and distribution process

### Pros
- Truly portable across Linux distributions
- Bundles all dependencies
- No installation required
- Consistent runtime environment
- Professional distribution format

### Cons
- Complex implementation
- Large file size
- Requires AppImage toolchain
- More complex build process
- Potential system integration issues

---

## Solution 4: Static Linking Solution

### Description
Investigate using statically linked Qt libraries for Prism Launcher.

### Implementation Details
This solution would require:
1. Building or obtaining Prism Launcher with statically linked Qt libraries
2. Modifying the download logic to prefer static builds
3. Updating the fetchLatestPrismPortableURL function:

```go
// In prism.go, modify Linux patterns
} else {
    // Linux: prioritize static builds for better compatibility
    // Priority 1: Static Qt6 build
    patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-Qt6-Static-%s.tar.gz", latestTag))
    // Priority 2: Static Qt5 build
    patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-Qt5-Static-%s.tar.gz", latestTag))
    // Fallback to portable builds
    patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-Qt6-Portable-%s.tar.gz", latestTag))
    patterns = append(patterns, fmt.Sprintf("PrismLauncher-Linux-Qt5-Portable-%s.tar.gz", latestTag))
}
```

### Implementation Complexity: **Very High**
- Requires custom Qt builds
- Complex dependency management
- May not be feasible with current Prism builds

### Maintenance Overhead: **Very High**
- Need to maintain custom Qt builds
- Complex update process
- Potential licensing issues

### User Experience Impact: **Excellent**
- No dependency issues
- Single executable
- Fast startup
- No environment setup required

### Compatibility: **Very High**
- Works across all Linux distributions
- No external dependencies
- Consistent behavior

### Build Process Changes: **Major**
- Requires Qt source and build environment
- Complex build configuration
- Significant build time increase

### Pros
- Ultimate compatibility
- No runtime dependencies
- Simple distribution
- Excellent user experience

### Cons
- Very complex implementation
- Large executable size
- Potential licensing issues
- Requires custom Prism builds
- May not be feasible

---

## Solution 5: System Dependency Solution

### Description
Automatically install required Qt packages via the system package manager.

### Implementation Details
Create a dependency installation function:
```go
func installQtDependencies() error {
    // Detect package manager
    var cmd *exec.Cmd
    if exists("/usr/bin/apt") {
        // Ubuntu/Debian
        cmd = exec.Command("sudo", "apt", "install", "-y", 
            "libqt6gui6", "libqt6widgets6", "libqt6network6",
            "libqt6svg6", "libqt6x11extras6", "libxcb-cursor0")
    } else if exists("/usr/bin/dnf") {
        // Fedora/RHEL
        cmd = exec.Command("sudo", "dnf", "install", "-y",
            "qt6-qtbase-gui", "qt6-qtbase-widgets", "qt6-qtsvg")
    } else if exists("/usr/bin/pacman") {
        // Arch Linux
        cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm",
            "qt6-base", "qt6-svg")
    } else {
        return fmt.Errorf("unsupported package manager")
    }
    
    return cmd.Run()
}
```

Modify the launcher to check and install dependencies:
```go
// In launcher.go, before launching Prism
if err := checkQtDependencies(); err != nil {
    logf("%s", warnLine("Qt dependencies missing, attempting to install..."))
    if err := installQtDependencies(); err != nil {
        return fmt.Errorf("failed to install Qt dependencies: %w", err)
    }
}
```

### Implementation Complexity: **Medium**
- Requires package manager detection
- Need to handle different package names
- User interaction for sudo access

### Maintenance Overhead: **High**
- Need to maintain package lists for different distributions
- Package names may change between versions
- Requires testing on multiple distributions

### User Experience Impact: **Mixed**
- Automatic installation is convenient
- Requires sudo access
- May take time to install packages
- Potential for installation failures

### Compatibility: **Medium**
- Works on major distributions
- May not work on all distributions
- Package availability varies

### Build Process Changes: **None**
- No changes to build process
- Only runtime changes

### Pros
- Uses system packages
- Proper integration with system
- Smaller launcher download
- Keeps system updated

### Cons
- Requires sudo access
- Distribution-specific maintenance
- Package name variations
- Installation time
- Potential conflicts with system packages

---

## Comparison Matrix

| Solution | Implementation Complexity | Maintenance Overhead | User Experience | Compatibility | Build Changes |
|----------|-------------------------|---------------------|-----------------|----------------|---------------|
| Environment Variables | Low | Low | Positive | High | None |
| Wrapper Script | Medium | Medium | Neutral | High | Minimal |
| AppImage | High | High | Mixed | Very High | Major |
| Static Linking | Very High | Very High | Excellent | Very High | Major |
| System Dependencies | Medium | High | Mixed | Medium | None |

---

## Recommendations

### Primary Recommendation: Environment Variable Solution

**Reasoning:**
1. **Simplicity**: Minimal code changes with maximum impact
2. **Reliability**: Directly addresses the root cause of the issue
3. **Maintainability**: Easy to understand and modify
4. **Compatibility**: Works across all Linux distributions
5. **User Experience**: Transparent to users with no additional steps

**Implementation Priority: High**
- This should be implemented immediately as it solves the core issue with minimal risk

### Secondary Recommendation: Wrapper Script Solution

**Reasoning:**
1. **Flexibility**: Easy to extend with additional environment setup
2. **Debugging**: Clear separation of environment setup from launcher logic
3. **Maintainability**: Environment logic can be modified without recompiling

**Implementation Priority: Medium**
- Consider this if the environment variable solution proves insufficient

### Tertiary Recommendation: AppImage Solution

**Reasoning:**
1. **Portability**: Ultimate cross-distribution compatibility
2. **Professional**: Industry-standard approach for Linux application distribution
3. **Self-contained**: No external dependencies

**Implementation Priority: Low**
- Consider for long-term distribution strategy if current approach proves insufficient

### Not Recommended: Static Linking and System Dependencies

**Static Linking:**
- Too complex to implement
- Requires custom Prism builds
- Potential licensing issues

**System Dependencies:**
- Requires sudo access
- High maintenance overhead
- Poor user experience

---

## Implementation Plan

### Phase 1: Environment Variable Solution (Immediate)
1. Modify [`launcher.go`](../launcher.go) to set Qt environment variables
2. Test with Ubuntu 24.04
3. Verify all Qt plugins are found
4. Test with other distributions

### Phase 2: Enhanced Error Handling (Short-term)
1. Add detection for missing Qt libraries
2. Provide clear error messages
3. Fall back to wrapper script if needed

### Phase 3: Wrapper Script Enhancement (Medium-term)
1. Implement wrapper script solution
2. Add comprehensive environment setup
3. Include system dependency detection

### Phase 4: AppImage Migration (Long-term)
1. Evaluate AppImage creation process
2. Plan migration strategy
3. Implement AppImage build pipeline

---

## Testing Strategy

### Unit Tests
- Test environment variable setting
- Verify path construction
- Test plugin path resolution

### Integration Tests
- Test Prism launch with Qt environment
- Verify plugin loading
- Test on multiple distributions

### User Acceptance Tests
- Test on clean systems
- Verify error handling
- Performance testing

---

## Conclusion

The Qt library bundling issue with Prism Launcher on Linux can be effectively solved using the Environment Variable Solution with minimal implementation complexity and maintenance overhead. This approach directly addresses the root cause while maintaining compatibility across all Linux distributions and preserving the current distribution model.

For long-term distribution strategy, consider migrating to AppImage format, but this should be approached as a separate initiative rather than a solution to the immediate problem.