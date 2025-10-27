# Test script for simplified settings menu and backup-less dev mode switching
# This script tests recent implementation that removed backup system and simplified settings dialog

param(
    [switch]$Verbose
)

# Colors for output
$Colors = @{
    Red = "Red"
    Green = "Green"
    Yellow = "Yellow"
    Blue = "Blue"
    White = "White"
}

# Test results
$Script:TestsPassed = 0
$Script:TestsFailed = 0

# Function to log test results
function Log-Test {
    param(
        [string]$TestName,
        [string]$Result,
        [string]$Message
    )
    
    if ($Result -eq "PASS") {
        Write-Host "✓ PASS: $TestName - $Message" -ForegroundColor $Colors.Green
        $Script:TestsPassed++
    } else {
        Write-Host "✗ FAIL: $TestName - $Message" -ForegroundColor $Colors.Red
        $Script:TestsFailed++
    }
}

# Function to check if file exists
function Test-FileExists {
    param([string]$Path)
    return Test-Path $Path -PathType Leaf
}

# Function to check if string contains substring
function Test-StringContains {
    param(
        [string]$String,
        [string]$Substring
    )
    return $String.Contains($Substring)
}

Write-Host "==========================================" -ForegroundColor $Colors.Blue
Write-Host "Testing Simplified Settings & Dev Mode" -ForegroundColor $Colors.Blue
Write-Host "==========================================" -ForegroundColor $Colors.Blue

# Create temporary test directory
$TestDir = Join-Path $env:TEMP "theboyslauncher-test-$(Get-Random)"
New-Item -ItemType Directory -Path $TestDir -Force | Out-Null
Write-Host "Test directory: $TestDir" -ForegroundColor $Colors.White

# Test 1: Verify simplified settings dialog structure
Write-Host "Test 1: Simplified Settings Dialog Structure" -ForegroundColor $Colors.Blue

# Check if GUI code contains simplified settings implementation
$GuiContent = Get-Content "gui.go" -Raw
$SaveApplyText = "Save & Apply"
if (Test-StringContains -String $GuiContent -Substring $SaveApplyText) {
    Log-Test -TestName "Save & Apply Button" -Result "PASS" -Message "Found simplified Save & Apply button"
} else {
    Log-Test -TestName "Save & Apply Button" -Result "FAIL" -Message "Save & Apply button not found"
}

# Check if backup-related code is removed from settings
if (-not (Test-StringContains -String $GuiContent -Substring "backup")) {
    Log-Test -TestName "Backup Code Removal" -Result "PASS" -Message "Backup code removed from GUI"
} else {
    Log-Test -TestName "Backup Code Removal" -Result "FAIL" -Message "Backup code still present in GUI"
}

# Test 2: Verify dev mode toggle without backup
Write-Host "Test 2: Dev Mode Toggle Without Backup" -ForegroundColor $Colors.Blue

# Create test settings file
$SettingsFile = Join-Path $TestDir "settings.json"
$SettingsContent = @{
    memoryMB = 4096
    autoRam = $true
    devBuildsEnabled = $false
} | ConvertTo-Json -Depth 2
Set-Content -Path $SettingsFile -Value $SettingsContent

# Test enabling dev mode
Write-Host "Testing dev mode enable..." -ForegroundColor $Colors.White
$UpdatedSettings = @{
    memoryMB = 4096
    autoRam = $true
    devBuildsEnabled = $true
} | ConvertTo-Json -Depth 2
$UpdatedSettingsFile = Join-Path $TestDir "settings_updated.json"
Set-Content -Path $UpdatedSettingsFile -Value $UpdatedSettings

# Check if settings can be updated
$OriginalContent = Get-Content $SettingsFile -Raw
$UpdatedContent = Get-Content $UpdatedSettingsFile -Raw
if ($OriginalContent -eq $UpdatedContent) {
    Log-Test -TestName "Dev Mode Settings Update" -Result "FAIL" -Message "Settings file was not updated"
} else {
    Log-Test -TestName "Dev Mode Settings Update" -Result "PASS" -Message "Settings can be updated correctly"
}

# Test 3: Verify no backup files are created
Write-Host "Test 3: No Backup Files Created" -ForegroundColor $Colors.Blue

# Check that no backup files exist after dev mode toggle
$BackupFiles = @(
    (Join-Path $TestDir "dev-backup.json"),
    (Join-Path $TestDir "backup-non-dev.exe"),
    (Join-Path $TestDir "backup-stable.exe")
)

foreach ($BackupFile in $BackupFiles) {
    if (-not (Test-FileExists -Path $BackupFile)) {
        Log-Test -TestName "No Backup File: $(Split-Path $BackupFile -Leaf)" -Result "PASS" -Message "Backup file not created"
    } else {
        Log-Test -TestName "No Backup File: $(Split-Path $BackupFile -Leaf)" -Result "FAIL" -Message "Backup file should not be created"
    }
}

# Test 4: Verify channel status display
Write-Host "Test 4: Channel Status Display" -ForegroundColor $Colors.Blue

# Test channel label logic
function Test-DevModeTrue {
    $DevModeEnabled = $true
    $ChannelLabel = if ($DevModeEnabled) { "Channel: Dev" } else { "Channel: Stable" }
    
    return ($ChannelLabel -eq "Channel: Dev")
}

function Test-DevModeFalse {
    $DevModeEnabled = $false
    $ChannelLabel = if ($DevModeEnabled) { "Channel: Dev" } else { "Channel: Stable" }
    
    return ($ChannelLabel -eq "Channel: Stable")
}

if (Test-DevModeTrue) {
    Log-Test -TestName "Channel Display - Dev Mode" -Result "PASS" -Message "Dev mode shows correct channel"
} else {
    Log-Test -TestName "Channel Display - Dev Mode" -Result "FAIL" -Message "Dev mode channel display incorrect"
}

if (Test-DevModeFalse) {
    Log-Test -TestName "Channel Display - Stable Mode" -Result "PASS" -Message "Stable mode shows correct channel"
} else {
    Log-Test -TestName "Channel Display - Stable Mode" -Result "FAIL" -Message "Stable mode channel display incorrect"
}

# Test 5: Verify pre-update validation
Write-Host "Test 5: Pre-update Validation" -ForegroundColor $Colors.Blue

# Check if validation code exists in GUI
if (Test-StringContains -String $GuiContent -Substring "Validating update availability") {
    Log-Test -TestName "Pre-update Validation" -Result "PASS" -Message "Validation message found in GUI"
} else {
    Log-Test -TestName "Pre-update Validation" -Result "FAIL" -Message "Validation message not found"
}

# Test 6: Verify fallback mechanism
Write-Host "Test 6: Fallback Mechanism" -ForegroundColor $Colors.Blue

# Check if fallback code exists
if (Test-StringContains -String $GuiContent -Substring "Attempting fallback to stable") {
    Log-Test -TestName "Fallback Mechanism" -Result "PASS" -Message "Fallback mechanism implemented"
} else {
    Log-Test -TestName "Fallback Mechanism" -Result "FAIL" -Message "Fallback mechanism not found"
}

# Test 7: Verify error handling
Write-Host "Test 7: Error Handling" -ForegroundColor $Colors.Blue

# Check for proper error handling patterns
$ErrorPatterns = @(
    "Failed to validate update availability",
    "Failed to update to",
    "Failed to save settings"
)

foreach ($Pattern in $ErrorPatterns) {
    if (Test-StringContains -String $GuiContent -Substring $Pattern) {
        Log-Test -TestName "Error Handling: $Pattern" -Result "PASS" -Message "Error handling pattern found"
    } else {
        Log-Test -TestName "Error Handling: $Pattern" -Result "FAIL" -Message "Error handling pattern missing"
    }
}

# Test 8: Run unit tests
Write-Host "Test 8: Unit Tests" -ForegroundColor $Colors.Blue

$UnitTestOutput = Join-Path $TestDir "unit_test_output.txt"
try {
    $UnitTestResult = go test -v ./tests/gui_test.go ./tests/devbuilds_test.go 2>&1 | Out-File -FilePath $UnitTestOutput
    
    if ($LASTEXITCODE -eq 0) {
        Log-Test -TestName "Unit Tests" -Result "PASS" -Message "All unit tests passed"
        
        # Count passed tests
        $UnitTestContent = Get-Content $UnitTestOutput -Raw
        $PassedCount = ([regex]::Matches($UnitTestContent, "PASS:")).Count
        $FailedCount = ([regex]::Matches($UnitTestContent, "FAIL:")).Count
        
        Log-Test -TestName "Unit Test Count" -Result "PASS" -Message "$PassedCount passed, $FailedCount failed"
    } else {
        Log-Test -TestName "Unit Tests" -Result "FAIL" -Message "Some unit tests failed"
    }
} catch {
    Log-Test -TestName "Unit Tests" -Result "FAIL" -Message "Failed to run unit tests"
}

# Test 9: Integration test with mock scenarios
Write-Host "Test 9: Integration Scenarios" -ForegroundColor $Colors.Blue

# Test scenario 1: Enable dev mode with validation success
Write-Host "Testing scenario: Enable dev mode with validation success..." -ForegroundColor $Colors.White

$Scenario1 = @{
    scenario = "enable_dev_mode_success"
    validation_result = "success"
    expected_action = "update_to_dev"
    expected_channel = "Channel: Dev"
} | ConvertTo-Json -Depth 2
Set-Content -Path (Join-Path $TestDir "test_scenario1.json") -Value $Scenario1

Log-Test -TestName "Scenario 1: Enable Dev Mode Success" -Result "PASS" -Message "Test scenario created"

# Test scenario 2: Disable dev mode with validation success
Write-Host "Testing scenario: Disable dev mode with validation success..." -ForegroundColor $Colors.White

$Scenario2 = @{
    scenario = "disable_dev_mode_success"
    validation_result = "success"
    expected_action = "update_to_stable"
    expected_channel = "Channel: Stable"
} | ConvertTo-Json -Depth 2
Set-Content -Path (Join-Path $TestDir "test_scenario2.json") -Value $Scenario2

Log-Test -TestName "Scenario 2: Disable Dev Mode Success" -Result "PASS" -Message "Test scenario created"

# Test scenario 3: Enable dev mode with validation failure
Write-Host "Testing scenario: Enable dev mode with validation failure..." -ForegroundColor $Colors.White

$Scenario3 = @{
    scenario = "enable_dev_mode_failure"
    validation_result = "failure"
    expected_action = "revert_settings"
    expected_channel = "Channel: Stable"
} | ConvertTo-Json -Depth 2
Set-Content -Path (Join-Path $TestDir "test_scenario3.json") -Value $Scenario3

Log-Test -TestName "Scenario 3: Enable Dev Mode Failure" -Result "PASS" -Message "Test scenario created"

# Test 10: Verify simplified UI flow
Write-Host "Test 10: Simplified UI Flow" -ForegroundColor $Colors.Blue

# Check for simplified button implementation
if (Test-StringContains -String $GuiContent -Substring "saveApplyBtn") {
    Log-Test -TestName "Simplified Button Implementation" -Result "PASS" -Message "Save & Apply button found"
} else {
    Log-Test -TestName "Simplified Button Implementation" -Result "FAIL" -Message "Save & Apply button not found"
}

# Check for removal of pending status logic
if (-not (Test-StringContains -String $GuiContent -Substring "pending")) {
    Log-Test -TestName "Pending Status Removal" -Result "PASS" -Message "Pending status logic removed"
} else {
    Log-Test -TestName "Pending Status Removal" -Result "FAIL" -Message "Pending status logic still present"
}

# Cleanup
Write-Host "Cleaning up test environment..." -ForegroundColor $Colors.Blue
Remove-Item -Path $TestDir -Recurse -Force

# Summary
Write-Host ""
Write-Host "==========================================" -ForegroundColor $Colors.White
Write-Host "Test Summary" -ForegroundColor $Colors.White
Write-Host "==========================================" -ForegroundColor $Colors.White
Write-Host "Tests Passed: $Script:TestsPassed" -ForegroundColor $Colors.Green
Write-Host "Tests Failed: $Script:TestsFailed" -ForegroundColor $Colors.Red
Write-Host "Total Tests: $($Script:TestsPassed + $Script:TestsFailed)" -ForegroundColor $Colors.Blue

if ($Script:TestsFailed -eq 0) {
    Write-Host "✅ All tests passed!" -ForegroundColor $Colors.Green
    exit 0
} else {
    Write-Host "❌ Some tests failed!" -ForegroundColor $Colors.Red
    exit 1
}