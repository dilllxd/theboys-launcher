param(
    [string]$Version = "v0.0.0"
)

Write-Host "Building installer for version $Version"

# Run the existing build script to embed icon and version into the exe
& .\tools\build.ps1 -Version $Version
if ($LASTEXITCODE -ne 0) { Write-Error "tools\\build.ps1 failed"; exit $LASTEXITCODE }

# Resolve paths
$projectDir = (Get-Location).Path
$exePath = Join-Path $projectDir 'TheBoysLauncher.exe'

if (-not (Test-Path $exePath)) {
    Write-Error "Built exe not found at $exePath"
    exit 2
}

Write-Host "Using candle.exe to compile wix/Product.wxs"

# Build arguments safely as an array
$candleArgs = @(
    "-dProjectDir=$projectDir",
    "-dTheBoysLauncher.TargetPath=$exePath",
    "wix\\Product.wxs",
    "-out",
    "installer.wixobj"
)

& candle.exe @candleArgs
if ($LASTEXITCODE -ne 0) { Write-Error "candle.exe failed"; exit $LASTEXITCODE }

Write-Host "Running light.exe to link the MSI"
& light.exe installer.wixobj -ext WixUIExtension -out "TheBoysLauncher-Setup-$Version.msi" -sval
if ($LASTEXITCODE -ne 0) { Write-Error "light.exe failed"; exit $LASTEXITCODE }

Write-Host "MSI created: TheBoysLauncher-Setup-$Version.msi"
Remove-Item installer.wixobj -ErrorAction SilentlyContinue

exit 0
