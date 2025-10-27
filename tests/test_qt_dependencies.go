//go:build linux
// +build linux

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("=== Qt Dependency Test ===")

	// Test Qt dependency detection
	depInfo := checkQtLibraries()

	fmt.Printf("Installed: %v\n", depInfo.Installed)
	if !depInfo.Installed {
		fmt.Printf("Missing libraries: %v\n", depInfo.MissingLibs)
		fmt.Printf("Package manager: %s\n", depInfo.PackageManager)
		fmt.Printf("Packages to install: %v\n", depInfo.Packages)

		// Test package manager detection
		pm := getPackageManager()
		if pm != nil {
			fmt.Printf("Detected package manager: %s\n", pm.Name)
		} else {
			fmt.Println("No supported package manager detected")
		}
	} else {
		fmt.Println("All Qt dependencies are installed!")
	}

	fmt.Println("=== Test Complete ===")
	os.Exit(0)
}
