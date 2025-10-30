Param(
    [string]$Version = "v3.2.68",
    [string]$ProjectDir = (Get-Location).Path,
    [string]$OutputDir = "$(Join-Path -Path $ProjectDir -ChildPath 'installer')"
)

Write-Host "Building Inno Setup installer for TheBoysLauncher..." -ForegroundColor Cyan

# Check for Inno Setup compiler (ISCC.exe)
if (-not (Get-Command ISCC.exe -ErrorAction SilentlyContinue)) {
    Write-Error "ISCC.exe not found in PATH. Install Inno Setup and add its directory to PATH."
    Write-Host ""
    Write-Host "To install Inno Setup:" -ForegroundColor Yellow
    Write-Host "1. Download from: https://jrsoftware.org/isdl.php" -ForegroundColor Cyan
    Write-Host "2. Run the installer with default settings" -ForegroundColor Cyan
    Write-Host "3. Ensure ISCC.exe is in your PATH (usually added automatically)" -ForegroundColor Cyan
    Write-Host ""
    exit 1
}

$issPath = Join-Path $ProjectDir "TheBoysLauncher.iss"
$targetExe = Join-Path $ProjectDir "TheBoysLauncher.exe"

if (-not (Test-Path $issPath)) {
    Write-Error "Cannot find $issPath"
    exit 1
}
if (-not (Test-Path $targetExe)) {
    Write-Error "The target EXE was not found at $targetExe. Build the exe first (make build-windows or tools/build.ps1)."
    exit 1
}

# Create output directory if it doesn't exist
New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null

# Update version in the .iss file
Write-Host "Updating version in Inno Setup script to: $Version" -ForegroundColor Yellow
$issContent = Get-Content $issPath
$updatedContent = $issContent | ForEach-Object {
    if ($_ -match '^#define MyAppVersion') {
        "#define MyAppVersion `"$Version`""
    } else {
        $_
    }
}
$updatedContent | Set-Content $issPath -Encoding UTF8

# Build the installer
Write-Host "Running ISCC.exe to compile installer..." -ForegroundColor Yellow
$compilerArgs = @(
    "/Q",  # Quiet mode (less output)
    "/O`"$OutputDir`"",  # Output directory
    "/F`"TheBoysLauncher-Setup-$Version`"",  # Output filename
    "/DMyAppVersion=`"$Version`"",  # Define version
    "`"$issPath`""
)

$process = Start-Process -FilePath "ISCC.exe" -ArgumentList $compilerArgs -Wait -PassThru -NoNewWindow

if ($process.ExitCode -ne 0) { 
    Write-Error "ISCC.exe failed with exit code $($process.ExitCode)"
    exit $process.ExitCode 
}

$installerPath = Join-Path $OutputDir "TheBoysLauncher-Setup-$Version.exe"

if (Test-Path $installerPath) {
    Write-Host "Installer built successfully: $installerPath" -ForegroundColor Green
    Write-Host "Notes: Run the installer to verify the license, icon, and feature selection (shortcuts)." -ForegroundColor Cyan
} else {
    Write-Error "Installer was not created at expected location: $installerPath"
    exit 1
}
