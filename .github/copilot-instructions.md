TheBoysLauncher — Copilot instructions

Quick orientation
- This is a cross-platform Go (1.23+) GUI application that bootstraps Prism Launcher, manages Java runtimes, and synchronizes modpacks (packwiz).
- Entry points: `main.go` (app startup), `launcher.go` / `runLauncherLogic` (core orchestration), `gui.go` (Fyne GUI wiring).

Architecture & important flows (high-level)
- Startup: `main.go` initializes logging, settings, and modpack list, then constructs the GUI via `NewGUI(...)` in `gui.go`.
- Launch flow: `launcher.go` implements modpack orchestration: ensure Prism (`prism.go`/`prism*`), ensure Java (`java.go`), ensure packwiz/bootstrap (`packwiz.go`/`download.go`), create MultiMC/Prism instance, run packwiz to sync mods, then invoke Prism to launch the instance.
- Platform abstraction: platform-specific behavior is implemented with per-OS files (`*_windows.go`, `*_darwin.go`, `*_linux.go`) — prefer reading platform.go for abstractions and OS-specific files for implementation details.

Build / dev workflows (exact commands)
- Local quick build (Windows default target):
  - make build
  - or: `go build -ldflags="-s -w -X main.version=vX.Y.Z" -o TheBoysLauncher .`
- Cross-platform / packaging: The `Makefile` contains convenient targets: `make build-windows`, `make build-macos-universal`, `make build-all`, `make package-macos-universal`, `make package-all`.
- Windows resource embedding: `tools/build.ps1` and `tools/build.ps1` call `rsrc`/`go build` to embed icon and version into `TheBoysLauncher.exe`.
- CI: GitHub Actions workflow is at `.github/workflows/build.yml` (native builds per OS; note it builds natively on runners rather than cross-compiling OpenGL/Fyne targets).

Conventions & patterns to follow
- Single binary: project is designed to produce a single native executable per platform. Avoid adding long-running background services or multiple binaries unless adding packaging artifacts.
- Platform extensions: follow existing pattern of a neutral declaration in `platform.go` and implementations in `platform_<os>.go` files. Use build tags implicitly provided by Go filename conventions.
- Data directories: launcher stores data in platform-specific home locations; use `getLauncherHome()` patterns already used in `main.go`/`config.go` when adding features.
- External downloads: network calls that download Java, Prism, or packwiz assets live in `download.go` and `java.go`. Respect retry/backoff and write stable temporary files to `util/` under the launcher home.

Key files to inspect when modifying behavior
- `main.go` — startup, logging, and signal handling
- `launcher.go` — high-level launch orchestration and calls into packwiz/Prism
- `java.go` — Java version resolution & installation; critical when changing Java sourcing or download logic
- `prism.go` — ensures Prism Launcher is present and updates prismlauncher.cfg
- `packwiz.go` / `download.go` — packwiz bootstrap, download helpers, and unpack logic
- `gui.go` — Fyne GUI entrypoints, progress callbacks, and user interactions
- `Makefile`, `tools/build.ps1`, `scripts/*` — build, packaging and versioning scripts

Testing & verification
- Run `make verify` for a quick compilation check. `make test` runs build + lint + runtime smoke test.
- CI runs platform-native builds in `.github/workflows/build.yml`; mimic runner environment (Ubuntu/macOS/Windows) when debugging CI failures.

Common pitfalls and gotchas
- Fyne/OpenGL builds can fail when cross-compiling — the repo builds natively in CI for macOS/Linux/Windows. Prefer native builds or the existing Makefile targets.
- macOS universal binaries are produced by building amd64 and arm64 then running `lipo` (see `Makefile` and `scripts/create-app-bundle.sh`).
- Resource embedding on Windows requires `rsrc` (or `goversioninfo`) and `icon.ico` at repo root; `Makefile` and `tools/build.ps1` assume `icon.ico` exists.
- Packwiz must be executed from the instance's `minecraft` working directory — see `runLauncherLogic` where `cmd.Dir = mcDir` is critical.

If you change download/installation code
- Write idempotent installers: `java.go`, `prism.go`, and `packwiz` helpers are defensive (check for existing files). Keep that pattern.
- Use `downloadTo(..., 0755)` and `os.MkdirAll(...)` patterns already used in the codebase.

Search hints for contributors
- To find platform-specific behavior: grep for `// +build` or search `platform_` and `*_windows.go`/`*_darwin.go`/`*_linux.go` files.
- To find where the GUI triggers launch logic: search for `launchWithCallback` or `runLauncherLogic`.

What I couldn't discover automatically
- Any external credentials or keys are not present in the repo (good). If you need code-signing credentials or distribution signing, the repo expects external provisioning (see `wix/` and scripts); coordinate with maintainers for those secrets.

If this file is outdated or missing details, reply with specific areas you want added (examples: deeper call graphs, specific function contracts, or test harnesses).
