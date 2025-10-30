Param(
    [string]$Version = "v3.0.1",
    [string]$ProjectDir = (Get-Location).Path,
    [string]$OutputDir = "$(Join-Path -Path $ProjectDir -ChildPath 'build')"
)

Write-Host "Building WiX MSI for TheBoysLauncher..." -ForegroundColor Cyan

# Simple checks
if (-not (Get-Command candle.exe -ErrorAction SilentlyContinue)) {
    Write-Error "candle.exe not found in PATH. Install WiX Toolset and add its bin to PATH."
    exit 1
}
if (-not (Get-Command light.exe -ErrorAction SilentlyContinue)) {
    Write-Error "light.exe not found in PATH. Install WiX Toolset and add its bin to PATH."
    exit 1
}

$wxsPath = Join-Path $ProjectDir "wix\TheBoysLauncher.wxs"
$targetExe = Join-Path $ProjectDir "build\windows\TheBoysLauncher.exe"

if (-not (Test-Path $wxsPath)) {
    Write-Error "Cannot find $wxsPath"
    exit 1
}
if (-not (Test-Path $targetExe)) {
    Write-Error "The target EXE was not found at $targetExe. Build the exe first (make build-windows or tools/build.ps1)."
    exit 1
}

New-Item -ItemType Directory -Path (Join-Path $ProjectDir 'wixobj') -Force | Out-Null
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

$wxsPath = Join-Path $ProjectDir "wix\TheBoysLauncher.wxs"
$customUIPath = Join-Path $ProjectDir "wix\CustomUI.wxs"
$wixObj = Join-Path $ProjectDir 'wixobj\TheBoysLauncher.wixobj'
$customUIObj = Join-Path $ProjectDir 'wixobj\CustomUI.wixobj'
$msiOut = Join-Path $OutputDir 'TheBoysLauncher.msi'

Write-Host "Running candle.exe for main wxs..." -ForegroundColor Yellow
candle.exe `
    -ext WixUIExtension `
    -dTheBoysLauncher.TargetPath="$targetExe" `
    -dProjectDir="$ProjectDir\" `
    -out $wixObj `
    $wxsPath

if ($LASTEXITCODE -ne 0) { Write-Error "candle.exe failed for main wxs"; exit $LASTEXITCODE }

Write-Host "Running candle.exe for custom UI..." -ForegroundColor Yellow
candle.exe `
    -ext WixUIExtension `
    -out $customUIObj `
    $customUIPath

if ($LASTEXITCODE -ne 0) { Write-Error "candle.exe failed for custom UI"; exit $LASTEXITCODE }

Write-Host "Running light.exe..." -ForegroundColor Yellow
light.exe $wixObj $customUIObj -ext WixUIExtension -cultures:"en-us" -out $msiOut

if ($LASTEXITCODE -ne 0) { Write-Error "light.exe failed"; exit $LASTEXITCODE }

Write-Host "MSI built: $msiOut" -ForegroundColor Green
Write-Host "Notes: Run the MSI to verify the license (RTF), icon and the feature selection (shortcuts)." -ForegroundColor Cyan
