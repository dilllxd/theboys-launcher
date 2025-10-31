@echo off
echo Upload Notifications Test
echo ==================
echo This test program verifies the fixed upload notification implementation.
echo.
echo Starting test application...
echo.

REM Check if executable exists
if not exist "upload_test.exe" (
    echo Building test application...
    go build -o upload_test main.go
    if errorlevel 1 (
        echo Failed to build test application
        pause
        exit /b 1
    )
)

REM Run the test
echo Running upload notifications test...
upload_test.exe

echo.
echo Test completed. Press any key to exit...
pause > nul