Dev -> Main release workflow

This repository includes an automated dev branch flow and helper scripts to bump versions and produce releases.

Flow overview

- Push to `dev` branch:
  - `dev-ci-autobump.yml` runs `scripts/auto-bump.sh` which increments patch, sets `PRERELEASE=dev.<short-sha>`, commits and tags the branch, then dispatches the main `build.yml` workflow to build artifacts.
  - The CI will produce prerelease artifacts and tags like `v3.2.7-dev.<sha>`.

- Promote to `main` branch:
  - Create a PR from `dev` to `main`. When merged, the main `build.yml` will detect the tag or you can manually tag for a release.

Helpers

- `scripts/bump-version.ps1 <version> [-Tag]` — updates `version.env` locally; use `-Tag` to commit & tag.
- `scripts/auto-bump.sh` — run by CI on pushes to `dev`; increments patch, sets `PRERELEASE=dev.<sha>`, commits & tags and pushes.

Notes

- The CI `GITHUB_TOKEN` is used to push commits/tags from the autobump job; ensure the token has repo write permissions (default in Actions).
- You can disable autobump by protecting the `dev` branch or adjusting `dev-ci-autobump.yml`.

Release naming guidance

- To make prerelease detection robust, use a consistent prerelease tag prefix for dev autobumps. For example:

  - v3.2.7-dev.<short-sha>

  The updater looks for tags containing the string "dev" when the user has enabled "dev builds" in the launcher settings. Using a consistent prefix like `dev.` makes the regex-based detection simple and reliable.

  If you use a different pattern for prereleases, update the detection regex in `update.go` (function `fetchLatestAssetPreferPrerelease`) to match your scheme.

If you want, I can also add an action to auto-create GitHub Releases when a tag is created (stable release on `main`), and optionally publish artifacts to GitHub Releases automatically.