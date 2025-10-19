//go:build darwin
// +build darwin

package main

// hideConsoleWindow on macOS is a no-op
// macOS GUI applications don't show console windows by default
func hideConsoleWindow() {
	// No implementation needed for macOS
	// GUI apps on macOS don't typically show console windows
}