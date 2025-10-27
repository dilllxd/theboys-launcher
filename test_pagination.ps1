# Test script for paginated scraping functionality
# This script runs all pagination-related tests for TheBoys Launcher

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Testing Paginated Scraping Functionality" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""

# Build launcher first
Write-Host "Building TheBoys Launcher..." -ForegroundColor Yellow
go build -o TheBoysLauncher.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host ""

# Run pagination tests
Write-Host "Running pagination tests..." -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor Gray

# Test core pagination functionality
Write-Host "1. Testing core pagination functionality..." -ForegroundColor Yellow
go test -v tests/pagination_test.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Core pagination tests failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Core pagination tests passed!" -ForegroundColor Green
Write-Host ""

# Test pagination edge cases
Write-Host "2. Testing pagination edge cases..." -ForegroundColor Yellow
go test -v tests/pagination_edge_cases_test.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Pagination edge cases tests failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Pagination edge cases tests passed!" -ForegroundColor Green
Write-Host ""

# Test pagination integration scenarios
Write-Host "3. Testing pagination integration scenarios..." -ForegroundColor Yellow
go test -v tests/pagination_integration_test.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Pagination integration tests failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Pagination integration tests passed!" -ForegroundColor Green
Write-Host ""

# Run all pagination tests together
Write-Host "4. Running all pagination tests together..." -ForegroundColor Yellow
go test -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Combined pagination tests failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Combined pagination tests passed!" -ForegroundColor Green
Write-Host ""

# Test with race detection
Write-Host "5. Running pagination tests with race detection..." -ForegroundColor Yellow
go test -race -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Pagination tests with race detection failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Pagination tests with race detection passed!" -ForegroundColor Green
Write-Host ""

# Test with coverage
Write-Host "6. Running pagination tests with coverage..." -ForegroundColor Yellow
go test -cover -v tests/pagination_test.go tests/pagination_edge_cases_test.go tests/pagination_integration_test.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Pagination tests with coverage failed!" -ForegroundColor Red
    exit 1
}
Write-Host "✅ Pagination tests with coverage passed!" -ForegroundColor Green
Write-Host ""

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "✅ All pagination tests completed successfully!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Test Summary:" -ForegroundColor White
Write-Host "- Core pagination functionality: ✅" -ForegroundColor Green
Write-Host "- Pagination edge cases: ✅" -ForegroundColor Green
Write-Host "- Pagination integration scenarios: ✅" -ForegroundColor Green
Write-Host "- Combined pagination tests: ✅" -ForegroundColor Green
Write-Host "- Race detection: ✅" -ForegroundColor Green
Write-Host "- Code coverage: ✅" -ForegroundColor Green
Write-Host ""
Write-Host "The paginated scraping feature is working correctly!" -ForegroundColor Green