#!/bin/bash

# Simplified Settings & Backup-less Dev Mode Test Script
# This script tests the simplified settings menu and backup-less dev mode switching

set -e

echo "=== TheBoys Launcher Simplified Settings Test ==="
echo "Testing simplified settings menu and backup-less dev mode switching"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -e "${YELLOW}Running test: $test_name${NC}"
    
    if eval "$test_command"; then
        echo -e "${GREEN}✓ PASSED: $test_name${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAILED: $test_name${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Function to check if file contains text
file_contains() {
    local file="$1"
    local text="$2"
    grep -q "$text" "$file" 2>/dev/null
}

# Function to check if file exists and is not empty
file_exists_not_empty() {
    local file="$1"
    [[ -f "$file" && -s "$file" ]]
}

echo "=== Phase 1: Build Verification ==="

# Test 1: Build the launcher
run_test "Build Launcher" "go build -ldflags='-s -w -X main.version=v3.0.1-test' -o TheBoysLauncher ."

# Test 2: Verify binary exists
run_test "Verify Binary Exists" "file_exists_not_empty TheBoysLauncher"

echo
echo "=== Phase 2: Code Verification ==="

# Test 3: Check for Save & Apply button
run_test "Save & Apply Button Present" "file_contains gui.go 'Save & Apply'"

# Test 4: Check for pre-update validation
run_test "Pre-update Validation Present" "file_contains gui.go 'Validating update availability'"

# Test 5: Check for fallback mechanism
run_test "Fallback Mechanism Present" "file_contains gui.go 'Attempting fallback to stable'"

# Test 6: Verify no backup code in GUI
run_test "No Backup Code in GUI" "! file_contains gui.go 'backup'"

echo
echo "=== Phase 3: Unit Tests ==="

# Test 7: Run all unit tests
run_test "All Unit Tests Pass" "go test -v ./tests/..."

echo
echo "=== Phase 4: Integration Tests ==="

# Test 8: Test simplified settings workflow
run_test "Simplified Settings Workflow" "go test -v -run TestSimplifiedSettings ./tests/"

# Test 9: Test backup-less dev mode
run_test "Backup-less Dev Mode" "go test -v -run TestBackuplessDevMode ./tests/"

# Test 10: Test error handling
run_test "Error Handling" "go test -v -run TestErrorHandling ./tests/"

echo
echo "=== Phase 5: Edge Case Tests ==="

# Test 11: Test network error handling
run_test "Network Error Handling" "go test -v -run TestNetworkError ./tests/"

# Test 12: Test version unavailability
run_test "Version Unavailability" "go test -v -run TestVersionUnavailable ./tests/"

# Test 13: Test settings corruption
run_test "Settings Corruption" "go test -v -run TestSettingsCorruption ./tests/"

echo
echo "=== Phase 6: Performance Tests ==="

# Test 14: Test build performance
run_test "Build Performance" "time go build -ldflags='-s -w' -o TheBoysLauncher-perf ."

# Test 15: Test test performance
run_test "Test Performance" "time go test ./tests/..."

echo
echo "=== Test Results Summary ==="
echo "Total Tests: $TESTS_TOTAL"
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo "The simplified settings implementation is working correctly."
    exit 0
else
    echo -e "${RED}✗ Some tests failed!${NC}"
    echo "Please review the failed tests and fix the issues."
    exit 1
fi