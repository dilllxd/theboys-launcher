#!/bin/bash
# Quick build verification script for TheBoysLauncher
# Run this before committing changes to ensure basic compilation

echo "🔍 Running quick build verification..."

# Store current directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Check if required files exist
if [ ! -f "go.mod" ]; then
    echo "❌ Error: go.mod not found. Are you in the right directory?"
    exit 1
fi

if [ ! -f "Makefile" ]; then
    echo "❌ Error: Makefile not found. Are you in the right directory?"
    exit 1
fi

# Run quick build check (without creating final executable)
echo "📦 Checking Go compilation..."
if ! go build -o /tmp/theboys-test-build .; then
    echo "❌ Build failed! Please fix compilation errors before committing."
    exit 1
fi

# Clean up test build
rm -f /tmp/theboys-test-build

# Check for common Go issues
echo "🔍 Running basic Go checks..."
if ! go fmt ./...; then
    echo "⚠️  Some files need formatting. Run 'go fmt ./...' to fix."
fi

if ! go vet ./...; then
    echo "❌ Go vet found issues. Please fix before committing."
    exit 1
fi

echo "✅ Build verification passed! Ready to commit."
exit 0