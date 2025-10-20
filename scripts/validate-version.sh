#!/bin/bash

# Version Validation Script for TheBoys Launcher
# This script validates version consistency across all files

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Source version configuration
if [[ -f "$PROJECT_ROOT/version.env" ]]; then
    source "$PROJECT_ROOT/version.env"
else
    echo -e "${RED}❌ Version file not found: $PROJECT_ROOT/version.env${NC}"
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

echo -e "${BLUE}🔍 Validating version consistency across project files...${NC}"
echo -e "${BLUE}📋 Target version: $FULL_VERSION${NC}"
echo ""

# Validation functions
validate_semver() {
    local version="$1"
    if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$ ]]; then
        echo -e "${RED}❌ Invalid semantic version format: $version${NC}"
        return 1
    fi
    return 0
}

check_file_contains_version() {
    local file="$1"
    local pattern="$2"
    local description="$3"

    if [[ -f "$PROJECT_ROOT/$file" ]]; then
        if grep -q "$pattern" "$PROJECT_ROOT/$file"; then
            echo -e "${GREEN}✅ $description: $file${NC}"
            return 0
        else
            echo -e "${RED}❌ $description not found in: $file${NC}"
            echo -e "   Expected pattern: $pattern"
            return 1
        fi
    else
        echo -e "${YELLOW}⚠️  File not found: $file${NC}"
        return 0  # Not an error for optional files
    fi
}

echo -e "${BLUE}📝 Checking version format...${NC}"
if validate_semver "$FULL_VERSION"; then
    echo -e "${GREEN}✅ Semantic version format is valid${NC}"
else
    echo -e "${RED}❌ Semantic version format is invalid${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}🔍 Checking version references in files...${NC}"

# Check version.env
echo -e "${BLUE}📁 version.env${NC}"
if [[ "$VERSION" == "$(grep "^VERSION=" "$PROJECT_ROOT/version.env" | cut -d'=' -f2)" ]]; then
    echo -e "${GREEN}✅ VERSION matches in version.env${NC}"
else
    echo -e "${RED}❌ VERSION mismatch in version.env${NC}"
    exit 1
fi

# Check GitHub Actions workflow
echo -e "${BLUE}📁 .github/workflows/build.yml${NC}"
check_file_contains_version ".github/workflows/build.yml" "VERSION_FILE: version.env" "VERSION_FILE reference"

# Check WiX file
echo -e "${BLUE}📁 wix/TheBoysLauncher.wxs${NC}"
if check_file_contains_version "wix/TheBoysLauncher.wxs" 'Version="' "Version definition"; then
    # Extract version from WiX file
    WIX_VERSION=$(grep -o 'Version="[^"]*"' "$PROJECT_ROOT/wix/TheBoysLauncher.wxs" | head -1 | cut -d'"' -f2)
    if [[ "$WIX_VERSION" == "$VERSION" ]]; then
        echo -e "${GREEN}✅ WiX version matches ($WIX_VERSION)${NC}"
    else
        echo -e "${RED}❌ WiX version mismatch: expected $VERSION, found $WIX_VERSION${NC}"
        echo -e "${YELLOW}💡 Run: ./scripts/update-wix-version.ps1${NC}"
    fi
fi

# Check Go source files for version references (optional)
echo -e "${BLUE}📁 Go source files${NC}"

# Use a simpler approach without process substitution
if find "$PROJECT_ROOT" -name "*.go" -type f >/dev/null 2>&1; then
    GO_FILES_WITH_VERSION=$(find "$PROJECT_ROOT" -name "*.go" -type f -exec grep -l "main.version" {} \; 2>/dev/null | wc -l)
    GO_FILES_TOTAL=$(find "$PROJECT_ROOT" -name "*.go" -type f 2>/dev/null | wc -l)

    if [[ $GO_FILES_WITH_VERSION -gt 0 ]]; then
        echo -e "${GREEN}✅ Found $GO_FILES_WITH_VERSION Go file(s) with version references${NC}"
    else
        echo -e "${YELLOW}⚠️  No Go files found with main.version references${NC}"
    fi
else
    echo -e "${YELLOW}⚠️  No Go files found${NC}"
fi

# Check build scripts
echo -e "${BLUE}📁 Build scripts${NC}"
check_file_contains_version "Makefile" "VERSION=" "VERSION variable"
check_file_contains_version "tools/build.bat" "VERSION%" "VERSION variable"
check_file_contains_version "tools/build.ps1" "VERSION" "VERSION variable"

echo ""
echo -e "${BLUE}🎯 Version Summary:${NC}"
echo -e "   Version: ${GREEN}$VERSION${NC}"
echo -e "   Major: ${GREEN}$MAJOR${NC}"
echo -e "   Minor: ${GREEN}$MINOR${NC}"
echo -e "   Patch: ${GREEN}$PATCH${NC}"
if [[ -n "$PRERELEASE" ]]; then
    echo -e "   Prerelease: ${YELLOW}$PRERELEASE${NC}"
fi
if [[ -n "$BUILD_METADATA" ]]; then
    echo -e "   Build Metadata: ${YELLOW}$BUILD_METADATA${NC}"
fi
echo -e "   Full Version: ${GREEN}$FULL_VERSION${NC}"

echo ""
echo -e "${BLUE}💡 Tips for version updates:${NC}"
echo -e "   1. Update $PROJECT_ROOT/version.env"
echo -e "   2. Run ./scripts/update-wix-version.ps1 (Windows)"
echo -e "   3. Commit changes and push"

echo ""
echo -e "${GREEN}✅ Version validation completed!${NC}"