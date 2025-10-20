#!/bin/bash

# Set Version Script for TheBoys Launcher
# This script updates the version in version.env and optionally updates related files

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Usage information
usage() {
    echo "Usage: $0 <major.minor.patch> [--update-inno]"
    echo ""
    echo "Examples:"
    echo "  $0 3.2.1              # Update version to 3.2.1"
    echo "  $0 3.3.0 --update-inno # Update version and WiX file"
    echo ""
    echo "Options:"
    echo "  --update-inno    Also update the WiX .wxs file"
    echo "  --help, -h       Show this help message"
    exit 1
}

# Parse arguments
if [[ $# -eq 0 || "$1" == "--help" || "$1" == "-h" ]]; then
    usage
fi

NEW_VERSION="$1"
UPDATE_INNO=false

if [[ "$2" == "--update-inno" ]]; then
    UPDATE_INNO=true
fi

# Validate version format
if [[ ! "$NEW_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}❌ Invalid version format: $NEW_VERSION${NC}"
    echo -e "${YELLOW}Expected format: major.minor.patch (e.g., 3.2.1)${NC}"
    exit 1
fi

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION_FILE="$PROJECT_ROOT/version.env"

# Check if version.env exists
if [[ ! -f "$VERSION_FILE" ]]; then
    echo -e "${RED}❌ Version file not found: $VERSION_FILE${NC}"
    exit 1
fi

# Extract current version
CURRENT_VERSION=$(grep "^VERSION=" "$VERSION_FILE" | cut -d'=' -f2)

echo -e "${BLUE}🔄 Updating version...${NC}"
echo -e "   Current version: ${YELLOW}$CURRENT_VERSION${NC}"
echo -e "   New version: ${GREEN}$NEW_VERSION${NC}"
echo ""

# Backup current version.env
cp "$VERSION_FILE" "$VERSION_FILE.backup"
echo -e "${BLUE}💾 Backed up current version.env to version.env.backup${NC}"

# Extract version components
IFS='.' read -ra PARTS <<< "$NEW_VERSION"
MAJOR="${PARTS[0]}"
MINOR="${PARTS[1]}"
PATCH="${PARTS[2]}"

# Update version.env
sed -i.bak "s/^VERSION=.*/VERSION=$NEW_VERSION/" "$VERSION_FILE"
sed -i.bak "s/^MAJOR=.*/MAJOR=$MAJOR/" "$VERSION_FILE"
sed -i.bak "s/^MINOR=.*/MINOR=$MINOR/" "$VERSION_FILE"
sed -i.bak "s/^PATCH=.*/PATCH=$PATCH/" "$VERSION_FILE"

# Remove backup files created by sed
rm -f "$VERSION_FILE.bak"

echo -e "${GREEN}✅ Updated version.env${NC}"

# Update WiX file if requested
if [[ "$UPDATE_INNO" == true ]]; then
    WIX_FILE="$PROJECT_ROOT/wix/TheBoysLauncher.wxs"
    if [[ -f "$WIX_FILE" ]]; then
        sed -i.bak "s/Version=\"[^\"]*\"/Version=\"$NEW_VERSION\"/g" "$WIX_FILE"
        rm -f "$WIX_FILE.bak"
        echo -e "${GREEN}✅ Updated wix/TheBoysLauncher.wxs${NC}"
    else
        echo -e "${YELLOW}⚠️  WiX file not found: $WIX_FILE${NC}"
    fi
fi

# Run validation
echo ""
echo -e "${BLUE}🔍 Running version validation...${NC}"
chmod +x "$SCRIPT_DIR/validate-version.sh"
"$SCRIPT_DIR/validate-version.sh"

echo ""
echo -e "${GREEN}🎉 Version update completed successfully!${NC}"
echo -e "${BLUE}💡 Next steps:${NC}"
echo -e "   1. Review the changes with: git diff"
echo -e "   2. Commit the changes: git add . && git commit -m \"chore: bump version to $NEW_VERSION\""
echo -e "   3. Create a tag: git tag $NEW_VERSION"
echo -e "   4. Push changes and tag: git push && git push --tags"