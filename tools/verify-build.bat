@echo off
REM Quick build verification script for TheBoysLauncher (Windows)
REM Run this before committing changes to ensure basic compilation

echo 🔍 Running quick build verification...

REM Check if required files exist
if not exist "go.mod" (
    echo ❌ Error: go.mod not found. Are you in the right directory?
    exit /b 1
)

if not exist "Makefile" (
    echo ❌ Error: Makefile not found. Are you in the right directory?
    exit /b 1
)

REM Run quick build check
echo 📦 Checking Go compilation...
go build -o %TEMP%\theboys-test-build.exe .
if errorlevel 1 (
    echo ❌ Build failed! Please fix compilation errors before committing.
    exit /b 1
)

REM Clean up test build
if exist "%TEMP%\theboys-test-build.exe" del "%TEMP%\theboys-test-build.exe"

echo 📦 Running basic Go checks...
go fmt ./...
if errorlevel 1 (
    echo ⚠️  Some files need formatting. Run 'go fmt ./...' to fix.
)

go vet ./...
if errorlevel 1 (
    echo ❌ Go vet found issues. Please fix before committing.
    exit /b 1
)

echo ✅ Build verification passed! Ready to commit.
exit /b 0