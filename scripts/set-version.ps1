# Set Version PowerShell Script for TheBoys Launcher
# This script updates the version in version.env and optionally updates related files

param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$NewVersion,

    [switch]$UpdateInno,

    [switch]$Help
)

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
}

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Usage information
function Show-Usage {
    Write-ColorOutput "Usage: .\set-version.ps1 <major.minor.patch> [-UpdateInno]" "Blue"
    Write-Host ""
    Write-ColorOutput "Examples:" "Blue"
    Write-Host "  .\set-version.ps1 3.2.1              # Update version to 3.2.1"
    Write-Host "  .\set-version.ps1 3.3.0 -UpdateInno # Update version and InnoSetup file"
    Write-Host ""
    Write-ColorOutput "Options:" "Blue"
    Write-Host "  -UpdateInno     Also update the InnoSetup .iss file"
    Write-Host "  -Help, -h       Show this help message"
    exit 1
}

if ($Help) {
    Show-Usage
}

# Validate version format
if ($NewVersion -notmatch '^\d+\.\d+\.\d+$') {
    Write-ColorOutput "❌ Invalid version format: $NewVersion" $Colors.Red
    Write-ColorOutput "Expected format: major.minor.patch (e.g., 3.2.1)" $Colors.Yellow
    exit 1
}

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$VersionFile = Join-Path $ProjectRoot "version.env"

# Check if version.env exists
if (-not (Test-Path $VersionFile)) {
    Write-ColorOutput "❌ Version file not found: $VersionFile" $Colors.Red
    exit 1
}

# Extract current version
$CurrentVersion = (Get-Content $VersionFile | Where-Object { $_ -match '^VERSION=' }) -replace '^VERSION=', ''

Write-ColorOutput "🔄 Updating version..." $Colors.Blue
Write-Host "   Current version: " -NoNewline; Write-ColorOutput $CurrentVersion $Colors.Yellow
Write-Host "   New version: " -NoNewline; Write-ColorOutput $NewVersion $Colors.Green
Write-Host ""

# Backup current version.env
$BackupFile = "$VersionFile.backup"
Copy-Item $VersionFile $BackupFile
Write-ColorOutput "💾 Backed up current version.env to version.env.backup" $Colors.Blue

# Extract version components
$Parts = $NewVersion -split '\.'
$Major = $Parts[0]
$Minor = $Parts[1]
$Patch = $Parts[2]

# Update version.env
$VersionContent = Get-Content $VersionFile
$UpdatedContent = $VersionContent | ForEach-Object {
    if ($_ -match '^VERSION=') {
        "VERSION=$NewVersion"
    } elseif ($_ -match '^MAJOR=') {
        "MAJOR=$Major"
    } elseif ($_ -match '^MINOR=') {
        "MINOR=$Minor"
    } elseif ($_ -match '^PATCH=') {
        "PATCH=$Patch"
    } else {
        $_
    }
}

$UpdatedContent | Set-Content $VersionFile -Encoding UTF8
Write-ColorOutput "✅ Updated version.env" $Colors.Green

# Update InnoSetup file if requested
if ($UpdateInno) {
    $IssFile = Join-Path $ProjectRoot "config\TheBoysLauncher.iss"
    if (Test-Path $IssFile) {
        $IssContent = Get-Content $IssFile
        $UpdatedIssContent = $IssContent | ForEach-Object {
            if ($_ -match '^#define MyAppVersion') {
                "#define MyAppVersion `"$NewVersion`"  ; This will be updated by update-inno-version.ps1"
            } else {
                $_
            }
        }
        $UpdatedIssContent | Set-Content $IssFile -Encoding UTF8
        Write-ColorOutput "✅ Updated config\TheBoysLauncher.iss" $Colors.Green
    } else {
        Write-ColorOutput "⚠️  InnoSetup file not found: $IssFile" $Colors.Yellow
    }
}

# Run validation if available
$ValidationScript = Join-Path $ScriptDir "validate-version.sh"
if (Test-Path $ValidationScript) {
    Write-Host ""
    Write-ColorOutput "🔍 Running version validation..." $Colors.Blue
    try {
        # Try to run bash script (if Git Bash or WSL is available)
        bash $ValidationScript
    } catch {
        Write-ColorOutput "⚠️  Could not run validation script (bash not available)" $Colors.Yellow
        Write-ColorOutput "   Manual validation recommended" $Colors.Yellow
    }
}

Write-Host ""
Write-ColorOutput "🎉 Version update completed successfully!" $Colors.Green
Write-ColorOutput "💡 Next steps:" $Colors.Blue
Write-Host "   1. Review the changes with: git diff"
Write-Host "   2. Commit the changes: git add . && git commit -m `"chore: bump version to $NewVersion`""
Write-Host "   3. Create a tag: git tag $NewVersion"
Write-Host "   4. Push changes and tag: git push && git push --tags"