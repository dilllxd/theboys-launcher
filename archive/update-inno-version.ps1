# Update InnoSetup Version Script
# This script updates the InnoSetup .iss file with version from version.env

param(
    [string]$IssFile = "config/TheBoysLauncher.iss",
    [string]$VersionFile = "version.env"
)

# Read version from version.env
if (-not (Test-Path $VersionFile)) {
    Write-Error "Version file not found: $VersionFile"
    exit 1
}

$VersionContent = Get-Content $VersionFile
$Version = ($VersionContent | Where-Object { $_ -match '^VERSION=' }) -replace '^VERSION=', ''

if (-not $Version) {
    Write-Error "VERSION not found in $VersionFile"
    exit 1
}

Write-Host "Updating InnoSetup version to: $Version"

# Read the .iss file
if (-not (Test-Path $IssFile)) {
    Write-Error "InnoSetup file not found: $IssFile"
    exit 1
}

$IssContent = Get-Content $IssFile

# Update the version line
$UpdatedContent = $IssContent | ForEach-Object {
    if ($_ -match '^#define MyAppVersion') {
        "#define MyAppVersion `"$Version`""
    } else {
        $_
    }
}

# Write back to file
$UpdatedContent | Set-Content $IssFile -Encoding UTF8

Write-Host "✅ Updated $IssFile with version $Version"