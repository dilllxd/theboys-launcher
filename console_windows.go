//go:build windows

package main

import (
	"syscall"
)

func hideConsoleWindow() {
	// Try to completely detach from any console
	kernel32 := syscall.NewLazyDLL("kernel32.dll")

	// First try to free the console if we're attached to one
	freeConsole := kernel32.NewProc("FreeConsole")
	freeConsole.Call()

	// Try to find any remaining console window and hide it
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	hwnd, _, _ := getConsoleWindow.Call()

	if hwnd != 0 {
		// Console window exists, force hide it
		user32 := syscall.NewLazyDLL("user32.dll")
		showWindow := user32.NewProc("ShowWindow")
		const SW_HIDE = 0

		showWindow.Call(hwnd, uintptr(SW_HIDE))

		// Try to destroy the window completely
		destroyWindow := user32.NewProc("DestroyWindow")
		destroyWindow.Call(hwnd)
	}

	// Try freeing console again after hiding
	freeConsole.Call()
}
