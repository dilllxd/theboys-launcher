param(
    [string]$Version = "v2.0.3"
)

# Update version in versioninfo.json dynamically
Write-Host "Updating version info to $Version..."

$versionClean = $Version -replace '^v', ''
$parts = $versionClean -split '\.'
$versionMajor = [int]$parts[0]
$versionMinor = [int]$parts[1]
$versionPatch = [int]$parts[2]

# Read and update the versioninfo.json file
$jsonContent = Get-Content "versioninfo.json" | ConvertFrom-Json

# Update FixedFileInfo version numbers
$jsonContent.FixedFileInfo.FileVersion.Major = $versionMajor
$jsonContent.FixedFileInfo.FileVersion.Minor = $versionMinor
$jsonContent.FixedFileInfo.FileVersion.Patch = $versionPatch
$jsonContent.FixedFileInfo.ProductVersion.Major = $versionMajor
$jsonContent.FixedFileInfo.ProductVersion.Minor = $versionMinor
$jsonContent.FixedFileInfo.ProductVersion.Patch = $versionPatch

# Update StringFileInfo version strings
$jsonContent.StringFileInfo.FileVersion = $versionClean
$jsonContent.StringFileInfo.ProductVersion = $versionClean

# Write back to JSON file without BOM
$jsonString = $jsonContent | ConvertTo-Json -Depth 10
$utf8NoBom = New-Object System.Text.UTF8Encoding $false
[System.IO.File]::WriteAllText("versioninfo.json", $jsonString, $utf8NoBom)

Write-Host "Version updated to $versionClean ($versionMajor,$versionMinor,$versionPatch,0)"