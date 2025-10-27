# Prism Launcher Portable Build Analysis - Qt Library Issues

## Executive Summary

After extensive investigation of the Prism Launcher portable build structure and Qt library loading issues, I've identified **5-7 potential root causes** and distilled them down to the **2 most likely sources** of the problem. The issue persists despite RPATH fixes, suggesting fundamental structural problems with how the portable builds are being handled.

## Most Likely Root Causes

### 1. **Incorrect Directory Structure Assumptions**
**Probability: HIGH**

The current code assumes the Prism Launcher portable builds extract to a flat structure with:
```
prism/
├── PrismLauncher (executable)
├── lib/
├── plugins/
└── java/
```

However, Prism Launcher portable builds may actually extract to:
```
prism/
└── PrismLauncher-9.4/
    ├── PrismLauncher (executable)
    ├── lib/
    ├── plugins/
    └── java/
```

**Evidence:**
- The same error paths persist after RPATH fixes
- The error messages show the exact same paths, suggesting the files don't exist at expected locations
- Portable builds often use versioned subdirectories

### 2. **Qt6 vs Qt5 Portable Build Mismatch**
**Probability: HIGH**

The code prioritizes Qt6 portable builds but may be downloading Qt5 builds, or vice versa. Qt5 and Qt6 have different:
- Library naming conventions (`libQt5Core.so.5` vs `libQt6Core.so.6`)
- Plugin structures
- RPATH requirements

**Evidence:**
- The RPATH fix targets Qt6-style paths but the actual build might be Qt5
- Environment variables may not match the actual Qt version in the portable build

## Other Potential Causes

### 3. **Missing Wrapper Script Usage**
**Probability: MEDIUM**
Prism Launcher portable builds may include wrapper scripts that handle Qt environment setup automatically. The current implementation bypasses these scripts.

### 4. **Incomplete Qt Plugin Dependencies**
**Probability: MEDIUM**
The portable builds may not include all necessary Qt plugins or system dependencies.

### 5. **Extraction Process Issues**
**Probability: LOW**
The tar.gz extraction might not preserve permissions or symlinks correctly.

## Research Findings

### Prism Launcher Portable Build Structure

Based on GitHub releases analysis:
- **Qt6 Portable**: `PrismLauncher-Linux-Qt6-Portable-9.4.tar.gz`
- **Qt5 Portable**: `PrismLauncher-Linux-Qt5-Portable-9.4.tar.gz`
- **AppImage**: `PrismLauncher-Linux-x86_64.AppImage`

### Expected vs Actual Structure

**Current Code Assumption:**
```go
// In prism.go
prismExe := GetPrismExecutablePath(prismDir) // Assumes prismDir/PrismLauncher
qtPluginPath := filepath.Join(prismDir, "plugins") // Assumes prismDir/plugins
qtLibPath := filepath.Join(prismDir, "lib") // Assumes prismDir/lib
```

**Likely Actual Structure:**
```
prism/
└── PrismLauncher-9.4/  ← Versioned subdirectory
    ├── PrismLauncher
    ├── lib/
    ├── plugins/
    └── java/
```

## Diagnostic Tools Created

### 1. Portable Structure Debug Tool
**File**: `debug_tools/portable_structure_debug/prism_portable_structure_debug.go`
**Usage**: `go run debug_tools/portable_structure_debug/prism_portable_structure_debug.go <directory_or_url>`

**Capabilities:**
- Downloads and analyzes Prism Launcher portable builds
- Detects nested directory structures
- Validates Qt library and plugin presence
- Identifies missing critical components
- Provides specific recommendations

### 2. Enhanced Qt Debug Tools
**Files**: 
- `debug_tools/prism_qt_debug.go`
- `debug_tools/prism_qt_debug.sh`

**Capabilities:**
- Environment variable validation
- Qt plugin accessibility testing
- RPATH verification
- System dependency checking

## Recommended Solutions

### Primary Solution: Fix Directory Structure Detection

```go
func GetPrismExecutablePath(dir string) string {
    // Check for direct executable first
    directPath := filepath.Join(dir, "PrismLauncher")
    if runtime.GOOS == "windows" {
        directPath += ".exe"
    }
    if exists(directPath) {
        return directPath
    }
    
    // Check for versioned subdirectory structure
    files, err := os.ReadDir(dir)
    if err != nil {
        return directPath // Fallback
    }
    
    for _, file := range files {
        if file.IsDir() && strings.Contains(file.Name(), "PrismLauncher") {
            nestedPath := filepath.Join(dir, file.Name(), "PrismLauncher")
            if runtime.GOOS == "windows" {
                nestedPath += ".exe"
            }
            if exists(nestedPath) {
                return nestedPath
            }
        }
    }
    
    return directPath // Fallback
}

func getPrismBaseDir(prismDir string) string {
    // Detect if we need to use a nested directory
    prismExe := GetPrismExecutablePath(prismDir)
    return filepath.Dir(prismExe)
}

func buildQtEnvironment(prismDir, jreDir string) []string {
    // Use the actual base directory where Prism executable is located
    actualPrismDir := getPrismBaseDir(prismDir)
    
    qtPluginPath := filepath.Join(actualPrismDir, "plugins")
    qtLibPath := filepath.Join(actualPrismDir, "lib")
    
    // ... rest of environment setup
}
```

### Secondary Solution: Qt Version Detection

```go
func detectQtVersion(prismDir string) (string, error) {
    libDir := filepath.Join(getPrismBaseDir(prismDir), "lib")
    
    // Check for Qt6
    if exists(filepath.Join(libDir, "libQt6Core.so.6")) {
        return "qt6", nil
    }
    
    // Check for Qt5
    if exists(filepath.Join(libDir, "libQt5Core.so.5")) {
        return "qt5", nil
    }
    
    return "", fmt.Errorf("unable to detect Qt version")
}

func buildQtEnvironment(prismDir, jreDir string) []string {
    qtVersion, err := detectQtVersion(prismDir)
    if err != nil {
        logf("Warning: %v", err)
        qtVersion = "qt6" // Default assumption
    }
    
    // Adjust environment based on detected Qt version
    if qtVersion == "qt5" {
        // Qt5-specific environment setup
    } else {
        // Qt6-specific environment setup
    }
}
```

### Tertiary Solution: Wrapper Script Detection

```go
func findPrismWrapperScript(prismDir string) string {
    actualPrismDir := getPrismBaseDir(prismDir)
    
    wrapperScripts := []string{
        "prismlauncher.sh",
        "launch-prism.sh", 
        "run-prism.sh",
    }
    
    for _, script := range wrapperScripts {
        scriptPath := filepath.Join(actualPrismDir, script)
        if exists(scriptPath) {
            return scriptPath
        }
    }
    
    return ""
}

func launchPrismWithWrapper(prismDir, wrapperScript string, args []string) error {
    actualPrismDir := getPrismBaseDir(prismDir)
    
    cmd := exec.Command(wrapperScript, args...)
    cmd.Dir = actualPrismDir
    cmd.Env = append(os.Environ(), qtEnv...)
    
    return cmd.Run()
}
```

## Implementation Priority

### Immediate (Critical)
1. **Fix directory structure detection** - This is the most likely root cause
2. **Add Qt version detection** - Ensures correct environment setup
3. **Enhance debugging output** - Better visibility into actual structure

### Short-term (High)
4. **Implement wrapper script detection** - Use official launch methods when available
5. **Add extraction validation** - Verify portable build integrity after download

### Medium-term (Medium)
6. **Consider AppImage alternative** - More reliable portable format
7. **Implement fallback mechanisms** - Graceful degradation when structure is unexpected

## Testing Strategy

### 1. Use Diagnostic Tools
```bash
# Test with actual Prism portable build
go run debug_tools/portable_structure_debug/prism_portable_structure_debug.go \
    https://github.com/PrismLauncher/PrismLauncher/releases/download/9.4/PrismLauncher-Linux-Qt6-Portable-9.4.tar.gz

# Test existing installation
go run debug_tools/portable_structure_debug/prism_portable_structure_debug.go /path/to/prism
```

### 2. Validate Directory Structure
- Check for nested versioned directories
- Verify Qt library and plugin locations
- Confirm executable paths

### 3. Test Environment Setup
- Verify QT_PLUGIN_PATH points to actual plugins directory
- Confirm LD_LIBRARY_PATH includes actual lib directory
- Test with both Qt5 and Qt6 builds

## Conclusion

The Qt library issue is most likely caused by **incorrect assumptions about the Prism Launcher portable build directory structure**. The portable builds extract to versioned subdirectories, but the current code assumes a flat structure. This causes environment variables to point to non-existent paths, resulting in the persistent Qt plugin loading errors.

The solution involves:
1. **Detecting the actual directory structure** (flat vs nested)
2. **Adjusting paths accordingly** for plugins, libraries, and executable
3. **Validating Qt version** to ensure correct environment setup
4. **Using wrapper scripts** when available

The diagnostic tools created will help validate these assumptions and confirm the root cause before implementing the fix.