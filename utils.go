package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func modpackLabel(mp Modpack) string {
	if name := strings.TrimSpace(mp.DisplayName); name != "" {
		return name
	}
	return mp.ID
}

func slugifyID(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return "modpack"
	}
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	result := strings.Trim(b.String(), "-")
	if result == "" {
		return "modpack"
	}
	return result
}

func versionFileNameFor(mp Modpack) string { return "." + slugifyID(mp.ID) + "-version" }
func backupPrefixFor(mp Modpack) string    { return slugifyID(mp.ID) + "-backup-" }

// -------------------- Helpers --------------------

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func flattenOneLevel(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return nil
	}
	top := filepath.Join(dir, entries[0].Name())
	items, err := os.ReadDir(top)
	if err != nil {
		return err
	}
	for _, it := range items {
		src := filepath.Join(top, it.Name())
		dst := filepath.Join(dir, it.Name())
		_ = os.Rename(src, dst)
	}
	_ = os.Remove(top)
	return nil
}

// -------------------- Logging --------------------

func setupLogging(root string) func() {
	logDir := filepath.Join(root, "logs")
	_ = os.MkdirAll(logDir, 0755)
	latest := filepath.Join(logDir, "latest.log")
	prev := filepath.Join(logDir, "previous.log")

	// rotate previous
	if _, err := os.Stat(latest); err == nil {
		_ = os.Remove(prev)         // best-effort
		_ = os.Rename(latest, prev) // best-effort
	}

	f, err := os.OpenFile(latest, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		// fall back to console only
		out = os.Stdout
		return func() {}
	}

	// Mirror to console + file
	out = io.MultiWriter(os.Stdout, f)

	return func() { _ = f.Close() }
}

func logf(format string, a ...any) {
	fmt.Fprintf(out, format+"\n", a...)
}

func pauseIfWanted() {
	if os.Getenv(envNoPause) == "1" {
		return
	}
	fmt.Fprint(out, "\nPress Enter to exit…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

func fail(err error) {
	fmt.Fprintf(out, "Error: %v\n", err)
	pauseIfWanted()
	os.Exit(1)
}

// A neutral "press enter" we can use inside flows without saying "exit"
func waitEnter() {
	fmt.Fprint(out, "\nPress Enter to continue…")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
}

// Returns true for yes, false for no (console version)
func yesNoBox(text, title string) int {
	const (
		IDYES = 6
		IDNO  = 7
	)

	fmt.Printf("\n%s\n", headerLine(title))
	fmt.Printf("%s\n", text)
	fmt.Printf("Continue? (y/n): ")

	var response string
	fmt.Scanln(&response)

	response = strings.ToLower(strings.TrimSpace(response))
	if response == "y" || response == "yes" {
		return IDYES
	}
	return IDNO
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// -------------------- Launcher Options --------------------

type launcherOptions struct {
	useCLI            bool
	modpackID         string
	listModpacks      bool
	openSettings      bool
	cleanupAfterUpdate bool
	cleanupOldExe     string
	cleanupNewExe     string
}

func parseOptions() launcherOptions {
	opts := launcherOptions{}

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options]\n\nOptions:\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}

	flag.BoolVar(&opts.useCLI, "cli", false, "Run the launcher in plain console mode (skip the interactive TUI).")
	flag.StringVar(&opts.modpackID, "modpack", "", "Select the modpack to launch by ID.")
	flag.BoolVar(&opts.listModpacks, "list-modpacks", false, "List available modpacks and exit.")
	flag.BoolVar(&opts.openSettings, "settings", false, "Open launcher settings menu.")

	// Handle cleanup after update (internal flag, not shown in help)
	if len(os.Args) >= 2 && os.Args[1] == "--cleanup-after-update" && len(os.Args) >= 4 {
		opts.cleanupAfterUpdate = true
		opts.cleanupOldExe = os.Args[2]
		opts.cleanupNewExe = os.Args[3]
		return opts
	}

	flag.Parse()

	return opts
}