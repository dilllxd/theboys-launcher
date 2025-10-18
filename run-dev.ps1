# TheBoys Launcher Development Script (PowerShell)
Write-Host "Starting TheBoys Launcher in development mode..." -ForegroundColor Green
Write-Host "Note: Press Ctrl+C to stop the development server" -ForegroundColor Yellow
Write-Host ""

# Check if Wails is installed
try {
    wails version | Out-Null
} catch {
    Write-Host "Wails CLI not found. Installing Wails CLI..." -ForegroundColor Yellow
    try {
        go install github.com/wailsapp/wails/v2/cmd/wails@latest
        Write-Host "Wails CLI installed successfully!" -ForegroundColor Green
    } catch {
        Write-Host "Failed to install Wails CLI. Please make sure Go is installed and in your PATH." -ForegroundColor Red
        Read-Host "Press Enter to exit"
        exit 1
    }
}

# Install dependencies
Write-Host "Installing dependencies..." -ForegroundColor Blue
go mod download
go mod tidy
Set-Location frontend
npm install
Set-Location ..

# Start development server
Write-Host ""
Write-Host "Starting Wails development server..." -ForegroundColor Green
wails dev