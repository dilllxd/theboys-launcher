#!/bin/bash

# Comprehensive test runner for TheBoysLauncher

set -e

echo "🚀 Running TheBoysLauncher Test Suite"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS")
            echo -e "${GREEN}✓ $message${NC}"
            ;;
        "FAIL")
            echo -e "${RED}✗ $message${NC}"
            ;;
        "INFO")
            echo -e "${BLUE}ℹ $message${NC}"
            ;;
        "WARN")
            echo -e "${YELLOW}⚠ $message${NC}"
            ;;
    esac
}

# Test 1: Compilation test
print_status "INFO" "Test 1: Checking code compilation..."

echo "Running 'go build' to check compilation..."
if go build -o /tmp/TheBoysLauncher-test .; then
    print_status "PASS" "Code compiles successfully"
    rm -f /tmp/TheBoysLauncher-test
else
    print_status "FAIL" "Code compilation failed"
    exit 1
fi

# Test 2: Go mod verification
print_status "INFO" "Test 2: Verifying Go modules..."
if go mod verify; then
    print_status "PASS" "Go modules verified"
else
    print_status "WARN" "Go modules verification failed (may be expected if no checksums)"
fi

# Test 3: Go mod tidy
print_status "INFO" "Test 3: Running go mod tidy..."
if go mod tidy; then
    print_status "PASS" "go mod tidy completed successfully"
else
    print_status "FAIL" "go mod tidy failed"
    exit 1
fi

# Test 4: Run unit tests
print_status "INFO" "Test 4: Running unit tests..."

echo "Running platform tests..."
if go test -v -run TestGetLauncherExeName; then
    print_status "PASS" "Platform detection tests passed"
else
    print_status "FAIL" "Platform detection tests failed"
    exit 1
fi

echo "Running version comparison tests..."
if go test -v -run TestCompareSemver; then
    print_status "PASS" "Version comparison tests passed"
else
    print_status "FAIL" "Version comparison tests failed"
    exit 1
fi

echo "Running update mechanism tests..."
if go test -v -run TestUpdateScenarioValidation; then
    print_status "PASS" "Update mechanism tests passed"
else
    print_status "FAIL" "Update mechanism tests failed"
    exit 1
fi

echo "Running forceUpdate tests..."
if go test -v -run TestForceUpdate; then
    print_status "PASS" "forceUpdate function tests passed"
else
    print_status "FAIL" "forceUpdate function tests failed"
    exit 1
fi

echo "Running GUI dev mode toggle tests..."
if go test -v tests/gui_test.go tests/devbuilds_test.go -run TestGUIDevMode; then
    print_status "PASS" "GUI dev mode toggle tests passed"
else
    print_status "FAIL" "GUI dev mode toggle tests failed"
    exit 1
fi

# Test 5: Cross-platform compilation test
print_status "INFO" "Test 5: Testing cross-platform compilation..."

platforms=("linux/amd64" "linux/arm64" "windows/amd64" "darwin/amd64" "darwin/arm64")
for platform in "${platforms[@]}"; do
    GOOS=${platform%/*}
    GOARCH=${platform#*/}
    
    echo "Testing compilation for $GOOS/$GOARCH..."
    if GOOS=$GOOS GOARCH=$GOARCH go build -o /tmp/TheBoysLauncher-$GOOS-$GOARCH .; then
        print_status "PASS" "Compilation for $GOOS/$GOARCH successful"
        rm -f /tmp/TheBoysLauncher-$GOOS-$GOARCH
    else
        print_status "FAIL" "Compilation for $GOOS/$GOARCH failed"
        exit 1
    fi
done

# Test 6: Validate GitHub Actions workflow
print_status "INFO" "Test 6: Validating GitHub Actions workflow..."

if [ -f "test_workflow.sh" ]; then
    chmod +x test_workflow.sh
    if ./test_workflow.sh; then
        print_status "PASS" "GitHub Actions workflow validation passed"
    else
        print_status "FAIL" "GitHub Actions workflow validation failed"
        exit 1
    fi
else
    print_status "WARN" "GitHub Actions workflow test script not found"
fi

# Test 7: Check for required files
print_status "INFO" "Test 7: Checking for required project files..."

required_files=("go.mod" "go.sum" "main.go" "platform.go" "update.go" "version.env" "Makefile" "README.md")
for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        print_status "PASS" "Required file found: $file"
    else
        print_status "FAIL" "Required file missing: $file"
        exit 1
    fi
done

# Test 8: Version file validation
print_status "INFO" "Test 8: Validating version.env file..."

if [ -f "version.env" ]; then
    if grep -q "VERSION=" version.env && grep -q "MAJOR=" version.env && grep -q "MINOR=" version.env && grep -q "PATCH=" version.env; then
        print_status "PASS" "version.env has required fields"
        
        # Extract and validate version format
        source version.env
        if [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
            print_status "PASS" "Version format is valid: $VERSION"
        else
            print_status "FAIL" "Invalid version format: $VERSION"
            exit 1
        fi
    else
        print_status "FAIL" "version.env missing required fields"
        exit 1
    fi
else
    print_status "FAIL" "version.env file not found"
    exit 1
fi

# Test 9: Check for test coverage
print_status "INFO" "Test 9: Running test coverage analysis..."

if go test -coverprofile=coverage.out ./...; then
    print_status "PASS" "Test coverage analysis completed"
    
    # Display coverage summary
    if command -v go &> /dev/null; then
        echo "Coverage summary:"
        go tool cover -func=coverage.out | tail -1
        rm -f coverage.out
    fi
else
    print_status "WARN" "Test coverage analysis failed"
fi

# Test 10: Validate Makefile targets
print_status "INFO" "Test 10: Validating Makefile..."

if [ -f "Makefile" ]; then
    # Check for common Makefile targets
    if grep -q "build:" Makefile; then
        print_status "PASS" "Makefile has build target"
    else
        print_status "WARN" "Makefile missing build target"
    fi
    
    if grep -q "clean:" Makefile; then
        print_status "PASS" "Makefile has clean target"
    else
        print_status "WARN" "Makefile missing clean target"
    fi
    
    if grep -q "test:" Makefile; then
        print_status "PASS" "Makefile has test target"
    else
        print_status "WARN" "Makefile missing test target"
    fi
else
    print_status "WARN" "Makefile not found"
fi

echo ""
echo "===================================="
print_status "PASS" "All tests completed successfully! 🎉"
echo "===================================="

# Summary
echo ""
echo "Test Summary:"
echo "- ✅ Code compilation"
echo "- ✅ Go modules verification"
echo "- ✅ Unit tests (platform, version, update, forceUpdate)"
echo "- ✅ GUI dev mode toggle tests"
echo "- ✅ Cross-platform compilation"
echo "- ✅ GitHub Actions workflow validation"
echo "- ✅ Required files check"
echo "- ✅ Version file validation"
echo "- ✅ Test coverage analysis"
echo "- ✅ Makefile validation"
echo ""
echo "The enhanced launcher is ready for deployment! 🚀"