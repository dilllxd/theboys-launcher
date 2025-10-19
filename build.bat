@echo off
echo Building TheBoys Launcher with icon and version info...

REM Set version from command line or use default
set VERSION=v2.0.3
if not "%1"=="" set VERSION=%1

REM Check if rsrc tool is available
where rsrc >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Installing rsrc tool...
    go install github.com/akavel/rsrc@latest
)

REM Check if icon file exists
if not exist "icon.ico" (
    echo ERROR: icon.ico not found!
    echo Please create/place an icon.ico file in this directory.
    echo See ICON_README.md for details.
    pause
    exit /b 1
)

REM Update version in resource.rc dynamically
echo Updating version info...
set VERSION_CLEAN=%VERSION:v=%
powershell.exe -Command "& {(Get-Content 'resource.rc') -replace '2,0,3,0', '%VERSION_CLEAN:,0,0,0%' -replace '\"2.0.3\"', '\"%VERSION_CLEAN%\"' | Out-File 'resource.rc' -Encoding UTF8}"

REM Compile resources
echo Compiling resources...
rsrc -ico icon.ico -manifest resource.rc -o resource.syso

REM Build the application with embedded resources
echo Building executable...
go build -ldflags="-s -w -H=windowsgui -X main.version=%VERSION%" -o TheBoysLauncher.exe .

if %ERRORLEVEL% EQU 0 (
    echo Build successful!
    echo.
    echo Output: TheBoysLauncher.exe
    echo Version: %VERSION%
    echo Icon: embedded
    echo Version info: embedded
) else (
    echo Build failed!
    pause
    exit /b 1
)