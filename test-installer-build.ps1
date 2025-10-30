# Test script to verify the installer build fix
param(
    [string]$Version = "v3.2.68-test"
)

Write-Host "Testing installer build with version $Version"
Write-Host "========================================"

# Run the build script
& .\scripts\build-installer.ps1 -Version $Version
$buildExitCode = $LASTEXITCODE

if ($buildExitCode -ne 0) {
    Write-Error "Build script failed with exit code $buildExitCode"
    exit $buildExitCode
}

# Check if the installer exists in the root directory
$expectedInstaller = "TheBoysLauncher-Setup-$Version.exe"
if (Test-Path $expectedInstaller) {
    $fileInfo = Get-Item $expectedInstaller
    Write-Host "SUCCESS: Installer found at root directory: $expectedInstaller"
    Write-Host "  Size: $($fileInfo.Length) bytes"
    Write-Host "  Created: $($fileInfo.CreationTime)"
} else {
    Write-Error "FAILURE: Installer not found at root directory: $expectedInstaller"
    
    # Check if it's in the installer subdirectory
    $installerDir = "installer"
    $subdirPath = Join-Path $installerDir $expectedInstaller
    if (Test-Path $subdirPath) {
        Write-Host "Installer found in installer subdirectory: $subdirPath"
        Write-Host "This indicates the move operation in build-installer.ps1 failed."
    } else {
        Write-Host "Installer not found in installer subdirectory either."
        Write-Host "Contents of installer directory:"
        if (Test-Path $installerDir) {
            Get-ChildItem -Path $installerDir -Force | ForEach-Object { 
                Write-Host "  $($_.Name)"
            }
        } else {
            Write-Host "  Installer directory does not exist"
        }
    }
    exit 1
}

Write-Host "========================================"
Write-Host "Test completed successfully!"