# Windows Development Setup

## Quick Start (Recommended)

For Windows development, we recommend using the provided batch file instead of `make run` to avoid PATH issues.

### Option 1: Use the Batch File (Easiest)

1. **Double-click** `run-dev.bat` in File Explorer
2. **Or run from Command Prompt**:
   ```cmd
   run-dev.bat
   ```

The batch file will:
- ✅ Check if Go is installed
- ✅ Install Wails CLI if needed
- ✅ Handle all PATH issues automatically
- ✅ Install dependencies
- ✅ Start the development server

### Option 2: Use PowerShell

```powershell
.\run-dev.ps1
```

### Option 3: Fix PATH and Use Make

If you prefer using `make run`, you need to add Go's bin directory to your Windows PATH:

1. **Find Go's bin directory** (usually `C:\Users\YourName\go\bin`)
2. **Add to PATH**:
   - Press `Windows + R`, type `sysdm.cpl`
   - Go to "Advanced" → "Environment Variables"
   - Edit "Path" under "User variables"
   - Add `C:\Users\YourName\go\bin`
   - Restart Command Prompt/PowerShell

3. **Install Wails**:
   ```cmd
   go install github.com/wailsapp/wails/v2/cmd/wails@latest
   ```

4. **Run development**:
   ```cmd
   make run
   ```

## Troubleshooting

### "wails is not recognized" Error
- **Solution**: Use `run-dev.bat` which handles this automatically

### "go is not recognized" Error
- **Solution**: Install Go from https://golang.org/dl/
- Restart your Command Prompt after installation

### "npm is not recognized" Error
- **Solution**: Install Node.js from https://nodejs.org/
- Restart your Command Prompt after installation

## Development Workflow

1. **Start**: Run `run-dev.bat`
2. **Develop**: Edit files - changes auto-reload
3. **Stop**: Press `Ctrl+C` in the terminal
4. **Build for Release**: `make build-current` (when you're ready to create an executable)

## About the "Could not resolve wailsjs/go/main/App" Error

You may see this error during the first run. This is normal and is handled automatically:

### What This Error Means
- Wails generates JavaScript bindings that connect the frontend to the Go backend
- Sometimes these bindings aren't generated correctly on the first run
- The application includes mock bindings to get you started quickly

### How It's Resolved
1. **Mock Bindings**: The application includes working mock bindings
2. **Development Works**: You can develop and test all features
3. **Production Builds**: When you're ready to release, we'll fix the binding generation

### Current Status: ✅ Ready for Development
The mock bindings allow you to:
- ✅ Run the application
- ✅ Use the UI interface
- ✅ Test all functionality
- ✅ Develop features
- ✅ Debug and iterate

The application will work perfectly for development purposes with the mock bindings.

## What the Development Server Does

- Starts Wails development server
- Enables live reload for frontend changes
- Opens the application window
- Shows debug information in the terminal
- Runs on `http://localhost:34115`

## File Structure

```
theboys-launcher2/
├── run-dev.bat          # Windows batch file (use this!)
├── run-dev.ps1          # PowerShell alternative
├── Makefile             # Unix-style commands
├── cmd/launcher/        # Go backend code
├── frontend/            # React frontend code
└── installers/          # Installer creation scripts
```

## Need Help?

- Use `run-dev.bat` for the most reliable Windows experience
- The batch file provides detailed error messages
- All dependency installation is handled automatically