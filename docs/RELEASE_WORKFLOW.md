Dev → Main Release Workflow
============================

This repository now automates both the nightly-style dev builds and the stable releases that follow. Use this doc as the playbook for how the CI handles each phase.

Flow overview
-------------

- Push to `dev`  
  `.github/workflows/autobump-dev.yml` runs `scripts/auto-bump.sh`, which bumps the patch, sets `PRERELEASE=dev.<short-sha>`, commits the updated `version.env`, and tags the commit (`v3.2.16-dev.<sha>`).  
  Every `v*` tag triggers `.github/workflows/build-on-version-change.yml`, building installers/binaries for Linux, Windows, and macOS and publishing a GitHub prerelease tied to that tag.

- Merge `dev` → `main`  
  Once the PR lands, `.github/workflows/promote-main-release.yml` runs on `main`. It executes `scripts/promote-release.sh`, which strips the `dev.*` suffix to stabilise `version.env`, commits `ci: promote release <version>`, and tags `v<version>`.  
  The stable tag re-runs the build workflow; because the tag lacks `-dev`, the release is published as a normal (non-prerelease) build with the same artifacts.

Helpers
-------

- `scripts/auto-bump.sh` – CI helper for dev pushes; do not run manually.  
- `scripts/promote-release.sh` – Converts the latest dev build into a stable release; run automatically on `main`, but also available locally if needed.  
- `scripts/bump-version.ps1 <version> [-Tag]` – Manual override when you need to force a specific version.

Operational notes
-----------------

- Add a classic PAT with `repo` and `workflow` scopes to the `PERSONAL_ACCESS_TOKEN` secret so CI commits/tags can trigger downstream workflows.  
- After the stable promotion finishes, merge `main` back into `dev` so future autobumps build on top of the latest stable version.  
- You can pause autobumps by disabling the workflow or protecting the `dev` branch.

Naming guidance
---------------

- Keep the prerelease prefix as `dev.` so `update.go` → `fetchLatestAssetPreferPrerelease` can distinguish dev builds during automatic update checks.  
- Stable tags should be plain `v<major>.<minor>.<patch>` so GitHub Releases default to non-prerelease mode. Adjust the updater regex only if you change this convention.
