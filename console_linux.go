//go:build linux
// +build linux

package main

// hideConsoleWindow on Linux is a no-op
// Linux GUI applications don't typically show console windows when launched properly
func hideConsoleWindow() {
	// No implementation needed for Linux
	// Console windows are not typically shown for GUI apps on Linux
}
