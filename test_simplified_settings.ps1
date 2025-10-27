# Simplified Settings & Backup-less Dev Mode Test Script (PowerShell)
# This script tests the simplified settings menu and backup-less dev mode switching

param(
    [switch]$Verbose,
    [switch]$SkipBuild
)

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    White = "White"
}

Write-Host "=== TheBoys Launcher Simplified Settings Test ===" -ForegroundColor $Colors.White
Write-Host "Testing simplified settings menu and backup-less dev mode switching" -ForegroundColor $Colors.White
Write-Host ""

# Test counters
$Script:TestsTotal = 0
$Script:TestsPassed = 0
$Script:TestsFailed = 0

# Function to run a test
function Run-Test {
    param(
        [string]$TestName,
        [scriptblock]$TestCommand
    )
    
    $Script:TestsTotal++
    Write-Host "Running test: $TestName" -ForegroundColor $Colors.Yellow
    
    try {
        $result = & $TestCommand
        if ($LASTEXITCODE -eq 0 -and $result) {
            Write-Host "✓ PASSED: $TestName" -ForegroundColor $Colors.Green
            $Script:TestsPassed++
            return $true
        } else {
            Write-Host "✗ FAILED: $TestName" -ForegroundColor $Colors.Red
            $Script:TestsFailed++
            return $false
        }
    } catch {
        Write-Host "✗ FAILED: $TestName - $($_.Exception.Message)" -ForegroundColor $Colors.Red
        $Script:TestsFailed++
        return $false
    }
}

# Function to check if file contains text
function File-Contains {
    param(
        [string]$File,
        [string]$Text
    )
    
    if (Test-Path $File) {
        $content = Get-Content $File -Raw
        return $content -match [regex]::Escape($Text)
    }
    return $false
}

# Function to check if file exists and is not empty
function File-Exists-Not-Empty {
    param([string]$File)
    
    return (Test-Path $File) -and ((Get-Item $File).Length -gt 0)
}

Write-Host "=== Phase 1: Build Verification ===" -ForegroundColor $Colors.White

# Test 1: Build the launcher
if (-not $SkipBuild) {
    Run-Test "Build Launcher" {
        go build -ldflags="-s -w -X main.version=v3.0.1-test" -o TheBoysLauncher.exe .
        $LASTEXITCODE -eq 0
    }
} else {
    Write-Host "Skipping build test (SkipBuild specified)" -ForegroundColor $Colors.Yellow
}

# Test 2: Verify binary exists
Run-Test "Verify Binary Exists" {
    File-Exists-Not-Empty "TheBoysLauncher.exe"
}

Write-Host ""
Write-Host "=== Phase 2: Code Verification ===" -ForegroundColor $Colors.White

# Test 3: Check for Save & Apply button
Run-Test "Save & Apply Button Present" {
    File-Contains "gui.go" "Save & Apply"
}

# Test 4: Check for pre-update validation
Run-Test "Pre-update Validation Present" {
    File-Contains "gui.go" "Validating update availability"
}

# Test 5: Check for fallback mechanism
Run-Test "Fallback Mechanism Present" {
    File-Contains "gui.go" "Attempting fallback to stable"
}

# Test 6: Verify no backup code in GUI
Run-Test "No Backup Code in GUI" {
    -not (File-Contains "gui.go" "backup")
}

Write-Host ""
Write-Host "=== Phase 3: Unit Tests ===" -ForegroundColor $Colors.White

# Test 7: Run all unit tests
Run-Test "All Unit Tests Pass" {
    if ($Verbose) {
        go test -v ./tests/...
    } else {
        go test ./tests/...
    }
    $LASTEXITCODE -eq 0
}

Write-Host ""
Write-Host "=== Phase 4: Integration Tests ===" -ForegroundColor $Colors.White

# Test 8: Test simplified settings workflow
Run-Test "Simplified Settings Workflow" {
    if ($Verbose) {
        go test -v -run TestSimplifiedSettings ./tests/
    } else {
        go test -run TestSimplifiedSettings ./tests/
    }
    $LASTEXITCODE -eq 0
}

# Test 9: Test backup-less dev mode
Run-Test "Backup-less Dev Mode" {
    if ($Verbose) {
        go test -v -run TestBackuplessDevMode ./tests/
    } else {
        go test -run TestBackuplessDevMode ./tests/
    }
    $LASTEXITCODE -eq 0
}

# Test 10: Test error handling
Run-Test "Error Handling" {
    if ($Verbose) {
        go test -v -run TestErrorHandling ./tests/
    } else {
        go test -run TestErrorHandling ./tests/
    }
    $LASTEXITCODE -eq 0
}

Write-Host ""
Write-Host "=== Phase 5: Edge Case Tests ===" -ForegroundColor $Colors.White

# Test 11: Test network error handling
Run-Test "Network Error Handling" {
    if ($Verbose) {
        go test -v -run TestNetworkError ./tests/
    } else {
        go test -run TestNetworkError ./tests/
    }
    $LASTEXITCODE -eq 0
}

# Test 12: Test version unavailability
Run-Test "Version Unavailability" {
    if ($Verbose) {
        go test -v -run TestVersionUnavailable ./tests/
    } else {
        go test -run TestVersionUnavailable ./tests/
    }
    $LASTEXITCODE -eq 0
}

# Test 13: Test settings corruption
Run-Test "Settings Corruption" {
    if ($Verbose) {
        go test -v -run TestSettingsCorruption ./tests/
    } else {
        go test -run TestSettingsCorruption ./tests/
    }
    $LASTEXITCODE -eq 0
}

Write-Host ""
Write-Host "=== Phase 6: Performance Tests ===" -ForegroundColor $Colors.White

# Test 14: Test build performance
Run-Test "Build Performance" {
    $startTime = Get-Date
    go build -ldflags="-s -w" -o TheBoysLauncher-perf.exe .
    $endTime = Get-Date
    $duration = $endTime - $startTime
    Write-Host "Build time: $($duration.TotalSeconds) seconds" -ForegroundColor $Colors.White
    $LASTEXITCODE -eq 0
}

# Test 15: Test test performance
Run-Test "Test Performance" {
    $startTime = Get-Date
    go test ./tests/...
    $endTime = Get-Date
    $duration = $endTime - $startTime
    Write-Host "Test time: $($duration.TotalSeconds) seconds" -ForegroundColor $Colors.White
    $LASTEXITCODE -eq 0
}

Write-Host ""
Write-Host "=== Test Results Summary ===" -ForegroundColor $Colors.White
Write-Host "Total Tests: $Script:TestsTotal" -ForegroundColor $Colors.White
Write-Host "Passed: $Script:TestsPassed" -ForegroundColor $Colors.Green
Write-Host "Failed: $Script:TestsFailed" -ForegroundColor $Colors.Red

if ($Script:TestsFailed -eq 0) {
    Write-Host "✓ All tests passed!" -ForegroundColor $Colors.Green
    Write-Host "The simplified settings implementation is working correctly." -ForegroundColor $Colors.White
    exit 0
} else {
    Write-Host "✗ Some tests failed!" -ForegroundColor $Colors.Red
    Write-Host "Please review the failed tests and fix the issues." -ForegroundColor $Colors.White
    exit 1
}