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

## About the Development Setup

The development server uses a two-server approach to ensure everything works smoothly:

### How It Works
1. **Frontend Dev Server**: Runs on `http://localhost:5173` with hot reload
2. **Wails Backend**: Connects to the external frontend server
3. **Mock Bindings**: Preserved for development functionality

### Why This Approach
- ✅ **Preserves Bindings**: Mock bindings aren't overwritten by Wails
- ✅ **Better Hot Reload**: Frontend changes update instantly
- ✅ **Stable Development**: No binding generation issues
- ✅ **Full Functionality**: All features work during development

### Current Status: ✅ Ready for Development
The setup allows you to:
- ✅ Run the application with full UI
- ✅ Use all features with mock data
- ✅ Develop and test functionality
- ✅ Get instant hot reload on changes
- ✅ Debug and iterate quickly

The application will work perfectly for development purposes with this setup.

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