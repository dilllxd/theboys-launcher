# Enhanced Dev Mode Toggle Behavior Guide

## Overview

TheBoys Launcher has been enhanced with a new dev mode toggle behavior that provides better user experience and prevents unintended updates. This guide explains the modified behavior and how it works.

## Key Changes

### 1. Checkbox Behavior Changes

**Previous Behavior:**
- Checkbox toggle immediately triggered backup creation and launcher updates
- Multiple rapid toggles could cause multiple update processes
- No way to preview changes before applying them
- Canceling dialog after checking checkbox left settings in inconsistent state

**New Behavior:**
- Checkbox toggle only updates temporary in-memory variables
- No backup or update processes are triggered by checkbox changes
- Users can toggle checkbox multiple times without side effects
- Settings are only applied when user explicitly clicks "Save"

### 2. Settings Dialog Workflow

The settings dialog now follows this enhanced workflow:

1. **Initial State**: Dialog opens with current settings loaded
2. **Temporary Variables**: Checkbox changes update `pendingDevBuildsEnabled` variable only
3. **Status Feedback**: Real-time status label shows pending changes
4. **Save Action**: Clicking "Save" applies all pending changes at once
5. **Cancel Action**: Clicking "Cancel" discards all pending changes
6. **Close Detection**: Dialog warns user if there are unsaved changes

### 3. UI Feedback System

#### Status Label Updates

The status label provides clear feedback about pending changes:

- **No Changes**: Status label is empty
- **Enabling Dev Mode**: "Pending changes: Dev builds will be enabled (backup will be created and launcher will update)"
- **Disabling Dev Mode**: "Pending changes: Dev builds will be disabled (stable version will be restored)"
- **Multiple Changes**: Lists all pending changes separated by commas

#### Visual Indicators

- **Checkbox State**: Reflects user's intended change (not current saved state)
- **Save Button**: Enabled when there are pending changes
- **Cancel Button**: Warns user if there are unsaved changes
- **Loading Overlay**: Shows specific messages during update process

### 4. Update Process Optimization

#### Single Update Execution

The update process now runs only once when settings are applied:

1. **Backup Creation**: Only when enabling dev mode and no backup exists
2. **Settings Save**: All changes are saved atomically
3. **Launcher Update**: Single update process to target version
4. **Error Handling**: Automatic rollback on failure with user notification

#### Error Recovery

- **Backup Failure**: Checkbox state is reverted, user notified
- **Update Failure**: Settings remain unchanged, user notified
- **Network Issues**: Graceful fallback with clear error messages

## User Experience Improvements

### 1. Predictable Behavior

- **No Surprises**: Updates only happen when user explicitly saves
- **Preview Changes**: Users can see exactly what will change before applying
- **Multiple Adjustments**: Users can fine-tune settings without triggering updates

### 2. Performance Benefits

- **Reduced Network Calls**: No unnecessary update requests
- **Faster UI Response**: Checkbox toggles are instant
- **Resource Efficiency**: Single backup/update process instead of multiple

### 3. Safety Features

- **Atomic Changes**: All settings applied together or not at all
- **Rollback Capability**: Failed updates don't leave launcher in broken state
- **Confirmation Dialogs**: Users must confirm potentially disruptive operations

## Technical Implementation

### 1. Variable Management

```go
// Original settings (persisted)
originalDevBuildsEnabled := settings.DevBuildsEnabled

// Temporary variables (for UI state)
pendingDevBuildsEnabled := settings.DevBuildsEnabled

// Checkbox changes only affect pending variables
devCheck.OnChanged = func(on bool) {
    pendingDevBuildsEnabled = on
    updateStatusLabel()
}
```

### 2. Status Update Logic

```go
func updateStatusLabel() {
    hasPendingChanges := originalDevBuildsEnabled != pendingDevBuildsEnabled ||
                      originalAutoRAM != pendingAutoRAM ||
                      originalMemoryMB != pendingMemoryMB
    
    if hasPendingChanges {
        var changes []string
        if originalDevBuildsEnabled != pendingDevBuildsEnabled {
            if pendingDevBuildsEnabled {
                changes = append(changes, "Dev builds will be enabled (backup will be created and launcher will update)")
            } else {
                changes = append(changes, "Dev builds will be disabled (stable version will be restored)")
            }
        }
        statusLabel.SetText("Pending changes: " + strings.Join(changes, ", "))
    } else {
        statusLabel.SetText("")
    }
}
```

### 3. Save Process

```go
saveBtn.OnTapped = func() {
    if hasPendingChanges() {
        // Apply all pending changes at once
        settings.DevBuildsEnabled = pendingDevBuildsEnabled
        settings.AutoRAM = pendingAutoRAM
        settings.MemoryMB = pendingMemoryMB
        
        // Handle dev mode specific operations
        if originalDevBuildsEnabled != pendingDevBuildsEnabled {
            if pendingDevBuildsEnabled {
                createBackupAndEnableDevMode()
            } else {
                restoreStableVersion()
            }
        }
        
        // Save settings to disk
        saveSettings(root)
        updateMemorySummaryLabel()
    }
}
```

## Testing Scenarios

### 1. Basic Toggle Test

1. Open settings dialog
2. Toggle dev mode checkbox (check/uncheck)
3. Verify status label updates accordingly
4. Click "Cancel"
5. Reopen settings dialog
6. Verify original setting is still active

### 2. Multiple Toggle Test

1. Open settings dialog
2. Toggle dev mode checkbox 3-4 times rapidly
3. Verify no update processes are triggered
4. Verify status label reflects current pending state
5. Click "Save"
6. Verify single update process occurs

### 3. Save and Cancel Test

1. Open settings dialog
2. Toggle dev mode checkbox
3. Click "Save"
4. Verify update process runs
5. Reopen settings dialog
6. Verify new setting is active
7. Toggle checkbox again
8. Click "Cancel"
9. Reopen settings dialog
10. Verify setting was not changed

## Troubleshooting

### Common Issues

**Issue**: Settings don't save after clicking "Save"
- **Solution**: Check file permissions in launcher directory
- **Log**: Look for "Failed to save settings" errors

**Issue**: No backup created when enabling dev mode
- **Solution**: Ensure sufficient disk space and write permissions
- **Log**: Look for "Failed to create backup" errors

**Issue**: Status label doesn't update
- **Solution**: This is a UI refresh issue, restart launcher
- **Log**: Look for UI-related errors

### Debug Information

Enable debug logging to troubleshoot issues:

1. Check launcher logs for "GUI: User enabled/disabled dev builds setting"
2. Verify backup file creation in launcher directory
3. Confirm settings.json contains expected values
4. Check for network connectivity during update processes

## Migration Notes

### For Users Upgrading from Previous Versions

- **No Action Required**: Existing settings are preserved
- **Enhanced Experience**: New dialog behavior is automatic
- **Backward Compatibility**: All existing functionality remains

### For Developers

- **Testing**: Use `TestDevModeToggleBehavior` test suite
- **UI Components**: All checkbox changes go through `pending*` variables
- **Settings Persistence**: Only save button triggers `saveSettings()`

## Conclusion

The enhanced dev mode toggle behavior provides a more predictable, safer, and user-friendly experience. Users can now experiment with settings without fear of unintended updates, and the clear feedback system ensures they always know what changes will be applied.

The implementation maintains full backward compatibility while adding significant improvements to the user experience and system reliability.