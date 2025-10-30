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
$issPath = Join-Path $projectDir 'TheBoysLauncher.iss'

Write-Host "Resolved ProjectDir: $projectDir"
Write-Host "Resolved EXE path: $exePath"
Write-Host "Resolved Inno Setup script path: $issPath"

if (-not (Test-Path $exePath)) {
    Write-Error "Built exe not found at $exePath"
    exit 2
}

if (-not (Test-Path $issPath)) {
    Write-Error "Inno Setup script not found at $issPath"
    exit 3
}

# Check for Inno Setup compiler (ISCC.exe)
if (-not (Get-Command ISCC.exe -ErrorAction SilentlyContinue)) {
    Write-Error "ISCC.exe not found in PATH. Install Inno Setup and add its directory to PATH."
    Write-Host ""
    Write-Host "To install Inno Setup:" -ForegroundColor Yellow
    Write-Host "1. Download from: https://jrsoftware.org/isdl.php" -ForegroundColor Cyan
    Write-Host "2. Run the installer with default settings" -ForegroundColor Cyan
    Write-Host "3. Ensure ISCC.exe is in your PATH (usually added automatically)" -ForegroundColor Cyan
    Write-Host ""
    exit 4
}

# Clean up version (strip leading 'v' if present)
$cleanVersion = $Version.Trim()
if ($cleanVersion.StartsWith('v')) { $cleanVersion = $cleanVersion.Substring(1) }

Write-Host "Using ISCC.exe to compile Inno Setup installer"
Write-Host "Clean version: $cleanVersion"

# Update version in the .iss file
Write-Host "Updating version in Inno Setup script to: $cleanVersion"
$issContent = Get-Content $issPath
$updatedContent = $issContent | ForEach-Object {
    if ($_ -match '^#define MyAppVersion') {
        "#define MyAppVersion `"$cleanVersion`""
    } else {
        $_
    }
}
$updatedContent | Set-Content $issPath -Encoding UTF8

# Build the installer
Write-Host "Running ISCC.exe to compile installer..."
$installerDir = Join-Path $projectDir 'installer'
New-Item -ItemType Directory -Path $installerDir -Force | Out-Null

$compilerArgs = @(
    "/Q",  # Quiet mode (less output)
    "/O`"$installerDir`"",  # Output directory
    "/F`"TheBoysLauncher-Setup-$cleanVersion`"",  # Output filename
    "/DMyAppVersion=`"$cleanVersion`"",  # Define version
    "`"$issPath`""
)

$process = Start-Process -FilePath "ISCC.exe" -ArgumentList $compilerArgs -Wait -PassThru -NoNewWindow

if ($process.ExitCode -ne 0) { 
    Write-Error "ISCC.exe failed with exit code $($process.ExitCode)"
    exit $process.ExitCode 
}

$installerPath = Join-Path $installerDir "TheBoysLauncher-Setup-$cleanVersion.exe"

if (Test-Path $installerPath) {
    Write-Host "Installer created: $installerPath"
    exit 0
} else {
    Write-Error "Installer was not created at expected location: $installerPath"
    exit 5
}
