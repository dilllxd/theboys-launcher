# Get-Version PowerShell Script for TheBoys Launcher
# This script reads version information and exports it in various formats

param(
    [ValidateSet("version", "json", "export", "validate")]
    [string]$Format = "version"
)

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

# Read version configuration
$VersionFile = Join-Path $ProjectRoot "version.env"
if (-not (Test-Path $VersionFile)) {
    Write-Error "Error: version.env file not found at $VersionFile"
    exit 1
}

# Parse version.env file
$VersionConfig = @{}
Get-Content $VersionFile | ForEach-Object {
    if ($_ -match '^([^#].+?)=(.*)$') {
        $VersionConfig[$matches[1].Trim()] = $matches[2].Trim()
    }
}

# Extract version variables
$Version = $VersionConfig["VERSION"]
$Major = $VersionConfig["MAJOR"]
$Minor = $VersionConfig["MINOR"]
$Patch = $VersionConfig["PATCH"]
$Prerelease = $VersionConfig["PRERELEASE"]
$BuildMetadata = $VersionConfig["BUILD_METADATA"]

# Construct full version string
$FullVersion = $Version
if ($Prerelease) {
    $FullVersion = "$FullVersion-$Prerelease"
}
if ($BuildMetadata) {
    $FullVersion = "$FullVersion+$BuildMetadata"
}

# Output functions
function Show-Version {
    param([string]$VersionString = $FullVersion)
    return $VersionString
}

function Show-VersionJson {
    return @{
        version = $Version
        major = [int]$Major
        minor = [int]$Minor
        patch = [int]$Patch
        prerelease = $Prerelease
        build_metadata = $BuildMetadata
        full_version = $FullVersion
    } | ConvertTo-Json
}

function Show-VersionExport {
    $export = @"
export VERSION="$Version"
export MAJOR="$Major"
export MINOR="$Minor"
export PATCH="$Patch"
export PRERELEASE="$Prerelease"
export BUILD_METADATA="$BuildMetadata"
export FULL_VERSION="$FullVersion"
"@
    return $export
}

function Test-VersionFormat {
    if ($Version -notmatch '^\d+\.\d+\.\d+$') {
        Write-Error "Error: Invalid version format: $Version"
        exit 1
    }
    return "✅ Version format is valid: $Version"
}

# Main execution
switch ($Format) {
    "version" {
        Show-Version
    }
    "json" {
        Show-VersionJson
    }
    "export" {
        Show-VersionExport
    }
    "validate" {
        Test-VersionFormat
    }
    default {
        Write-Error "Invalid format: $Format"
        exit 1
    }
}

# Export variables for use in other scripts
$env:VERSION = $Version
$env:MAJOR = $Major
$env:MINOR = $Minor
$env:PATCH = $Patch
$env:PRERELEASE = $Prerelease
$env:BUILD_METADATA = $BuildMetadata
$env:FULL_VERSION = $FullVersion