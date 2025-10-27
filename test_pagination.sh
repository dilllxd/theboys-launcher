#!/bin/bash

# Test script for paginated scraping functionality
# This script runs all pagination-related tests for TheBoys Launcher

echo "=========================================="
echo "Testing Paginated Scraping Functionality"
echo "=========================================="
echo ""

# Build the launcher first
echo "Building TheBoys Launcher..."
go build -o TheBoysLauncher .
if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi
echo "✅ Build successful!"
echo ""

# Run pagination tests
echo "Running pagination tests..."
echo "----------------------------------------"

# Test core pagination functionality
echo "1. Testing core pagination functionality..."
go test -v tests/pagination_test.go
if [ $? -ne 0 ]; then
    echo "❌ Core pagination tests failed!"
    exit 1
fi
echo "✅ Core pagination tests passed!"
echo ""

# Test pagination edge cases
echo "2. Testing pagination edge cases..."
go test -v tests/pagination_edge_cases_test.go
if [ $? -ne 0 ]; then
    echo "❌ Pagination edge cases tests failed!"
    exit 1
fi
echo "✅ Pagination edge cases tests passed!"
echo ""

# Test pagination integration scenarios
echo "3. Testing pagination integration scenarios..."
go test -v tests/pagination_integration_test.go
if [ $? -ne 0 ]; then
    echo "❌ Pagination integration tests failed!"
    exit 1
fi
echo "✅ Pagination integration tests passed!"
echo ""

# Run all pagination tests together
echo "4. Running all pagination tests together..."
go test -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
if [ $? -ne 0 ]; then
    echo "❌ Combined pagination tests failed!"
    exit 1
fi
echo "✅ Combined pagination tests passed!"
echo ""

# Test with race detection
echo "5. Running pagination tests with race detection..."
go test -race -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
if [ $? -ne 0 ]; then
    echo "❌ Pagination tests with race detection failed!"
    exit 1
fi
echo "✅ Pagination tests with race detection passed!"
echo ""

# Test with coverage
echo "6. Running pagination tests with coverage..."
go test -cover -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
if [ $? -ne 0 ]; then
    echo "❌ Pagination tests with coverage failed!"
    exit 1
fi
echo "✅ Pagination tests with coverage passed!"
echo ""

echo "=========================================="
echo "✅ All pagination tests completed successfully!"
echo "=========================================="
echo ""
echo "Test Summary:"
echo "- Core pagination functionality: ✅"
echo "- Pagination edge cases: ✅"
echo "- Pagination integration scenarios: ✅"
echo "- Combined pagination tests: ✅"
echo "- Race detection: ✅"
echo "- Code coverage: ✅"
echo ""
echo "The paginated scraping feature is working correctly!"