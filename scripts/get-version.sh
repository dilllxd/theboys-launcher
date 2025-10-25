#!/bin/bash

# Get-Version Script for TheBoysLauncher
# This script reads version information and exports it in various formats

set -euo pipefail

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Source version configuration
if [[ -f "$PROJECT_ROOT/version.env" ]]; then
    source "$PROJECT_ROOT/version.env"
else
    echo "Error: version.env file not found at $PROJECT_ROOT/version.env" >&2
    exit 1
fi

# Construct full version string
FULL_VERSION="$VERSION"
if [[ -n "$PRERELEASE" ]]; then
    FULL_VERSION="$FULL_VERSION-$PRERELEASE"
fi
if [[ -n "$BUILD_METADATA" ]]; then
    FULL_VERSION="$FULL_VERSION+$BUILD_METADATA"
fi

# Export all version variables
export VERSION
export MAJOR
export MINOR
export PATCH
export PRERELEASE
export BUILD_METADATA
export FULL_VERSION

# Output functions
show_version() {
    echo "$FULL_VERSION"
}

show_version_json() {
    cat << EOF
{
    "version": "$VERSION",
    "major": $MAJOR,
    "minor": $MINOR,
    "patch": $PATCH,
    "prerelease": "$PRERELEASE",
    "build_metadata": "$BUILD_METADATA",
    "full_version": "$FULL_VERSION"
}
EOF
}

show_version_make() {
    echo "VERSION=$FULL_VERSION"
    echo "MAJOR=$MAJOR"
    echo "MINOR=$MINOR"
    echo "PATCH=$PATCH"
}

# Main execution
case "${1:-version}" in
    "version"|"")
        show_version
        ;;
    "json")
        show_version_json
        ;;
    "make")
        show_version_make
        ;;
    "export")
        echo "export VERSION=\"$VERSION\""
        echo "export MAJOR=\"$MAJOR\""
        echo "export MINOR=\"$MINOR\""
        echo "export PATCH=\"$PATCH\""
        echo "export PRERELEASE=\"$PRERELEASE\""
        echo "export BUILD_METADATA=\"$BUILD_METADATA\""
        echo "export FULL_VERSION=\"$FULL_VERSION\""
        ;;
    "validate")
        # Validate version format
        if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            echo "Error: Invalid version format: $VERSION" >&2
            exit 1
        fi
        echo "✅ Version format is valid: $VERSION"
        ;;
    *)
        echo "Usage: $0 [version|json|make|export|validate]"
        echo "  version  - Show full version (default)"
        echo "  json     - Show version as JSON"
        echo "  make     - Show version in Makefile format"
        echo "  export   - Show version as export statements"
        echo "  validate - Validate version format"
        exit 1
        ;;
esac