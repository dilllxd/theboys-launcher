# Manual Testing Guide for Dev Builds Toggle Functionality

This document provides step-by-step instructions for manually testing the dev builds toggle functionality in TheBoys Launcher GUI.

## Overview

The dev builds toggle allows users to:
1. Enable or disable pre-release/dev build updates through the settings window
2. Have their preference persist across launcher restarts
3. Ensure automatic override logic doesn't interfere with user preferences
4. Get dev builds enabled by default when running a dev version (new installations only)

## Prerequisites

1. Build the launcher with the current changes:
   ```bash
   go build -o TheBoysLauncher .
   ```

2. Have a dev version available for testing (either build with `-ldflags="-X main.version=dev"` or use an existing dev build)

## Test Scenarios

### Test 1: New Installation with Dev Version

**Objective**: Verify that new installations with a dev version enable dev builds by default

**Steps**:
1. Remove or rename any existing `settings.json` file from your launcher data directory
   - Windows: `%APPDATA%\theboyslauncher\settings.json`
   - macOS: `~/Library/Application Support/theboyslauncher/settings.json`
   - Linux: `~/.config/theboyslauncher/settings.json`

2. Launch the dev version of TheBoysLauncher

3. Open the Settings window (click Settings button in sidebar)

4. **Expected Result**: The "Enable dev builds (pre-release)" checkbox should be checked

5. Check the log output or console for message:
   ```
   New installation detected with dev build (version: dev), dev builds enabled by default
   ```

### Test 2: New Installation with Stable Version

**Objective**: Verify that new installations with a stable version disable dev builds by default

**Steps**:
1. Remove any existing `settings.json` file

2. Launch a stable version of TheBoysLauncher

3. Open the Settings window

4. **Expected Result**: The "Enable dev builds (pre-release)" checkbox should be unchecked

### Test 3: User Preference Persistence

**Objective**: Verify that user's dev builds preference persists across launcher restarts

**Steps**:
1. Launch TheBoysLauncher

2. Open Settings window

3. Toggle the "Enable dev builds (pre-release)" checkbox:
   - If checked, uncheck it
   - If unchecked, check it

4. Click "Save" button

5. Close the launcher completely

6. Relaunch TheBoysLauncher

7. Open Settings window again

8. **Expected Result**: The checkbox should reflect your previous choice (not the default)

9. Verify `settings.json` contains your choice:
   ```bash
   # Check the devBuildsEnabled field
   cat ~/.config/theboyslauncher/settings.json  # Linux
   # or check the appropriate path for your OS
   ```

### Test 4: Update Logic Respects User Preference

**Objective**: Verify that the update check respects the user's dev builds setting

**Steps**:
1. Launch TheBoysLauncher

2. Open Settings and ensure "Enable dev builds (pre-release)" is unchecked

3. Save settings

4. Trigger an update check (happens automatically on startup)

5. **Expected Result**: Launcher should only check for stable releases

6. Repeat with "Enable dev builds (pre-release)" checked

7. **Expected Result**: Launcher should check for both stable and dev releases, preferring dev releases

### Test 5: Backup and Restore Functionality

**Objective**: Verify the backup/restore feature works when enabling/disabling dev builds

**Steps**:
1. Launch TheBoysLauncher with a stable version

2. Open Settings and enable "Enable dev builds (pre-release)"

3. **Expected Result**: 
   - Should see a loading indicator
   - A backup of the stable executable should be created
   - Settings should show "Channel: Dev (enabled)"
   - Backup information should be displayed

4. Disable dev builds

5. **Expected Result**: 
   - Should prompt to restore the previous stable version
   - Settings should show "Channel: Stable"
   - Backup information should remain visible

6. Test the "Restore backup now" and "Delete backup" buttons

## Verification Checklist

For each test scenario, verify:

- [ ] GUI checkbox state matches the underlying setting
- [ ] Setting persists after launcher restart
- [ ] `settings.json` file contains the correct `devBuildsEnabled` value
- [ ] Update check behavior changes based on the setting
- [ ] Log messages indicate the correct behavior
- [ ] No automatic override of user preferences occurs

## Expected Log Messages

### New Dev Installation:
```
New installation detected with dev build (version: dev), dev builds enabled by default
```

### New Stable Installation:
```
(no special dev build message)
```

### User Enables Dev Builds:
```
GUI: User enabled dev builds setting
GUI: Dev builds setting enabled and saved
```

### User Disables Dev Builds:
```
GUI: User disabled dev builds setting
GUI: Dev builds setting disabled and saved
```

### Dev Build Detected with User Preference:
```
Dev build detected (version: dev), dev builds already enabled by user preference
```
or
```
Dev build detected (version: dev), dev builds disabled by user preference
```

## Troubleshooting

### Issue: Checkbox doesn't reflect actual setting
- Check `settings.json` for correct `devBuildsEnabled` value
- Verify file permissions allow writing
- Check for multiple launcher instances

### Issue: Setting doesn't persist
- Verify `settings.json` is being written
- Check for file system errors in logs
- Ensure no other process is modifying the file

### Issue: Update check ignores setting
- Verify the setting is loaded before update check
- Check network connectivity
- Verify GitHub releases are accessible

## Automated Test Results

The automated tests (`test_devbuilds_functionality.ps1` or `test_devbuilds_functionality.sh`) verify:

- [x] Dev build detection logic works correctly
- [x] Settings file operations (create, read, update)
- [x] Default settings by version type
- [x] Integration test setup

These tests should all pass before proceeding with manual GUI testing.