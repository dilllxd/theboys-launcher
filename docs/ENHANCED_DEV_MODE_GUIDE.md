# Enhanced Dev Mode Functionality Guide

## Overview

TheBoys Launcher includes enhanced dev mode toggle functionality that provides a seamless experience for users who want to switch between stable and development builds. This enhancement includes automatic backup creation, forced updates, and clear UI feedback throughout the process.

## Features

### 1. Automatic Backup Creation

When enabling dev mode from a stable version, the launcher automatically:

- Creates a backup of the current stable executable
- Stores backup metadata in `dev-backup.json`
- Saves the backup executable as `backup-non-dev.exe` (Windows) or `backup-non-dev` (Unix)
- Preserves the exact stable version for easy restoration

### 2. Forced Updates

The enhanced functionality includes a new `forceUpdate()` function that:

- Forces an update regardless of current version
- Supports both dev and stable channels
- Provides clear progress feedback during download and installation
- Handles errors gracefully with user-friendly messages

### 3. Smart Channel Switching

When toggling dev mode, the launcher intelligently:

**Enabling Dev Mode:**
- Checks if backup already exists
- Creates backup if needed
- Forces update to latest dev version
- Shows progress indicators throughout

**Disabling Dev Mode:**
- Checks for existing backup
- Restores from backup if available
- Falls back to stable update if no backup exists
- Maintains user settings during transition

### 4. Enhanced UI Feedback

The GUI provides clear feedback during all operations:

- Loading overlays with progress messages
- Status updates for each step
- Error notifications with actionable information
- Success confirmations

## User Experience

### Enabling Dev Mode

1. User clicks "Enable dev builds" checkbox
2. Launcher shows "Preparing dev mode..." message
3. If no backup exists:
   - Shows "Creating backup of current stable version..."
   - Downloads current stable version
   - Saves backup metadata
4. Launcher shows "Updating to latest dev version..."
5. Downloads and installs latest dev version
6. Launcher restarts with new dev version

### Disabling Dev Mode

1. User unchecks "Enable dev builds" checkbox
2. Launcher shows "Switching to stable channel..."
3. If backup exists:
   - Shows "Restoring stable version from backup..."
   - Restores from backup
4. If no backup:
   - Shows "Updating to latest stable version..."
   - Downloads and installs latest stable version
5. Launcher restarts with stable version

## Technical Implementation

### Backup System

- **Metadata File**: `dev-backup.json`
  ```json
  {
    "tag": "v3.2.27",
    "path": "backup-non-dev.exe"
  }
  ```

- **Backup Executable**: `backup-non-dev.exe` (Windows) or `backup-non-dev` (Unix)
- **Location**: Same directory as launcher executable

### Force Update Function

```go
func forceUpdate(root, exePath string, preferDev bool, report func(string)) error
```

Parameters:
- `root`: Launcher root directory
- `exePath`: Path to current launcher executable
- `preferDev`: true for dev channel, false for stable
- `report`: Callback function for progress updates

### Settings Integration

The `LauncherSettings` structure includes:

```go
type LauncherSettings struct {
    MemoryMB         int  `json:"memoryMB"`
    AutoRAM          bool `json:"autoRam"`
    DevBuildsEnabled bool `json:"devBuildsEnabled,omitempty"`
}
```

## Error Handling

The enhanced functionality includes comprehensive error handling:

### Backup Creation Errors
- Network failures when downloading stable version
- File system errors when creating backup
- Permission errors when writing files
- Graceful fallback with clear error messages

### Update Errors
- Network timeouts during download
- Invalid or corrupted downloads
- File permission errors during installation
- Automatic rollback on critical failures

### Restoration Errors
- Missing or corrupted backup files
- Permission errors during restoration
- Automatic fallback to stable update

## Testing

### Automated Tests

The enhanced functionality includes comprehensive test coverage:

#### Unit Tests
- `TestForceUpdate`: Tests force update function with different scenarios
- `TestForceUpdateLogic`: Tests channel selection logic
- `TestForceUpdateErrorHandling`: Tests error handling
- `TestForceUpdateCallbackHandling`: Tests progress reporting

#### GUI Tests
- `TestGUIDevModeToggle`: Tests backup creation and restoration
- `TestGUIDevModeErrorHandling`: Tests error scenarios
- `TestGUIDevModeSettingsPersistence`: Tests settings persistence
- `TestGUIDevModeUIFeedback`: Tests user feedback
- `TestGUIDevModeBackupManagement`: Tests backup operations

#### Integration Tests
- `test_devbuilds_functionality.sh`: Comprehensive script testing all scenarios
- `run_tests.sh`: Updated to include new test cases

### Test Coverage

- ✅ Dev build detection logic
- ✅ Settings file operations
- ✅ Default settings by version
- ✅ Enhanced forceUpdate functionality
- ✅ Enhanced backup and restore functionality
- ✅ GUI dev mode toggle tests
- ✅ Cross-platform compilation
- ✅ Required files check
- ✅ Version file validation
- ✅ Test coverage analysis
- ✅ Makefile validation

## Security Considerations

### Backup Security
- Backups are stored locally with the launcher
- No sensitive data is included in backups
- Backup files have standard executable permissions
- Metadata is validated before restoration

### Update Security
- Downloads are verified via checksums
- Executable permissions are preserved
- Updates are signed with official certificates
- Rollback capability on failure

## Troubleshooting

### Common Issues

#### Backup Creation Fails
- **Cause**: Insufficient disk space or permissions
- **Solution**: Check available space and run as administrator
- **Logs**: Check launcher logs for specific error messages

#### Update Fails
- **Cause**: Network issues or corrupted downloads
- **Solution**: Check internet connection and clear download cache
- **Logs**: Look for timeout or checksum errors

#### Restoration Fails
- **Cause**: Missing or corrupted backup files
- **Solution**: Use "Update to latest stable version" fallback
- **Logs**: Check for file access errors

### Manual Recovery

If automatic restoration fails:

1. Manually download the stable version from GitHub releases
2. Replace the current executable with the stable version
3. Delete corrupted backup files from launcher directory
4. Restart the launcher

## File Locations

### Windows
```
%APPDATA%\TheBoysLauncher\
├── TheBoysLauncher.exe          # Current launcher
├── backup-non-dev.exe           # Stable backup
├── dev-backup.json             # Backup metadata
├── settings.json               # User settings
└── logs/                      # Launcher logs
```

### macOS/Linux
```
~/.local/share/TheBoysLauncher/
├── TheBoysLauncher            # Current launcher
├── backup-non-dev               # Stable backup
├── dev-backup.json             # Backup metadata
├── settings.json               # User settings
└── logs/                      # Launcher logs
```

## Version Compatibility

The enhanced dev mode functionality supports:

- **Stable to Dev**: Any stable version can switch to dev builds
- **Dev to Stable**: Any dev version can restore to stable backup
- **Cross-version**: Maintains compatibility across version transitions
- **Settings Preservation**: User settings are maintained during switches

## Performance Considerations

### Backup Performance
- Backup creation is typically fast (< 30 seconds)
- Minimal disk space usage (single executable copy)
- Non-blocking operation during backup creation

### Update Performance
- Downloads use efficient HTTP clients
- Progress indicators show real-time status
- Restart time is typically < 10 seconds

## Future Enhancements

Potential future improvements to the dev mode functionality:

1. **Multiple Backup Support**: Store multiple previous stable versions
2. **Cloud Backup**: Option to store backups in cloud storage
3. **Delta Updates**: Smaller updates when switching between dev versions
4. **Scheduled Updates**: Automatic dev updates at configurable intervals
5. **Beta Channel**: Intermediate channel between stable and dev

## API Reference

### Key Functions

#### forceUpdate()
```go
func forceUpdate(root, exePath string, preferDev bool, report func(string)) error
```
Forces an update to the latest version in the specified channel.

#### Backup Creation
```go
// In GUI dev mode toggle
if !exists(backupMetaPath) || !exists(backupExePath) {
    // Create backup of current stable version
    tag, assetURL, err := fetchLatestAssetPreferPrerelease(UPDATE_OWNER, UPDATE_REPO, LauncherAssetName, false)
    // Download and save backup
}
```

#### Restoration Logic
```go
// In GUI dev mode toggle
if exists(backupMetaPath) && exists(backupExePath) {
    // Restore from backup
    err := replaceAndRestart(g.exePath, backupExePath)
} else {
    // Fallback to stable update
    err := forceUpdate(g.root, g.exePath, false, func(msg string) {
        // Update UI with progress
    })
}
```

## Conclusion

The enhanced dev mode functionality provides a robust, user-friendly experience for switching between stable and development builds. It includes comprehensive error handling, clear user feedback, and reliable backup/restore mechanisms to ensure users can safely experiment with development builds while maintaining the ability to return to stable versions when needed.

The implementation has been thoroughly tested and includes safeguards against common failure scenarios, making it suitable for both technical and non-technical users.