@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   TheBoys Launcher Development Server
echo ========================================
echo.
echo Starting TheBoys Launcher in development mode...
echo Note: Press Ctrl+C to stop the development server
echo.

REM Check if Go is installed
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed or not in your PATH.
    echo Please install Go from https://golang.org/dl/
    echo.
    pause
    exit /b 1
)

echo [✓] Go found

REM Check if Wails is installed in various possible locations
set WAILS_FOUND=0

REM Check if wails is in PATH
where wails >nul 2>nul
if %errorlevel% equ 0 (
    echo [✓] Wails CLI found in PATH
    set WAILS_FOUND=1
    set WAILS_CMD=wails
)

REM Check if wails is in Go's default bin directory
if %WAILS_FOUND% equ 0 (
    if exist "%USERPROFILE%\go\bin\wails.exe" (
        echo [✓] Wails CLI found in %%USERPROFILE%%\go\bin
        set WAILS_FOUND=1
        set WAILS_CMD="%USERPROFILE%\go\bin\wails.exe"
    )
)

REM Check if wails is in GOPATH/bin
if %WAILS_FOUND% equ 0 (
    for /f "tokens=*" %%i in ('go env GOPATH') do set GOPATH_DIR=%%i
    if exist "%GOPATH_DIR%\bin\wails.exe" (
        echo [✓] Wails CLI found in GOPATH/bin
        set WAILS_FOUND=1
        set WAILS_CMD="%GOPATH_DIR%\bin\wails.exe"
    )
)

REM If still not found, try to install it
if %WAILS_FOUND% equ 0 (
    echo [!] Wails CLI not found. Installing Wails CLI...
    echo This may take a moment...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %errorlevel% neq 0 (
        echo [ERROR] Failed to install Wails CLI.
        echo Please check your Go installation and try again.
        pause
        exit /b 1
    )

    REM Check if installation succeeded
    if exist "%USERPROFILE%\go\bin\wails.exe" (
        echo [✓] Wails CLI installed successfully
        set WAILS_CMD="%USERPROFILE%\go\bin\wails.exe"
    ) else (
        echo [ERROR] Wails CLI installation failed
        pause
        exit /b 1
    )
)

REM Install Go dependencies
echo.
echo [→] Installing Go dependencies...
go mod download
if %errorlevel% neq 0 (
    echo [ERROR] Failed to download Go dependencies
    pause
    exit /b 1
)
go mod tidy
if %errorlevel% neq 0 (
    echo [ERROR] Failed to tidy Go dependencies
    pause
    exit /b 1
)
echo [✓] Go dependencies installed

REM Install Node.js dependencies
echo.
echo [→] Installing Node.js dependencies...
cd frontend
call npm install
if %errorlevel% neq 0 (
    echo [ERROR] Failed to install Node.js dependencies
    cd ..
    pause
    exit /b 1
)
cd ..
echo [✓] Node.js dependencies installed

REM Start development server
echo.
echo ========================================
echo   Starting Wails Development Server
echo ========================================
echo.
echo Starting frontend development server...
echo This will preserve our mock bindings and provide better development experience
echo.

REM Start frontend dev server in background
echo [→] Starting frontend development server...
cd frontend
start "Frontend Dev Server" cmd /c "npm run dev"
cd ..

REM Wait a moment for frontend to start
timeout /t 3 /nobreak >nul

echo.
echo [→] Starting Wails backend with external frontend...
echo The application will open in a new window
echo.
echo Press Ctrl+C in this window to stop both servers
echo.

%WAILS_CMD% dev -frontenddevserverurl http://localhost:5173