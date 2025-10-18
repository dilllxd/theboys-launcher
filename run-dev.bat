@echo off
echo Starting TheBoys Launcher in development mode...
echo Note: Press Ctrl+C to stop the development server
echo.

REM Check if Wails is installed
where wails >nul 2>nul
if %errorlevel% neq 0 (
    echo Wails CLI not found. Installing Wails CLI...
    go install github.com/wailsapp/wails/v2/cmd/wails@latest
    if %errorlevel% neq 0 (
        echo Failed to install Wails CLI. Please make sure Go is installed and in your PATH.
        pause
        exit /b 1
    )
)

REM Install dependencies
echo Installing dependencies...
go mod download
go mod tidy
cd frontend
call npm install
cd ..

REM Start development server
echo.
echo Starting Wails development server...
wails dev