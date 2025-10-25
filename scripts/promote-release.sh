#!/usr/bin/env bash
set -euo pipefail

# Promote the current dev build to a stable release.
# Expected to run on pushes to main after dev has been merged.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  echo "GITHUB_TOKEN not provided; this script expects to run in GitHub Actions with GITHUB_TOKEN set." >&2
  exit 1
fi

git config user.name "github-actions[bot]"
git config user.email "github-actions[bot]@users.noreply.github.com"

if [[ ! -f version.env ]]; then
  echo "version.env not found" >&2
  exit 1
fi

source ./version.env

if [[ -z "${PRERELEASE:-}" ]]; then
  echo "version.env already represents a stable release ($VERSION); nothing to promote."
  exit 0
fi

if [[ ! "$PRERELEASE" =~ ^dev\. ]]; then
  echo "Refusing to promote prerelease '$PRERELEASE'. Expected prefix 'dev.'." >&2
  exit 1
fi

# Increment the PATCH version for stable release
PATCH=$((PATCH + 1))
STABLE_VERSION="$MAJOR.$MINOR.$PATCH"

# Log the incremented version
echo "Incremented version to $STABLE_VERSION"

echo "Promoting $VERSION-$PRERELEASE to stable release $STABLE_VERSION"

cat > version.env <<EOF
# TheBoysLauncher Version Configuration (stabilized)
VERSION=$STABLE_VERSION
MAJOR=$MAJOR
MINOR=$MINOR
PATCH=$PATCH
BUILD_METADATA=
PRERELEASE=

# Full version string is constructed by scripts/get-version.sh
EOF

git add version.env
if git diff --cached --quiet; then
  echo "No changes detected after promotion; exiting."
  exit 0
fi

git commit -m "ci: promote release $STABLE_VERSION"

git push origin HEAD

TAG="v${STABLE_VERSION}"
git tag -a "$TAG" -m "Stable release $TAG"
git push origin "$TAG"

echo "Promoted stable release $TAG"
