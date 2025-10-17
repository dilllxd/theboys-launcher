@echo off
REM TheBoys Launcher - Cross-Platform Build Script (Windows)
REM This script builds single-file executables for Windows, macOS, and Linux
REM Each executable is self-contained and drops all files in the same directory

setlocal enabledelayedexpansion

echo TheBoys Launcher - Cross-Platform Build Script
echo ==================================================
echo Version: %1
echo.

REM Set version (default to "dev" if not provided)
set VERSION=%1
if "%VERSION%"=="" set VERSION=dev

REM Clean previous builds
echo Cleaning previous builds...
if exist build rmdir /s /q build
if exist dist rmdir /s /q dist
for %%f in (theboys-launcher-*.exe) do del "%%f" 2>nul

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: Go is not installed or not in PATH
    echo Please install Go from https://golang.org/dl/
    exit /b 1
)

REM Check if Node.js is installed
where node >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Error: Node.js is not installed or not in PATH
    echo Please install Node.js from https://nodejs.org/
    exit /b 1
)

REM Install Go dependencies
echo Installing Go dependencies...
go mod download
go mod tidy

REM Install frontend dependencies
echo Installing frontend dependencies...
cd frontend
call npm install --silent
cd ..

REM Check if Wails CLI is available, install if needed
where wails >nul 2>nul
if %ERRORLEVEL% neq 0 (
    echo Wails CLI not found, installing...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
)

REM Build frontend assets
echo Building frontend assets...
cd frontend
call npm run build
cd ..

echo.
echo Starting cross-platform builds...
echo.

REM Build flags for version injection
set LDFLAGS=-s -w -X main.version=%VERSION%

REM Function to build for a specific platform
:build_platform
set os=%~1
set arch=%~2
set output_name=%~3

echo Building for %os%/%arch%...

REM Set environment variables for cross-compilation
set GOOS=%os%
set GOARCH=%arch%
set CGO_ENABLED=0

REM Create build directory
set platform_dir=build\%os%-%arch%
if not exist "%platform_dir%" mkdir "%platform_dir%"

REM Build the application
if "%os%"=="windows" (
    REM Windows build with .exe extension
    set exe_name=%output_name%.exe
    go build -ldflags="%LDFLAGS%" -o "%platform_dir%\%exe_name%" .\cmd\launcher

    REM Verify the executable
    if exist "%platform_dir%\%exe_name%" (
        echo ✓ Windows build successful: %exe_name%
    ) else (
        echo ✗ Windows build failed
        exit /b 1
    )
) else (
    REM Unix-like build (Linux, macOS)
    go build -ldflags="%LDFLAGS%" -o "%platform_dir%\%output_name%" .\cmd\launcher

    REM Verify the executable
    if exist "%platform_dir%\%output_name%" (
        echo ✓ %os% build successful: %output_name%
    ) else (
        echo ✗ %os% build failed
        exit /b 1
    )
)

REM Unset environment variables
set GOOS=
set GOARCH=
set CGO_ENABLED=

goto :eof

REM Build for Windows (amd64)
call :build_platform windows amd64 theboys-launcher-windows-amd64

REM Build for Windows (arm64)
call :build_platform windows arm64 theboys-launcher-windows-arm64

REM Build for Linux (amd64)
call :build_platform linux amd64 theboys-launcher-linux-amd64

REM Build for Linux (arm64)
call :build_platform linux arm64 theboys-launcher-linux-arm64

REM Build for macOS (amd64)
call :build_platform darwin amd64 theboys-launcher-macos-amd64

REM Build for macOS (arm64)
call :build_platform darwin arm64 theboys-launcher-macos-arm64

echo.
echo All builds completed successfully!

REM Create distribution directory with proper naming
echo Creating distribution packages...
if not exist dist mkdir dist

REM Copy builds to dist directory with proper names
cd build

REM Windows builds
if exist windows-amd64\theboys-launcher-windows-amd64.exe (
    copy "windows-amd64\theboys-launcher-windows-amd64.exe" "..\dist\TheBoysLauncher.exe" >nul
    echo ✓ Created: TheBoysLauncher.exe (Windows x64)
)

if exist windows-arm64\theboys-launcher-windows-arm64.exe (
    copy "windows-arm64\theboys-launcher-windows-arm64.exe" "..\dist\TheBoysLauncher-arm64.exe" >nul
    echo ✓ Created: TheBoysLauncher-arm64.exe (Windows ARM64)
)

cd ..

echo.
echo Build Summary
echo ==============
echo All executables are self-contained and portable:
echo.
echo Windows:
echo   - TheBoysLauncher.exe (x64)
echo   - TheBoysLauncher-arm64.exe (ARM64)
echo.
echo Linux:
echo   - theboys-launcher-linux-amd64 (x64)
echo   - theboys-launcher-linux-arm64 (ARM64)
echo.
echo macOS:
echo   - theboys-launcher-macos-amd64 (Intel)
echo   - theboys-launcher-macos-arm64 (Apple Silicon)
echo.
echo Single-File Deployment: ✔
echo Each executable drops all files in the same directory as the executable
echo just like the legacy launcher. No installation required!
echo.
echo Portable Operation: ✔
echo - Windows: Creates files beside the .exe
echo - macOS: Creates ~/.theboys-launcher/ in user home
echo - Linux: Creates ~/.theboys-launcher/ in user home
echo.
echo Build artifacts are available in: dist

pause