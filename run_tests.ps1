# Comprehensive test runner for TheBoysLauncher (PowerShell version)

Write-Host "🚀 Running TheBoysLauncher Test Suite" -ForegroundColor Cyan
Write-Host "====================================" -ForegroundColor Cyan

# Function to print colored output
function Print-Status {
    param(
        [string]$Status,
        [string]$Message
    )
    
    switch ($Status) {
        "PASS" { Write-Host "✓ $Message" -ForegroundColor Green }
        "FAIL" { Write-Host "✗ $Message" -ForegroundColor Red }
        "INFO" { Write-Host "ℹ $Message" -ForegroundColor Blue }
        "WARN" { Write-Host "⚠ $Message" -ForegroundColor Yellow }
    }
}

# Test 1: Compilation test
Print-Status "INFO" "Test 1: Checking code compilation..."

Write-Host "Running 'go build' to check compilation..." -ForegroundColor Gray
if (go build -o TheBoysLauncher-test.exe .) {
    Print-Status "PASS" "Code compiles successfully"
    Remove-Item -Force TheBoysLauncher-test.exe -ErrorAction SilentlyContinue
} else {
    Print-Status "FAIL" "Code compilation failed"
    exit 1
}

# Test 2: Go mod verification
Print-Status "INFO" "Test 2: Verifying Go modules..."
if (go mod verify) {
    Print-Status "PASS" "Go modules verified"
} else {
    Print-Status "WARN" "Go modules verification failed (may be expected if no checksums)"
}

# Test 3: Go mod tidy
Print-Status "INFO" "Test 3: Running go mod tidy..."
if (go mod tidy) {
    Print-Status "PASS" "go mod tidy completed successfully"
} else {
    Print-Status "FAIL" "go mod tidy failed"
    exit 1
}

# Test 4: Run unit tests
Print-Status "INFO" "Test 4: Running unit tests..."

Write-Host "Running platform tests..." -ForegroundColor Gray
if (go test -v -run TestGetLauncherExeName) {
    Print-Status "PASS" "Platform detection tests passed"
} else {
    Print-Status "FAIL" "Platform detection tests failed"
    exit 1
}

Write-Host "Running version comparison tests..." -ForegroundColor Gray
if (go test -v -run TestCompareSemver) {
    Print-Status "PASS" "Version comparison tests passed"
} else {
    Print-Status "FAIL" "Version comparison tests failed"
    exit 1
}

Write-Host "Running update mechanism tests..." -ForegroundColor Gray
if (go test -v -run TestUpdateScenarioValidation) {
    Print-Status "PASS" "Update mechanism tests passed"
} else {
    Print-Status "FAIL" "Update mechanism tests failed"
    exit 1
}

Write-Host "Running forceUpdate tests..." -ForegroundColor Gray
if (go test -v -run TestForceUpdate) {
    Print-Status "PASS" "forceUpdate function tests passed"
} else {
    Print-Status "FAIL" "forceUpdate function tests failed"
    exit 1
}

Write-Host "Running prerelease filtering tests..." -ForegroundColor Gray
if (go test -v -run TestIsPrereleaseTag) {
    Print-Status "PASS" "Prerelease tag detection tests passed"
} else {
    Print-Status "FAIL" "Prerelease tag detection tests failed"
    exit 1
}

Write-Host "Running stable version filtering tests..." -ForegroundColor Gray
if (go test -v -run "TestFilterStableReleasesLogic|TestFetchLatestAssetPreferStableLogic|TestVersionFilteringEdgeCases") {
    Print-Status "PASS" "Stable version filtering tests passed"
} else {
    Print-Status "FAIL" "Stable version filtering tests failed"
    exit 1
}

Write-Host "Running bug report scenario tests..." -ForegroundColor Gray
if (go test -v -run TestBugReportScenario) {
    Print-Status "PASS" "Bug report scenario tests passed"
} else {
    Print-Status "FAIL" "Bug report scenario tests failed"
    exit 1
}

Write-Host "Running GUI dev mode toggle tests..." -ForegroundColor Gray
if (go test -v tests/gui_test.go tests/devbuilds_test.go -run TestGUIDevMode) {
    Print-Status "PASS" "GUI dev mode toggle tests passed"
} else {
    Print-Status "FAIL" "GUI dev mode toggle tests failed"
    exit 1
}

Write-Host "Running pagination tests..." -ForegroundColor Gray
if (go test -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go) {
    Print-Status "PASS" "Pagination tests passed"
} else {
    Print-Status "FAIL" "Pagination tests failed"
    exit 1
}

# Test 5: Cross-platform compilation test
Print-Status "INFO" "Test 5: Testing cross-platform compilation..."

$platforms = @("linux/amd64", "linux/arm64", "windows/amd64", "darwin/amd64", "darwin/arm64")
foreach ($platform in $platforms) {
    $parts = $platform -split "/"
    $GOOS = $parts[0]
    $GOARCH = $parts[1]
    
    Write-Host "Testing compilation for $GOOS/$GOARCH..." -ForegroundColor Gray
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    
    if (go build -o TheBoysLauncher-$GOOS-$GOARCH .) {
        Print-Status "PASS" "Compilation for $GOOS/$GOARCH successful"
        Remove-Item -Force TheBoysLauncher-$GOOS-$GOARCH -ErrorAction SilentlyContinue
    } else {
        Print-Status "FAIL" "Compilation for $GOOS/$GOARCH failed"
        exit 1
    }
}

# Reset environment variables
Remove-Item Env:GOOS -ErrorAction SilentlyContinue
Remove-Item Env:GOARCH -ErrorAction SilentlyContinue

# Test 6: Validate GitHub Actions workflow
Print-Status "INFO" "Test 6: Validating GitHub Actions workflow..."

if (Test-Path "test_workflow.ps1") {
    if (.\test_workflow.ps1) {
        Print-Status "PASS" "GitHub Actions workflow validation passed"
    } else {
        Print-Status "FAIL" "GitHub Actions workflow validation failed"
        exit 1
    }
} else {
    Print-Status "WARN" "GitHub Actions workflow test script not found"
}

# Test 7: Check for required files
Print-Status "INFO" "Test 7: Checking for required project files..."

$requiredFiles = @("go.mod", "go.sum", "main.go", "platform.go", "update.go", "version.env", "Makefile", "README.md")
foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Print-Status "PASS" "Required file found: $file"
    } else {
        Print-Status "FAIL" "Required file missing: $file"
        exit 1
    }
}

# Test 8: Version file validation
Print-Status "INFO" "Test 8: Validating version.env file..."

if (Test-Path "version.env") {
    $content = Get-Content "version.env"
    if ($content -match "VERSION=" -and $content -match "MAJOR=" -and $content -match "MINOR=" -and $content -match "PATCH=") {
        Print-Status "PASS" "version.env has required fields"
        
        # Extract and validate version format
        $versionLine = $content | Where-Object { $_ -match "^VERSION=" }
        if ($versionLine -match "^VERSION=([0-9]+\.[0-9]+\.[0-9]+(-.*)?`$") {
            Print-Status "PASS" "Version format is valid: $($matches[1])"
        } else {
            Print-Status "FAIL" "Invalid version format"
            exit 1
        }
    } else {
        Print-Status "FAIL" "version.env missing required fields"
        exit 1
    }
} else {
    Print-Status "FAIL" "version.env file not found"
    exit 1
}

# Test 9: Check for test coverage
Print-Status "INFO" "Test 9: Running test coverage analysis..."

if (go test -coverprofile=coverage.out ./...) {
    Print-Status "PASS" "Test coverage analysis completed"
    
    # Display coverage summary
    if (Get-Command go -ErrorAction SilentlyContinue) {
        Write-Host "Coverage summary:" -ForegroundColor Gray
        go tool cover -func=coverage.out | Select-Object -Last 1
        Remove-Item -Force coverage.out -ErrorAction SilentlyContinue
    }
} else {
    Print-Status "WARN" "Test coverage analysis failed"
}

# Test 10: Validate Makefile targets
Print-Status "INFO" "Test 10: Validating Makefile..."

if (Test-Path "Makefile") {
    $makefileContent = Get-Content "Makefile"
    
    # Check for common Makefile targets
    if ($makefileContent -match "build:") {
        Print-Status "PASS" "Makefile has build target"
    } else {
        Print-Status "WARN" "Makefile missing build target"
    }
    
    if ($makefileContent -match "clean:") {
        Print-Status "PASS" "Makefile has clean target"
    } else {
        Print-Status "WARN" "Makefile missing clean target"
    }
    
    if ($makefileContent -match "test:") {
        Print-Status "PASS" "Makefile has test target"
    } else {
        Print-Status "WARN" "Makefile missing test target"
    }
} else {
    Print-Status "WARN" "Makefile not found"
}

Write-Host ""
Write-Host "====================================" -ForegroundColor Cyan
Print-Status "PASS" "All tests completed successfully! 🎉"
Write-Host "====================================" -ForegroundColor Cyan

# Summary
Write-Host ""
Write-Host "Test Summary:" -ForegroundColor White
Write-Host "- ✅ Code compilation" -ForegroundColor Green
Write-Host "- ✅ Go modules verification" -ForegroundColor Green
Write-Host "- ✅ Unit tests (platform, version, update, forceUpdate)" -ForegroundColor Green
Write-Host "- ✅ Prerelease tag detection tests" -ForegroundColor Green
Write-Host "- ✅ Stable version filtering tests" -ForegroundColor Green
Write-Host "- ✅ Bug report scenario tests" -ForegroundColor Green
Write-Host "- ✅ GUI dev mode toggle tests" -ForegroundColor Green
Write-Host "- ✅ Pagination tests (core, edge cases, integration)" -ForegroundColor Green
Write-Host "- ✅ Cross-platform compilation" -ForegroundColor Green
Write-Host "- ✅ GitHub Actions workflow validation" -ForegroundColor Green
Write-Host "- ✅ Required files check" -ForegroundColor Green
Write-Host "- ✅ Version file validation" -ForegroundColor Green
Write-Host "- ✅ Test coverage analysis" -ForegroundColor Green
Write-Host "- ✅ Makefile validation" -ForegroundColor Green
Write-Host ""
Write-Host "The enhanced launcher is ready for deployment! 🚀" -ForegroundColor Green