param(
    [string]$Version = "v0.0.0"
)

Write-Host "Building installer for version $Version"

# Run the existing build script to embed icon and version into the exe
& .\tools\build.ps1 -Version $Version
if ($LASTEXITCODE -ne 0) { Write-Error "tools\\build.ps1 failed"; exit $LASTEXITCODE }

# Resolve paths
$projectDir = (Get-Location).Path
# Ensure ProjectDir ends with a backslash so WiX "$(var.ProjectDir)" + "icon.ico" resolves to a valid path
if (-not $projectDir.EndsWith('\')) { $projectDir = $projectDir + '\' }
$exePath = Join-Path $projectDir 'TheBoysLauncher.exe'

Write-Host "Resolved ProjectDir: $projectDir"
Write-Host "Resolved EXE path: $exePath"
Write-Host "Resolved icon path: ${projectDir}icon.ico"
Write-Host "Resolved license path: ${projectDir}wix\LICENSE.rtf"

if (-not (Test-Path $exePath)) {
    Write-Error "Built exe not found at $exePath"
    exit 2
}

Write-Host "Using candle.exe to compile wix/Product.wxs"

# Build arguments safely as an array
# Clean up version for product metadata (strip leading 'v' if present)
$cleanVersion = $Version.Trim()
if ($cleanVersion.StartsWith('v')) { $cleanVersion = $cleanVersion.Substring(1) }
Write-Host "Clean version: $cleanVersion"

# Pass ProductVersion to candle so Product/@Version uses $(var.ProductVersion)
$candleArgs = @(
    "-dProjectDir=$projectDir",
    "-dTheBoysLauncher.TargetPath=$exePath",
    "-dProductVersion=$cleanVersion",
    "wix\\Product.wxs",
    "-out",
    "installer.wixobj"
)

& candle.exe @candleArgs
if ($LASTEXITCODE -ne 0) { Write-Error "candle.exe failed"; exit $LASTEXITCODE }

Write-Host "Running light.exe to link the MSI"
& light.exe installer.wixobj -ext WixUIExtension -out "TheBoysLauncher-Setup-$cleanVersion.msi" -sval
if ($LASTEXITCODE -ne 0) { Write-Error "light.exe failed"; exit $LASTEXITCODE }

Write-Host "MSI created: TheBoysLauncher-Setup-$cleanVersion.msi"
Remove-Item installer.wixobj -ErrorAction SilentlyContinue

exit 0
