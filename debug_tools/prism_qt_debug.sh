#!/bin/bash

# Prism Launcher Qt Debug Script
# Usage: ./prism_qt_debug.sh <prism_directory>

if [ $# -lt 1 ]; then
    echo "Usage: $0 <prism_directory>"
    exit 1
fi

PRISM_DIR="$1"
JRE_DIR="$PRISM_DIR/java/jre17"  # Assuming JRE 17

echo "=== Prism Launcher Qt Debug Tool ==="
echo "Prism Directory: $PRISM_DIR"
echo ""

# 1. Environment Variable Analysis
echo "1. Environment Variable Analysis:"
echo "================================"

# Simulate the environment variables that would be set
QT_PLUGIN_PATH="$PRISM_DIR/plugins"
LD_LIBRARY_PATH="$PRISM_DIR/lib"
JAVA_HOME="$JRE_DIR"
PATH="$JRE_DIR/bin:$PATH"

echo "Qt environment variables that would be set:"
echo "  QT_PLUGIN_PATH=$QT_PLUGIN_PATH"
echo "  LD_LIBRARY_PATH=$LD_LIBRARY_PATH"
echo "  JAVA_HOME=$JAVA_HOME"
echo "  QT_QPA_PLATFORM=xcb"
echo "  QT_XCB_GL_INTEGRATION=xcb_glx"
echo ""

# 2. Qt Plugin File Verification
echo "2. Qt Plugin File Verification:"
echo "==============================="

if [ ! -d "$PRISM_DIR/plugins" ]; then
    echo "❌ Plugins directory does not exist: $PRISM_DIR/plugins"
else
    echo "✅ Plugins directory exists: $PRISM_DIR/plugins"
    
    # Check for critical plugin directories
    for dir in platforms imageformats iconengines tls; do
        if [ -d "$PRISM_DIR/plugins/$dir" ]; then
            echo "✅ Plugin directory exists: $dir"
        else
            echo "❌ Plugin directory missing: $dir"
        fi
    done
    
    # Check for critical plugin files
    echo ""
    echo "Critical plugin files:"
    for plugin in platforms/libqxcb.so imageformats/libqjpeg.so iconengines/libqsvgicon.so tls/libqopensslbackend.so; do
        if [ -f "$PRISM_DIR/plugins/$plugin" ]; then
            echo "✅ $plugin"
        else
            echo "❌ $plugin"
        fi
    done
fi
echo ""

# 3. Library Directory Analysis
echo "3. Library Directory Analysis:"
echo "==============================="

if [ ! -d "$PRISM_DIR/lib" ]; then
    echo "❌ Library directory does not exist: $PRISM_DIR/lib"
else
    echo "✅ Library directory exists: $PRISM_DIR/lib"
    
    # Check for Qt libraries
    qt_libs=$(find "$PRISM_DIR/lib" -name "libQt*.so*" 2>/dev/null | wc -l)
    if [ "$qt_libs" -eq 0 ]; then
        echo "❌ No Qt libraries found in $PRISM_DIR/lib"
    else
        echo "✅ Found $qt_libs Qt libraries:"
        find "$PRISM_DIR/lib" -name "libQt*.so*" 2>/dev/null | head -10 | while read lib; do
            echo "  - $(basename "$lib")"
        done
        if [ "$qt_libs" -gt 10 ]; then
            echo "  ... ($(($qt_libs - 10)) more libraries)"
        fi
    fi
fi
echo ""

# 4. Prism Executable Analysis
echo "4. Prism Executable Analysis:"
echo "============================="

PRISM_EXE="$PRISM_DIR/PrismLauncher"
if [ ! -f "$PRISM_EXE" ]; then
    echo "❌ Prism executable does not exist: $PRISM_EXE"
else
    echo "✅ Prism executable exists: $PRISM_EXE"
    
    # Check if it's executable
    if [ -x "$PRISM_EXE" ]; then
        echo "✅ Prism executable has execute permissions"
    else
        echo "❌ Prism executable lacks execute permissions"
    fi
    
    # Try to get library dependencies
    if command -v ldd >/dev/null 2>&1; then
        echo ""
        echo "Checking Prism library dependencies:"
        ldd "$PRISM_EXE" 2>/dev/null | grep -E "(Qt|not found)" | while read line; do
            echo "  $line"
        done
    fi
fi
echo ""

# 5. Wrapper Script Analysis
echo "5. Wrapper Script Analysis:"
echo "==========================="

for script in launch-prism.sh prism-launcher.sh run-prism.sh; do
    if [ -f "$PRISM_DIR/$script" ]; then
        echo "✅ Found wrapper script: $script"
        
        # Check script content
        if grep -q "QT_PLUGIN_PATH" "$PRISM_DIR/$script"; then
            echo "  ✅ Script sets QT_PLUGIN_PATH"
        else
            echo "  ❌ Script does not set QT_PLUGIN_PATH"
        fi
        
        if grep -q "LD_LIBRARY_PATH" "$PRISM_DIR/$script"; then
            echo "  ✅ Script sets LD_LIBRARY_PATH"
        else
            echo "  ❌ Script does not set LD_LIBRARY_PATH"
        fi
    fi
done
echo ""

# 6. System Qt Analysis
echo "6. System Qt Analysis:"
echo "====================="

# Check for system Qt packages
if command -v dpkg >/dev/null 2>&1; then
    echo "Checking system Qt packages (dpkg):"
    qt5_packages=$(dpkg -l | grep -i qt5 | wc -l)
    qt6_packages=$(dpkg -l | grep -i qt6 | wc -l)
    
    if [ "$qt5_packages" -gt 0 ]; then
        echo "✅ Found $qt5_packages Qt5 packages"
        dpkg -l | grep -i qt5 | head -3 | while read line; do
            echo "  $line"
        done
    else
        echo "ℹ️  No Qt5 packages found"
    fi
    
    if [ "$qt6_packages" -gt 0 ]; then
        echo "✅ Found $qt6_packages Qt6 packages"
        dpkg -l | grep -i qt6 | head -3 | while read line; do
            echo "  $line"
        done
    else
        echo "ℹ️  No Qt6 packages found"
    fi
elif command -v rpm >/dev/null 2>&1; then
    echo "Checking system Qt packages (rpm):"
    qt_packages=$(rpm -qa | grep -i qt | wc -l)
    if [ "$qt_packages" -gt 0 ]; then
        echo "✅ Found $qt_packages Qt packages"
        rpm -qa | grep -i qt | head -3 | while read line; do
            echo "  $line"
        done
    else
        echo "ℹ️  No Qt packages found"
    fi
fi
echo ""

# 7. Environment Variable Test
echo "7. Environment Variable Test:"
echo "============================"

# Create a test script to verify environment variables
TEST_SCRIPT="$PRISM_DIR/test_env.sh"
cat > "$TEST_SCRIPT" << 'EOF'
#!/bin/bash
echo "=== Environment Test ==="
echo "QT_PLUGIN_PATH: $QT_PLUGIN_PATH"
echo "LD_LIBRARY_PATH: $LD_LIBRARY_PATH"
echo "JAVA_HOME: $JAVA_HOME"
echo "PATH: $PATH"
echo "========================"

# Test if Qt plugins are accessible
if [ -n "$QT_PLUGIN_PATH" ]; then
    echo "Testing Qt plugin access:"
    if [ -f "$QT_PLUGIN_PATH/platforms/libqxcb.so" ]; then
        echo "✅ libqxcb.so found"
    else
        echo "❌ libqxcb.so not found"
    fi
    
    if [ -f "$QT_PLUGIN_PATH/imageformats/libqjpeg.so" ]; then
        echo "✅ libqjpeg.so found"
    else
        echo "❌ libqjpeg.so not found"
    fi
else
    echo "❌ QT_PLUGIN_PATH not set"
fi
EOF

chmod +x "$TEST_SCRIPT"
echo "✅ Created test script: $TEST_SCRIPT"

# Run the test script with the simulated environment
echo "Test script output:"
(
    export QT_PLUGIN_PATH="$QT_PLUGIN_PATH"
    export LD_LIBRARY_PATH="$LD_LIBRARY_PATH"
    export JAVA_HOME="$JAVA_HOME"
    export PATH="$PATH"
    export QT_QPA_PLATFORM="xcb"
    export QT_XCB_GL_INTEGRATION="xcb_glx"
    
    cd "$PRISM_DIR"
    bash "$TEST_SCRIPT"
)

# Clean up
rm -f "$TEST_SCRIPT"

echo ""
echo "=== Debug Analysis Complete ==="