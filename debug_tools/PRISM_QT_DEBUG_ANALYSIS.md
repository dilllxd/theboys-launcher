# Prism Launcher Qt Debugging Analysis

## Problem Summary

The environment variable solution implemented for Prism Launcher Qt library issues is not working. Users are still getting the same error messages about missing Qt libraries:

```
/home/server/.theboyslauncher/prism/plugins/iconengines/libqsvgicon.so:, 
/home/server/.theboyslauncher/prism/plugins/imageformats/libqgif.so:, ...
```

## Root Cause Analysis

Based on my investigation, I've identified **5-7 potential sources** of the problem, which I've distilled down to the **2 most likely causes**:

### Most Likely Causes

1. **RPATH Issues in Qt Plugins**: The Qt plugins (especially `libqxcb.so`) have incorrect RPATH settings, preventing them from finding the bundled Qt libraries even when environment variables are set correctly.

2. **Missing System Dependencies**: The portable Prism build may require certain system libraries (like `libxcb-cursor.so.0`) that aren't installed on the target system.

### Other Potential Causes

3. **Environment Variable Timing**: Environment variables might be set after Qt initialization has already begun.
4. **Incorrect Directory Structure**: The portable build structure might be different than expected.
5. **Plugin Loading Order**: Qt might be finding system plugins before bundled ones.
6. **Permissions Issues**: The plugins or libraries might not have proper execute permissions.
7. **Version Mismatch**: Qt version mismatch between plugins and libraries.

## Debugging Tools Created

I've created two debugging tools to help validate the issue:

### 1. Go Debug Tool
**File**: `debug_tools/prism_qt_debug.go`
**Usage**: `go run debug_tools/prism_qt_debug.go <prism_directory>`

### 2. Bash Debug Script
**File**: `debug_tools/prism_qt_debug.sh`
**Usage**: `chmod +x debug_tools/prism_qt_debug.sh && ./debug_tools/prism_qt_debug.sh <prism_directory>`

Both tools will:
- Verify environment variable setup
- Check if Qt plugin files exist at expected paths
- Analyze library directory structure
- Check Prism executable permissions and dependencies
- Look for wrapper scripts
- Test environment variable effectiveness

## Recommended Solutions

### Primary Solution: Fix RPATH Issues

The most likely issue is that Qt plugins have incorrect RPATH settings. This is a common problem with Qt portable applications.

**Implementation**:
```bash
# Fix RPATH for Qt platform plugins
patchelf --set-rpath '$ORIGIN/../../lib' plugins/platforms/libqxcb.so
patchelf --set-rpath '$ORIGIN/../../lib' plugins/imageformats/libqjpeg.so
patchelf --set-rpath '$ORIGIN/../../lib' plugins/iconengines/libqsvgicon.so
```

**Code Implementation**:
```go
func fixQtPluginRPATH(prismDir string) error {
    pluginsDir := filepath.Join(prismDir, "plugins")
    
    // Fix RPATH for critical plugins
    criticalPlugins := []string{
        "platforms/libqxcb.so",
        "imageformats/libqjpeg.so", 
        "iconengines/libqsvgicon.so",
    }
    
    for _, plugin := range criticalPlugins {
        pluginPath := filepath.Join(pluginsDir, plugin)
        if exists(pluginPath) {
            cmd := exec.Command("patchelf", "--set-rpath", "$ORIGIN/../../lib", pluginPath)
            if err := cmd.Run(); err != nil {
                logf("Failed to fix RPATH for %s: %v", plugin, err)
                return err
            }
            logf("Fixed RPATH for %s", plugin)
        }
    }
    
    return nil
}
```

### Secondary Solution: Wrapper Script Approach

Create a wrapper script that ensures proper environment setup before launching Prism:

```bash
#!/bin/bash
# File: prism-launcher-wrapper.sh
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRISM_DIR="$SCRIPT_DIR"

# Set Qt environment variables
export QT_PLUGIN_PATH="$PRISM_DIR/plugins"
export LD_LIBRARY_PATH="$PRISM_DIR/lib:$LD_LIBRARY_PATH"
export QT_DEBUG_PLUGINS=1  # Enable Qt plugin debugging
export QT_LOGGING_RULES="*=true"

# Fix RPATH if needed
if command -v patchelf >/dev/null 2>&1; then
    find "$PRISM_DIR/plugins" -name "*.so" -exec patchelf --set-rpath '$ORIGIN/../../lib' {} \;
fi

# Launch Prism with provided arguments
exec "$PRISM_DIR/PrismLauncher" "$@"
```

### Tertiary Solution: System Dependency Installation

Check for and install missing system dependencies:

```go
func installSystemDependencies() error {
    // Check for missing system libraries
    missingLibs := []string{
        "libxcb-cursor.so.0",
        "libxcb-xinerama.so.0", 
        "libxcb-util.so.1",
    }
    
    var needsInstall []string
    for _, lib := range missingLibs {
        if !checkLibraryExists(lib) {
            needsInstall = append(needsInstall, lib)
        }
    }
    
    if len(needsInstall) > 0 {
        logf("Installing missing system dependencies...")
        // Install appropriate packages based on distribution
        if exists("/usr/bin/apt") {
            return exec.Command("sudo", "apt", "install", "-y", 
                "libxcb-cursor0", "libxcb-xinerama0", "libxcb-util1").Run()
        } else if exists("/usr/bin/dnf") {
            return exec.Command("sudo", "dnf", "install", "-y",
                "libxcb-cursor", "libxcb-xinerama", "libxcb-util").Run()
        }
    }
    
    return nil
}
```

## Enhanced Environment Variable Implementation

The current implementation is mostly correct but needs these enhancements:

### 1. Add Qt Debug Variables
```go
func buildQtEnvironment(prismDir, jreDir string) []string {
    qtEnv := []string{
        "JAVA_HOME=" + jreDir,
        "PATH=" + BuildPathEnv(filepath.Join(jreDir, "bin")),
    }

    if runtime.GOOS == "linux" {
        qtPluginPath := filepath.Join(prismDir, "plugins")
        qtLibPath := filepath.Join(prismDir, "lib")
        
        if exists(qtPluginPath) {
            qtEnv = append(qtEnv, "QT_PLUGIN_PATH="+qtPluginPath)
        }
        
        if exists(qtLibPath) {
            existingLdPath := os.Getenv("LD_LIBRARY_PATH")
            if existingLdPath != "" {
                qtEnv = append(qtEnv, "LD_LIBRARY_PATH="+qtLibPath+":"+existingLdPath)
            } else {
                qtEnv = append(qtEnv, "LD_LIBRARY_PATH="+qtLibPath)
            }
        }
        
        // Add debugging variables
        qtEnv = append(qtEnv, "QT_DEBUG_PLUGINS=1")
        qtEnv = append(qtEnv, "QT_LOGGING_RULES*=true")
        qtEnv = append(qtEnv, "QT_QPA_PLATFORM=xcb")
        qtEnv = append(qtEnv, "QT_XCB_GL_INTEGRATION=xcb_glx")
    }

    return qtEnv
}
```

### 2. Pre-launch Validation
```go
func validateQtEnvironment(prismDir string) error {
    // Check critical plugins exist
    criticalPlugins := []string{
        "plugins/platforms/libqxcb.so",
        "plugins/imageformats/libqjpeg.so",
        "plugins/iconengines/libqsvgicon.so",
    }
    
    for _, plugin := range criticalPlugins {
        pluginPath := filepath.Join(prismDir, plugin)
        if !exists(pluginPath) {
            return fmt.Errorf("critical Qt plugin missing: %s", plugin)
        }
    }
    
    // Check library directory
    libDir := filepath.Join(prismDir, "lib")
    if !exists(libDir) {
        return fmt.Errorf("Qt library directory missing: %s", libDir)
    }
    
    return nil
}
```

## Testing Strategy

### 1. Use Debug Tools
Run the debugging tools to identify specific issues:
```bash
# On the target system
./debug_tools/prism_qt_debug.sh /home/server/.theboyslauncher/prism
```

### 2. Test RPATH Fix
```bash
# Test RPATH fix manually
cd /home/server/.theboyslauncher/prism
patchelf --set-rpath '$ORIGIN/../../lib' plugins/platforms/libqxcb.so
./PrismLauncher
```

### 3. Test Environment Variables
```bash
# Test with manual environment setup
cd /home/server/.theboyslauncher/prism
export QT_PLUGIN_PATH="$(pwd)/plugins"
export LD_LIBRARY_PATH="$(pwd)/lib:$LD_LIBRARY_PATH"
export QT_DEBUG_PLUGINS=1
./PrismLauncher
```

## Implementation Priority

1. **Immediate**: Implement RPATH fixing in the launcher
2. **Short-term**: Add enhanced debugging and validation
3. **Medium-term**: Implement wrapper script fallback
4. **Long-term**: Consider AppImage distribution for better portability

## Conclusion

The Qt library issue is most likely caused by incorrect RPATH settings in the Qt plugins, preventing them from finding the bundled libraries even when environment variables are set correctly. The solution involves fixing the RPATH settings using `patchelf` and enhancing the environment variable setup with debugging capabilities.

The debugging tools I've created will help validate this diagnosis and identify any additional issues that might be present.