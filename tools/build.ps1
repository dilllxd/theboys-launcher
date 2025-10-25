param(
    [string]$Version = "v2.0.3"
)

Write-Host "Building TheBoysLauncher with icon and version info..." -ForegroundColor Green

# Check if rsrc tool is available
try {
    $null = Get-Command rsrc -ErrorAction Stop
} catch {
    Write-Host "Installing rsrc tool..." -ForegroundColor Yellow
    go install github.com/akavel/rsrc@latest
}

# Check if icon file exists
if (-not (Test-Path "icon.ico")) {
    Write-Host "ERROR: icon.ico not found!" -ForegroundColor Red
    Write-Host "Please create/place an icon.ico file in this directory."
    Write-Host "See ICON_README.md for details."
    Read-Host "Press Enter to exit"
    exit 1
}

# Update version in resource.rc dynamically
Write-Host "Updating version info..." -ForegroundColor Blue
$versionClean = $Version -replace '^v', ''

# Convert version like "2.0.3" to "2,0,3,0"
$versionParts = $versionClean.Split('.')
if ($versionParts.Length -ge 3) {
    $versionNumbers = "$($versionParts[0]),$($versionParts[1]),$($versionParts[2]),0"
} else {
    $versionNumbers = "2,0,3,0"  # fallback
}

# Read, update, and write back the resource file
$resourceContent = Get-Content "resource.rc"
$resourceContent = $resourceContent -replace 'FILEVERSION\s+[\d,]+', "FILEVERSION         $versionNumbers"
$resourceContent = $resourceContent -replace 'PRODUCTVERSION\s+[\d,]+', "PRODUCTVERSION      $versionNumbers"
$resourceContent = $resourceContent -replace '"FileVersion"\s*,\s*"[^"]*"', ('"FileVersion",      "' + $versionClean + '"')
$resourceContent = $resourceContent -replace '"ProductVersion"\s*,\s*"[^"]*"', ('"ProductVersion",   "' + $versionClean + '"')
$resourceContent | Set-Content "resource.rc" -Encoding UTF8

# Compile resources
Write-Host "Compiling resources..." -ForegroundColor Blue
& rsrc -ico icon.ico -manifest resource.rc -o resource.syso

if ($LASTEXITCODE -ne 0) {
    Write-Host "Resource compilation failed!" -ForegroundColor Red
    exit 1
}

# Build the application with embedded resources
Write-Host "Building executable..." -ForegroundColor Blue
$env:GOOS = "windows"
$env:GOARCH = "amd64"
$env:CGO_ENABLED = "1"

go build -ldflags="-s -w -H=windowsgui -X main.version=$Version" -o TheBoysLauncher.exe .

if ($LASTEXITCODE -eq 0) {
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Output: TheBoysLauncher.exe" -ForegroundColor Cyan
    Write-Host "Version: $Version" -ForegroundColor Cyan
    Write-Host "Icon: embedded" -ForegroundColor Cyan
    Write-Host "Version info: embedded" -ForegroundColor Cyan
} else {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}