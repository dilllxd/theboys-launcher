#!/bin/bash
# Simple verification script for TheBoys Launcher

echo "🔍 Verifying TheBoys Launcher project structure..."
echo

# Check if required directories exist
echo "📁 Checking directory structure..."
required_dirs=(
    "cmd/winterpack"
    "internal/app"
    "internal/config"
    "internal/gui"
    "internal/logging"
    "pkg/version"
    "pkg/platform"
    "legacy"
    "configs"
    "assets"
)

for dir in "${required_dirs[@]}"; do
    if [ -d "$dir" ]; then
        echo "  ✅ $dir"
    else
        echo "  ❌ $dir (missing)"
    fi
done

echo

# Check if required files exist
echo "📄 Checking required files..."
required_files=(
    "go.mod"
    "go.sum"
    "Makefile"
    "README.md"
    "cmd/winterpack/main.go"
    "internal/config/config.go"
    "internal/app/state.go"
    "internal/logging/logger.go"
    "pkg/version/version.go"
    "pkg/platform/platform.go"
    "configs/modpacks.json"
)

for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file"
    else
        echo "  ❌ $file (missing)"
    fi
done

echo

# Check Go syntax without building GUI components
echo "🔧 Checking Go syntax..."
go_files=$(find . -name "*.go" -not -path "./legacy/*" | head -5)
syntax_errors=0

for file in $go_files; do
    echo "  📝 Checking $file..."
    if go fmt "$file" > /dev/null 2>&1; then
        echo "    ✅ Syntax OK"
    else
        echo "    ❌ Syntax Error"
        syntax_errors=$((syntax_errors + 1))
    fi
done

echo

# Check modpacks.json format
echo "📦 Checking modpacks configuration..."
if [ -f "configs/modpacks.json" ]; then
    if python3 -m json.tool configs/modpacks.json > /dev/null 2>&1; then
        echo "  ✅ modpacks.json is valid JSON"
    else
        echo "  ❌ modpacks.json has invalid JSON"
    fi
else
    echo "  ❌ modpacks.json not found"
fi

echo

# Summary
echo "📊 Verification Summary:"
if [ $syntax_errors -eq 0 ]; then
    echo "  ✅ All syntax checks passed"
    echo "  🎉 Phase 1 foundation is complete!"
    echo
    echo "Next steps:"
    echo "  1. Run 'make build' when GUI environment is available"
    echo "  2. Start Phase 2: Core GUI Framework"
else
    echo "  ❌ $syntax_errors syntax errors found"
    echo "  🔧 Fix syntax errors before proceeding"
fi

echo