# TheBoysLauncher GitHub Workflow Documentation

## Table of Contents
1. [Executive Summary](#executive-summary)
2. [Development Workflow](#development-workflow)
3. [PR-Based Stable Release Process](#pr-based-stable-release-process)
4. [CI/CD Pipeline Details](#cicd-pipeline-details)
5. [Build Process](#build-process)
6. [Version Management System](#version-management-system)
7. [Release Process](#release-process)
8. [Testing and Quality Assurance](#testing-and-quality-assurance)
9. [Infrastructure and Dependencies](#infrastructure-and-dependencies)
10. [Key Files and Their Roles](#key-files-and-their-roles)
11. [Workflow Diagram](#workflow-diagram)

---

## Executive Summary

TheBoysLauncher project implements a sophisticated GitHub workflow that automates the entire software development lifecycle from code commits to user releases. This workflow is designed to provide continuous delivery of development builds while maintaining a structured release process for stable versions through Pull Request-based automation.

### Key Features
- **Automated Dev Builds**: Every push to the `dev` branch triggers an automated build and release process
- **PR-Based Stable Releases**: Pull Requests from `dev` to `main` automatically trigger stable releases
- **Cross-Platform Support**: Simultaneous builds for Windows, macOS (Intel/ARM64/Universal), and Linux
- **Semantic Versioning**: Automated version management with semantic versioning compliance
- **Code Review Integration**: Stable releases require PR review and approval before automation
- **Quality Assurance**: Integrated testing and validation at multiple stages

### Workflow Overview
The workflow operates on a branch-based model where:
- `dev` branch serves as the integration branch for ongoing development
- `main` branch represents stable releases
- Automated CI/CD processes handle building, testing, and releasing
- Version management is centralized and automated
- Pull Requests from `dev` to `main` trigger automated stable releases

---

## Development Workflow

### Branch Strategy

The project follows a simplified but effective branching strategy:

#### `dev` Branch
- **Purpose**: Primary development branch where features are integrated
- **Trigger**: Automatically triggers dev prerelease workflow on every push
- **Versioning**: Automatically increments patch version and adds dev prerelease identifier
- **Artifacts**: Produces prerelease builds for all platforms

#### `main` Branch
- **Purpose**: Stable release branch
- **Source**: Receives code from `dev` branch after testing and validation
- **Releases**: Contains production-ready code
- **Versioning**: Stable versions without prerelease identifiers

### Development Process

1. **Feature Development**
   ```bash
   # Create feature branch from dev
   git checkout -b feature/new-feature dev
   
   # Develop and commit changes
   git add .
   git commit -m "feat: add new feature"
   
   # Push and create PR to dev
   git push origin feature/new-feature
   ```

2. **Integration to dev**
   ```bash
   # Merge PR to dev branch
   git checkout dev
   git merge feature/new-feature
   git push origin dev
   # This triggers automated dev prerelease workflow
   ```

3. **Promotion to Stable via PR**
   ```bash
   # When ready for release, create PR from dev to main
   # 1. Ensure dev branch is up to date
   git checkout dev
   git pull origin dev
   
   # 2. Create PR from dev to main through GitHub UI
   # 3. After PR review and approval, merge to main
   # 4. This automatically triggers stable release workflow
   ```
   
   **Alternative Direct Merge (for maintainers)**
   ```bash
   # Direct merge option (bypasses PR review)
   git checkout main
   git merge dev
   git push origin main
   # This triggers stable release promotion
   ```

### Commit Conventions

The project follows conventional commit messages:
- `feat:` - New features
- `fix:` - Bug fixes
- `docs:` - Documentation changes
- `style:` - Code style changes
- `refactor:` - Code refactoring
- `test:` - Test additions/changes
- `chore:` - Maintenance tasks

---

## PR-Based Stable Release Process

### Overview

The project has implemented an automated stable release process triggered by Pull Requests (PRs) from the `dev` branch to the `main` branch. This process ensures that stable releases are properly reviewed before being automatically published.

### PR Creation Process

1. **Preparation**
   - Ensure all desired changes are merged to the `dev` branch
   - Verify that dev builds have been tested and are functioning correctly
   - Confirm that the `dev` branch is up to date

2. **Creating the PR**
   - Navigate to the GitHub repository
   - Click "New pull request"
   - Select `dev` as the source branch and `main` as the target branch
   - Provide a descriptive title and detailed description of changes
   - Include any testing results or special considerations

3. **Review Process**
   - Team members review the PR for code quality and functionality
   - Automated checks run to ensure build compatibility
   - Any issues found must be resolved before merging

4. **Merging and Release**
   - Once approved, merge the PR using "Create a merge commit"
   - The merge to `main` automatically triggers the stable release workflow
   - A new stable version is built and published without manual intervention

### Automated Stable Release Workflow

When a PR is merged to `main`, the following automated process occurs:

1. **Version Bumping**
   - The PATCH version is automatically incremented
   - Prerelease identifiers are cleared
   - A new stable version tag is created

2. **Build Process**
   - All platform binaries are built with the new stable version
   - Windows MSI installer is created with embedded version information
   - Artifacts are prepared for distribution

3. **Release Publishing**
   - A new GitHub stable release is created
   - All platform binaries and installers are uploaded
   - Changelog is automatically generated from git history

### Benefits of PR-Based Releases

- **Code Review**: Ensures all stable releases are properly reviewed
- **Traceability**: Clear record of what changes are included in each release
- **Automation**: Eliminates manual steps in the release process
- **Quality Control**: Prevents accidental releases from untested code
- **Version Management**: Automatic version incrementing and tagging

### Version Progression

The version progression through the entire process follows this pattern:

1. **Development**: `3.2.32-dev.b05b4ce` (dev branch)
2. **Stable Release**: `3.2.33` (main branch after PR merge)

Each stable release increments the PATCH version from the previous dev build, ensuring a continuous version sequence.

---

## CI/CD Pipeline Details

### Primary Workflow: Dev Prerelease

**File**: [`.github/workflows/dev-prerelease.yml`](../.github/workflows/dev-prerelease.yml)

**Trigger**: Push to `dev` branch (excluding bot commits)

#### Pipeline Stages

##### 1. Version Bumping Stage
```yaml
bump_version:
  if: ${{ github.actor != 'github-actions[bot]' }}
  runs-on: ubuntu-latest
```

**Process**:
- Fetches all tags and version history
- Reads current version from [`version.env`](../version.env)
- Increments PATCH version
- Adds dev prerelease identifier with short SHA: `dev.<short-sha>`
- Updates [`version.env`](../version.env) with new version
- Commits version changes back to repository
- Creates and pushes annotated tag

**Outputs**:
- `version`: Base version (e.g., "3.2.32")
- `full_version`: Complete version with prerelease (e.g., "3.2.32-dev.b05b4ce")
- `prerelease`: Prerelease identifier (e.g., "dev.b05b4ce")
- `tag`: Git tag (e.g., "v3.2.32-dev.b05b4ce")
- `commit`: Commit SHA of version bump

##### 2. Build Matrix Stage
```yaml
build:
  needs: bump_version
  strategy:
    matrix:
      os: [ubuntu-latest, windows-latest, macos-latest]
```

**Platform Configurations**:
- **Windows**: `windows-latest` → `TheBoysLauncher.exe` + MSI installer
- **Linux**: `ubuntu-latest` → `TheBoysLauncher-linux`
- **macOS**: `macos-latest` → `TheBoysLauncher-mac-universal`

**Build Process**:
1. **Checkout**: Uses the specific commit from version bump
2. **Go Setup**: Installs Go 1.23+ with caching
3. **Dependencies**: Downloads Go modules
4. **System Dependencies**: Installs platform-specific build tools
5. **Compilation**: Builds with version ldflags
6. **Packaging**: Creates platform-specific packages

##### 3. Release Publishing Stage
```yaml
publish:
  needs: [bump_version, build]
  runs-on: ubuntu-latest
```

**Process**:
1. **Artifact Collection**: Downloads all build artifacts
2. **Changelog Generation**: Creates changelog from git history
3. **Release Creation**: Creates GitHub prerelease
4. **Asset Upload**: Uploads all platform artifacts

### Workflow Concurrency

```yaml
concurrency:
  group: dev-prerelease
  cancel-in-progress: false
```

- Ensures only one dev prerelease workflow runs at a time
- Does not cancel in-progress workflows to allow completion

### Permissions

```yaml
permissions:
  contents: write
```

- Required for:
  - Creating commits (version bump)
  - Pushing tags
  - Creating releases
  - Uploading artifacts

### Secondary Workflow: Stable Release

**File**: [`.github/workflows/stable-release.yml`](../.github/workflows/stable-release.yml)

**Trigger**: Push to `main` branch (excluding bot commits)

#### Pipeline Stages

##### 1. Version Bumping Stage
```yaml
bump_version:
  if: ${{ github.actor != 'github-actions[bot]' }}
  runs-on: ubuntu-latest
```

**Process**:
- Fetches all tags and version history
- Reads current version from [`version.env`](../version.env)
- Increments PATCH version
- Clears prerelease identifier (creates stable version)
- Updates [`version.env`](../version.env) with new version
- Commits version changes back to repository
- Creates and pushes annotated tag

**Outputs**:
- `version`: Base version (e.g., "3.2.33")
- `full_version`: Complete version without prerelease (e.g., "3.2.33")
- `prerelease`: Empty string for stable releases
- `tag`: Git tag (e.g., "v3.2.33")
- `commit`: Commit SHA of version bump

##### 2. Build Matrix Stage
```yaml
build:
  needs: bump_version
  strategy:
    matrix:
      os: [ubuntu-latest, windows-latest, macos-latest]
```

**Platform Configurations**:
- **Windows**: `windows-latest` → `TheBoysLauncher.exe` + MSI installer
- **Linux**: `ubuntu-latest` → `TheBoysLauncher-linux`
- **macOS**: `macos-latest` → `TheBoysLauncher-mac-universal`

**Build Process**:
1. **Checkout**: Uses the specific commit from version bump
2. **Go Setup**: Installs Go 1.23+ with caching
3. **Dependencies**: Downloads Go modules
4. **System Dependencies**: Installs platform-specific build tools
5. **Compilation**: Builds with version ldflags
6. **Packaging**: Creates platform-specific packages

##### 3. Release Publishing Stage
```yaml
publish:
  needs: [bump_version, build]
  runs-on: ubuntu-latest
```

**Process**:
1. **Artifact Collection**: Downloads all build artifacts
2. **Changelog Generation**: Creates changelog from git history
3. **Release Creation**: Creates GitHub stable release (not prerelease)
4. **Asset Upload**: Uploads all platform artifacts

### Workflow Concurrency

```yaml
concurrency:
  group: stable-release
  cancel-in-progress: false
```

- Ensures only one stable release workflow runs at a time
- Does not cancel in-progress workflows to allow completion

### Permissions

```yaml
permissions:
  contents: write
```

- Required for:
  - Creating commits (version bump)
  - Pushing tags
  - Creating releases
  - Uploading artifacts

---

## Build Process

### Build Matrix Configuration

The build process uses a matrix strategy to build for all supported platforms simultaneously:

#### Windows Build
- **Runner**: `windows-latest`
- **Target**: `TheBoysLauncher.exe`
- **Additional**: MSI installer via WiX
- **Special Requirements**: WiX Toolset for installer creation

**Build Command**:
```bash
go build -ldflags="-s -w -H=windowsgui -X main.version=${{ needs.bump_version.outputs.full_version }}" -o TheBoysLauncher.exe .
```

**Installer Process**:
1. Calls [`scripts/build-installer.ps1`](../scripts/build-installer.ps1)
2. Embeds version information
3. Creates MSI with WiX Toolset
4. Renames installer with version: `TheBoysLauncher-Setup-${VERSION}.msi`

#### Linux Build
- **Runner**: `ubuntu-latest`
- **Target**: `TheBoysLauncher-linux`
- **Dependencies**: System libraries for GUI support

**System Dependencies**:
```bash
sudo apt-get install -y --no-install-recommends \
  build-essential \
  pkg-config \
  libgl1-mesa-dev \
  libglu1-mesa-dev \
  libx11-dev \
  libxrandr-dev \
  libxinerama-dev \
  libxcursor-dev \
  libxi-dev \
  libxxf86vm-dev \
  libasound2-dev
```

**Build Command**:
```bash
go build -ldflags="-s -w -X main.version=${{ needs.bump_version.outputs.full_version }}" -o TheBoysLauncher-linux .
```

#### macOS Build
- **Runner**: `macos-latest`
- **Target**: `TheBoysLauncher-mac-universal`
- **Architecture**: Universal binary (Intel + Apple Silicon)

**Build Command**:
```bash
go build -ldflags="-s -w -X main.version=${{ needs.bump_version.outputs.full_version }}" -o TheBoysLauncher-mac-universal .
```

### Build Flags and Optimization

All builds use consistent build flags for optimization:

```bash
-ldflags="-s -w -X main.version=${VERSION}"
```

- `-s`: Strip symbol table
- `-w`: Strip DWARF debugging information
- `-X main.version`: Inject version information at build time

### Artifact Management

#### Primary Artifacts
- **Windows**: `TheBoysLauncher.exe`
- **Linux**: `TheBoysLauncher-linux`
- **macOS**: `TheBoysLauncher-mac-universal`

#### Secondary Artifacts
- **Windows Installer**: `TheBoysLauncher-Setup-${VERSION}.msi`

#### Artifact Retention
- **Retention Period**: 14 days for dev prereleases
- **Storage**: GitHub Actions artifact storage
- **Cleanup**: Automatic deletion after retention period

---

## Version Management System

### Version Configuration

**File**: [`version.env`](../version.env)

**Format**:
```bash
# TheBoysLauncher Version Configuration (auto-generated)
VERSION=3.2.32
MAJOR=3
MINOR=2
PATCH=32
BUILD_METADATA=
PRERELEASE=dev.b05b4ce

# Full version string is constructed by scripts/get-version.sh
```

### Version Components

#### Base Version (MAJOR.MINOR.PATCH)
- **MAJOR**: Breaking changes (2.x.x → 3.x.x)
- **MINOR**: New features (3.1.x → 3.2.x)
- **PATCH**: Bug fixes (3.2.1 → 3.2.2)

#### Prerelease Identifiers
- **Dev Builds**: `dev.<short-sha>` (e.g., `dev.b05b4ce`)
- **Stable Releases**: Empty (no prerelease identifier)

#### Build Metadata
- Currently unused but available for future needs
- Format: `+metadata` (e.g., `+build.123`)

### Version Scripts

#### Version Retrieval
**Unix**: [`scripts/get-version.sh`](../scripts/get-version.sh)
**Windows**: [`scripts/get-version.ps1`](../scripts/get-version.ps1)

**Usage**:
```bash
# Get full version
./scripts/get-version.sh

# Get JSON format
./scripts/get-version.sh json

# Get Makefile format
./scripts/get-version.sh make

# Validate version format
./scripts/get-version.sh validate
```

#### Version Setting
**Unix**: [`scripts/set-version.sh`](../scripts/set-version.sh)
**Windows**: [`scripts/set-version.ps1`](../scripts/set-version.ps1)

**Usage**:
```bash
# Set new version
./scripts/set-version.sh 3.3.0

# Set version and update WiX
./scripts/set-version.sh 3.3.0 --update-inno
```

#### Version Validation
**File**: [`scripts/validate-version.sh`](../scripts/validate-version.sh)

**Validates**:
- Semantic version format compliance
- Consistency across project files
- WiX file version synchronization
- Build script version references

### Version Bumping Logic

#### Dev Branch Version Bump
1. **Increment PATCH**: `PATCH = PATCH + 1`
2. **Add Prerelease**: `PRERELEASE = "dev.<short-sha>"`
3. **Commit Changes**: Automatic commit with version bump
4. **Create Tag**: Annotated tag with full version

#### Stable Release Promotion
1. **Increment PATCH**: `PATCH = PATCH + 1`
2. **Clear Prerelease**: `PRERELEASE = ""`
3. **Commit Changes**: Commit with promotion message
4. **Create Tag**: Stable version tag

### Version Integration Points

#### Go Build Integration
```bash
go build -ldflags="-X main.version=${FULL_VERSION}" .
```

#### WiX Installer Integration
```xml
<Product Id="*" Version="3.2.0" ...>
```

#### Application Display
Version information is embedded in:
- Windows executable properties
- macOS app bundle Info.plist
- Application about dialog
- Update checking logic

---

## Release Process

### Dev Prerelease Process

#### Trigger
- **Automatic**: Every push to `dev` branch
- **Exclusion**: Skips commits by `github-actions[bot]`

#### Process Flow
1. **Version Bump**: Automated version increment and tagging
2. **Build Matrix**: Parallel builds for all platforms
3. **Release Creation**: GitHub prerelease with artifacts
4. **Changelog**: Auto-generated from git history

#### Release Characteristics
- **Type**: Prerelease
- **Naming**: `v<version>-dev.<sha>` (e.g., `v3.2.32-dev.b05b4ce`)
- **Artifacts**: All platform binaries and installers
- **Retention**: 14 days for artifacts, permanent for releases

### Stable Release Process

#### Automated PR-Based Release (Primary Method)

The primary method for creating stable releases is through Pull Requests from `dev` to `main`:

1. **PR Creation**:
   - Create a Pull Request from `dev` branch to `main` branch
   - Include detailed description of changes and testing results
   - Ensure all automated checks pass

2. **Review and Approval**:
   - Team members review the PR for code quality
   - Automated checks verify build compatibility
   - All issues must be resolved before merging

3. **Automated Release**:
   - Once the PR is merged to `main`, the stable-release.yml workflow triggers automatically
   - Version is bumped and tagged without manual intervention
   - All platform binaries are built and published as a stable release

#### Direct Merge Method (Alternative)

For maintainers who need to bypass the PR process:

1. **Direct Merge**:
   ```bash
   git checkout main
   git merge dev
   git push origin main
   ```

2. **Automated Release**:
   - The push to `main` automatically triggers the stable-release.yml workflow
   - Same automated process as PR-based releases

#### Legacy Manual Promotion (Deprecated)

Previously, stable releases required manual promotion using the `scripts/promote-release.sh` script. This method is now deprecated in favor of the automated PR-based process.

**File**: [`scripts/promote-release.sh`](../scripts/promote-release.sh)

**Process**:
1. **Validate Current State**: Ensure dev prerelease exists
2. **Increment Version**: Patch version increment for stable
3. **Clear Prerelease**: Remove dev prerelease identifier
4. **Commit and Tag**: Create stable release commit and tag
5. **Push Changes**: Update repository with new stable version

#### Release Characteristics
- **Type**: Stable release
- **Naming**: `v<version>` (e.g., `v3.3.0`)
- **Artifacts**: All platform binaries and installers
- **Distribution**: GitHub releases page
- **Trigger**: PR merge or direct push to `main` branch

### Version Progression

The version progression through the entire process follows this pattern:

1. **Development**: `3.2.32-dev.b05b4ce` (dev branch)
2. **Stable Release**: `3.2.33` (main branch after PR merge)

Each stable release increments the PATCH version from the previous dev build, ensuring a continuous version sequence.

### Release Artifacts

#### Primary Binaries
- **Windows**: `TheBoysLauncher.exe`
- **Linux**: `TheBoysLauncher-linux`
- **macOS**: `TheBoysLauncher-mac-universal`

#### Installers
- **Windows**: `TheBoysLauncher-Setup-<version>.msi`
- **macOS**: `TheBoysLauncher-Universal.dmg` (when created)

#### File Naming Convention
```
TheBoysLauncher-<platform>[-<version>][.<extension>]
```

Examples:
- `TheBoysLauncher-Windows`
- `TheBoysLauncher-Linux`
- `TheBoysLauncher-macOS`
- `TheBoysLauncher-Setup-3.2.32-dev.b05b4ce.msi`

### Release Changelog

#### Automatic Generation
Changelog is automatically generated from git history:

```bash
# Between tags
git log --pretty=format:'%h %s (%an)' ${PREV_TAG}..${TAG}

# For first release
git log --pretty=format:'%h %s (%an)' -n 50
```

#### Format
```
<short-sha> <commit message> (<author name>)
```

Example:
```
b05b4ce feat: add new feature (John Doe)
a1b2c3d fix: resolve crash issue (Jane Smith)
```

---

## Testing and Quality Assurance

### Built-in Testing

#### Build Verification
**Makefile Target**: `make verify`

**Process**:
1. **Compilation Test**: Verifies code compiles without errors
2. **Linking Test**: Ensures all dependencies are properly linked
3. **Basic Validation**: Confirms executable is created

#### Code Quality
**Makefile Target**: `make lint`

**Checks**:
1. **Go Formatting**: `go fmt ./...`
2. **Go Vet**: `go vet ./...`
3. **Code Style**: Consistent formatting and style

#### Runtime Testing
**Makefile Target**: `make test-runtime`

**Process**:
1. **Execution Test**: Runs built executable briefly
2. **Panic Detection**: Checks for immediate crashes
3. **Basic Functionality**: Verifies application starts

### Cross-Platform Testing

#### Test Scripts
- **Unix**: [`test_workflow.sh`](../test_workflow.sh)
- **Windows**: [`test_workflow.ps1`](../test_workflow.ps1)

**Test Coverage**:
- Platform detection
- Configuration loading
- Java API integration
- Archive extraction
- File operations
- Memory detection
- Executable naming
- PATH handling
- Update system
- Script availability
- Dependencies

#### Unit Tests
**Command**: `go test -v ./...`

**Test Files**:
- [`tests/platform_test.go`](../tests/platform_test.go)
- [`tests/devbuilds_test.go`](../tests/devbuilds_test.go)
- [`tests/gui_test.go`](../tests/gui_test.go)
- [`tests/pagination_test.go`](../tests/pagination_test.go)

### CI/CD Testing Integration

#### Workflow Testing
The CI/CD pipeline includes multiple testing stages:

1. **Pre-build Testing**:
   - Go module validation
   - Dependency verification

2. **Build Testing**:
   - Compilation success
   - Linking verification
   - Artifact creation

3. **Post-build Testing**:
   - Artifact validation
   - File integrity checks

#### Quality Gates

#### Version Validation
**Script**: [`scripts/validate-version.sh`](../scripts/validate-version.sh)

**Validations**:
- Semantic version format
- Cross-file consistency
- WiX synchronization
- Build script references

#### Build Verification
**Scripts**:
- [`tools/verify-build.bat`](../tools/verify-build.bat) (Windows)
- [`tools/verify-build.sh`](../tools/verify-build.sh) (Unix)

**Checks**:
- Binary existence
- File permissions
- Basic functionality

### Test Reports

#### Documentation
Test results and reports are documented in:
- [`docs/TESTING_REPORT.md`](../docs/TESTING_REPORT.md)
- [`docs/DEV_BUILDS_TEST_REPORT.md`](../docs/DEV_BUILDS_TEST_REPORT.md)
- Various feature-specific test reports

#### Manual Testing
Manual testing procedures are documented for:
- Development builds
- Release validation
- Platform-specific functionality

---

## Infrastructure and Dependencies

### Development Environment

#### Go Toolchain
- **Version**: Go 1.23+ (specified in [`go.mod`](../go.mod))
- **Modules**: Go modules for dependency management
- **CGO**: Required for Fyne GUI framework

#### Platform Dependencies

##### Windows
- **PowerShell 5.1+**: For build scripts
- **WiX Toolset**: For MSI installer creation
- **MinGW-w64**: Optional for CGO dependencies

##### macOS
- **Xcode Command Line Tools**: For CGO compilation
- **Homebrew**: Recommended for tool management
- **create-dmg**: Optional for styled DMG creation

##### Linux
- **GCC/Build Essential**: For CGO compilation
- **System Libraries**: OpenGL, X11, ALSA for GUI support

### Build Dependencies

#### Go Modules
**Primary Dependencies**:
- `fyne.io/fyne/v2 v2.7.0` - GUI framework
- `github.com/BurntSushi/toml v1.5.0` - TOML parsing
- `golang.org/x/sys v0.36.0` - System calls

#### Build Tools
- **rsrc**: Windows resource compilation
- **WiX Toolset**: Windows installer creation
- **lipo**: macOS universal binary creation
- **iconutil**: macOS icon conversion

### CI/CD Infrastructure

#### GitHub Actions
- **Runners**: Ubuntu, Windows, macOS latest
- **Concurrency**: Controlled to prevent conflicts
- **Permissions**: Content write for releases and commits

#### Artifact Storage
- **GitHub Releases**: Primary distribution
- **GitHub Actions Artifacts**: Temporary storage (14 days)
- **Repository**: Source code and documentation

### External Services

#### Version Control
- **GitHub**: Source code hosting and CI/CD
- **Git Tags**: Version marking and release management

#### Distribution
- **GitHub Releases**: Primary distribution channel
- **Direct Downloads**: Binary distribution

### Development Tools

#### Build System
- **Make**: Cross-platform build automation
- **PowerShell/Bash**: Platform-specific scripting
- **Go Build**: Native Go compilation

#### Packaging Tools
- **WiX**: Windows MSI creation
- **DMG Tools**: macOS disk image creation
- **App Bundle**: macOS application packaging

---

## Key Files and Their Roles

### Core Configuration Files

#### [`version.env`](../version.env)
**Role**: Centralized version configuration
**Contents**: Semantic version components and prerelease identifiers
**Maintenance**: Auto-generated by CI/CD, should not be manually edited

#### [`go.mod`](../go.mod)
**Role**: Go module definition and dependency management
**Contents**: Module requirements and Go version specification
**Maintenance**: Updated with `go mod tidy` when dependencies change

#### [`Makefile`](../Makefile)
**Role**: Cross-platform build automation
**Contents**: Build targets for all platforms and development tasks
**Maintenance**: Extended for new build targets or platforms

### CI/CD Configuration

#### [`.github/workflows/dev-prerelease.yml`](../.github/workflows/dev-prerelease.yml)
**Role**: Automated dev build and release workflow
**Contents**: Complete CI/CD pipeline definition for dev branch
**Maintenance**: Updated for new build steps or platform support

#### [`.github/workflows/stable-release.yml`](../.github/workflows/stable-release.yml)
**Role**: Automated stable release workflow triggered by PR merges
**Contents**: Complete CI/CD pipeline definition for main branch
**Maintenance**: Updated for new build steps or platform support

#### [`.github/copilot-instructions.md`](../.github/copilot-instructions.md)
**Role**: AI assistant configuration
**Contents**: Project-specific instructions for GitHub Copilot
**Maintenance**: Updated with new development guidelines

### Build Scripts

#### Version Management
- [`scripts/get-version.sh`](../scripts/get-version.sh) - Unix version retrieval
- [`scripts/get-version.ps1`](../scripts/get-version.ps1) - Windows version retrieval
- [`scripts/set-version.sh`](../scripts/set-version.sh) - Unix version setting
- [`scripts/set-version.ps1`](../scripts/set-version.ps1) - Windows version setting
- [`scripts/validate-version.sh`](../scripts/validate-version.sh) - Version validation
- [`scripts/promote-release.sh`](../scripts/promote-release.sh) - Release promotion

#### Platform Building
- [`scripts/build-installer.ps1`](../scripts/build-installer.ps1) - Windows MSI creation
- [`scripts/create-app-bundle.sh`](../scripts/create-app-bundle.sh) - macOS app bundle
- [`scripts/create-dmg.sh`](../scripts/create-dmg.sh) - macOS DMG creation
- [`scripts/convert-icon.sh`](../scripts/convert-icon.sh) - Icon format conversion

#### Development Tools
- [`tools/build.ps1`](../tools/build.ps1) - Windows build with resources
- [`tools/build.bat`](../tools/build.bat) - Windows batch build
- [`tools/verify-build.ps1`](../tools/verify-build.ps1) - Windows build verification
- [`tools/verify-build.sh`](../tools/verify-build.sh) - Unix build verification

### Platform Configuration

#### WiX Installer
- [`wix/TheBoysLauncher.wxs`](../wix/TheBoysLauncher.wxs) - Main installer configuration
- [`wix/Product.wxs`](../wix/Product.wxs) - Product definition
- [`wix/TheBoysLauncher.wixproj`](../wix/TheBoysLauncher.wixproj) - WiX project file

#### Resources
- [`resources/`](../resources/) - Platform-specific resources
  - `windows/` - Windows-specific files
  - `darwin/` - macOS-specific files
  - `common/` - Shared resources

#### Configuration
- [`config/modpacks.json`](../config/modpacks.json) - Modpack configurations
- [`config/openssl.cnf`](../config/openssl.cnf) - OpenSSL configuration

### Application Source

#### Core Files
- [`main.go`](../main.go) - Application entry point
- [`launcher.go`](../launcher.go) - Main launcher logic
- [`gui.go`](../gui.go) - GUI interface
- [`config.go`](../config.go) - Configuration management

#### Platform-Specific
- [`platform_windows.go`](../platform_windows.go) - Windows implementation
- [`platform_darwin.go`](../platform_darwin.go) - macOS implementation
- [`platform_linux.go`](../platform_linux.go) - Linux implementation

#### Feature Modules
- [`java.go`](../java.go) - Java runtime management
- [`prism.go`](../prism.go) - Prism Launcher integration
- [`download.go`](../download.go) - Download utilities
- [`update.go`](../update.go) - Update mechanism

### Testing

#### Test Files
- [`tests/`](../tests/) - Unit and integration tests
- [`test_workflow.sh`](../test_workflow.sh) - Unix workflow tests
- [`test_workflow.ps1`](../test_workflow.ps1) - Windows workflow tests

#### Documentation
- [`docs/`](../docs/) - Comprehensive documentation
- [`README.md`](../README.md) - Project overview and quick start

---

## Workflow Diagram

```
┌─────────────────┐    Push to dev    ┌──────────────────┐
│   Developer    │ ──────────────────►│   GitHub Repo    │
└─────────────────┘                   └──────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ Dev Prerelease      │
                                   │ GitHub Actions      │
                                   │ Workflow            │
                                   └─────────────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
        ┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
        │   Version Bump     │   │   Build Matrix     │   │   Release Publish   │
        │   Job              │   │   Job              │   │   Job              │
        └─────────────────────┘   └─────────────────────┘   └─────────────────────┘
                    │                         │                         │
                    ▼                         ▼                         ▼
        ┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
        │ • Increment PATCH  │   │ • Windows Build    │   │ • Download Artifacts│
        │ • Add dev prerelease│   │ • Linux Build      │   │ • Generate Changelog│
        │ • Commit version    │   │ • macOS Build      │   │ • Create Release    │
        │ • Create tag        │   │ • Create MSI       │   │ • Upload Assets     │
        └─────────────────────┘   └─────────────────────┘   └─────────────────────┘
                    │                         │                         │
                    └─────────────────────────┼─────────────────────────┘
                                              ▼
                                   ┌─────────────────────┐
                                   │ GitHub Prerelease   │
                                   │ v3.2.32-dev.b05b4ce │
                                   └─────────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │   Users Download   │
                                   │   and Test         │
                                   └─────────────────────┘
                                              │
                                    Ready for stable release?
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ Create PR dev→main  │
                                   │ (GitHub UI)         │
                                   └─────────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │   PR Review &      │
                                   │   Approval         │
                                   └─────────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │   Merge PR to      │
                                   │   main branch      │
                                   └─────────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ Stable Release      │
                                   │ GitHub Actions      │
                                   │ Workflow            │
                                   └─────────────────────┘
                                              │
                    ┌─────────────────────────┼─────────────────────────┐
                    ▼                         ▼                         ▼
        ┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
        │   Version Bump     │   │   Build Matrix     │   │   Release Publish   │
        │   Job              │   │   Job              │   │   Job              │
        └─────────────────────┘   └─────────────────────┘   └─────────────────────┘
                    │                         │                         │
                    ▼                         ▼                         ▼
        ┌─────────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
        │ • Increment PATCH  │   │ • Windows Build    │   │ • Download Artifacts│
        │ • Clear prerelease │   │ • Linux Build      │   │ • Generate Changelog│
        │ • Commit version    │   │ • macOS Build      │   │ • Create Stable     │
        │ • Create tag        │   │ • Create MSI       │   │   Release           │
        └─────────────────────┘   └─────────────────────┘   │ • Upload Assets     │
                    │                         │                         └─────────────────────┘
                    └─────────────────────────┼─────────────────────────┘
                                              ▼
                                   ┌─────────────────────┐
                                   │ GitHub Stable       │
                                   │ Release v3.2.33    │
                                   └─────────────────────┘
```

### Alternative Direct Merge Path

For maintainers who need to bypass the PR process:

```
┌─────────────────┐    Push to dev    ┌──────────────────┐
│   Developer    │ ──────────────────►│   GitHub Repo    │
└─────────────────┘                   └──────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ Dev Prerelease      │
                                   │ GitHub Actions      │
                                   │ Workflow            │
                                   └─────────────────────┘
                                              │
                                    (Testing complete)
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ Direct Merge        │
                                   │ dev → main         │
                                   │ (Maintainer only)   │
                                   └─────────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ Stable Release      │
                                   │ GitHub Actions      │
                                   │ Workflow            │
                                   └─────────────────────┘
                                              │
                                              ▼
                                   ┌─────────────────────┐
                                   │ GitHub Stable       │
                                   │ Release v3.2.33    │
                                   └─────────────────────┘
```

### Workflow States

#### Development State
- **Branch**: `dev`
- **Version**: `3.2.32-dev.b05b4ce`
- **Release Type**: Prerelease
- **Frequency**: Every push to dev

#### Stable State
- **Branch**: `main`
- **Version**: `3.2.33`
- **Release Type**: Stable
- **Frequency**: PR merge or direct push to main

### Key Decision Points

1. **Automatic Dev Release**: Every push to `dev` triggers automated release
2. **PR-Based Stable Release**: PR from dev to main triggers automated stable release
3. **Direct Stable Release**: Maintainers can directly merge dev to main
4. **Version Increment**: Automatically handled in both dev and stable releases
5. **Quality Gates**: Build verification and testing at each stage
6. **Code Review**: PR process ensures stable releases are properly reviewed

---

## Conclusion

The TheBoysLauncher GitHub workflow provides a comprehensive, automated system for managing the entire software development lifecycle. By combining automated dev releases with PR-based stable release automation, the project ensures continuous delivery while maintaining release quality through code review.

### Key Benefits

1. **Automation**: Reduces manual effort and human error
2. **Code Review**: PR process ensures stable releases are properly reviewed
3. **Consistency**: Standardized process across all platforms
4. **Traceability**: Complete git history and version tracking
5. **Quality**: Integrated testing and validation
6. **Flexibility**: Supports both rapid development and stable releases

### Future Enhancements

Potential improvements to the workflow:
1. **Automated Testing**: Extended automated integration and end-to-end testing
2. **Distribution**: Additional distribution channels
3. **Monitoring**: Release health and usage monitoring
4. **Rollback**: Automated rollback mechanisms for failed releases
5. **Release Criteria**: Automated criteria-based PR approval

This workflow serves as a robust foundation for TheBoysLauncher development and can be extended as the project grows and evolves.