# Legacy Code

This directory contains the original Winterpack Launcher code before the rebuild.

## Files

- `main.go` - Original single-file implementation (Windows-only, TUI-based)
- `go.mod` - Original Go module configuration
- `go.sum` - Original dependency checksums
- `modpacks.json` - Original modpack configuration

## Architecture Notes

The legacy code was a single-file Go application with the following characteristics:
- Windows-only support
- Terminal User Interface (TUI) using Bubble Tea
- All functionality in one monolithic main.go file
- Direct Windows API calls embedded in main code
- Hard-coded platform assumptions

## Key Functionality

- Self-updating from GitHub Releases
- Prism Launcher integration
- Dynamic Java runtime management
- Modpack selection and management
- Console-based user interface

## Status

This code is preserved for reference but should not be used for active development.
See the parent directory for the new cross-platform GUI implementation.