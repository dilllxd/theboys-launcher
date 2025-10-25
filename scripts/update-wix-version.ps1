# Update WiX Version Script for TheBoysLauncher
# This script updates the WiX .wxs file with version from version.env

param(
    [string]$WxsFile = "wix/TheBoysLauncher.wxs",
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

Write-Host "Updating WiX version to: $Version"

# Read the .wxs file
if (-not (Test-Path $WxsFile)) {
    Write-Error "WiX file not found: $WxsFile"
    exit 1
}

$WxsContent = Get-Content $WxsFile

# Update the version in Product element
$UpdatedContent = $WxsContent | ForEach-Object {
    if ($_ -match 'Version="[^"]*"') {
        $_ -replace 'Version="[^"]*"', "Version=`"$Version`""
    } else {
        $_
    }
}

# Also update the version registry value
$UpdatedContent = $UpdatedContent | ForEach-Object {
    if ($_ -match '<RegistryValue Type="string" Name="Version" Value="[^"]*"') {
        $_ -replace 'Value="[^"]*"', "Value=`"$Version`""
    } else {
        $_
    }
}

# Write back to file
$UpdatedContent | Set-Content $WxsFile -Encoding UTF8

Write-Host "✅ Updated $WxsFile with version $Version"