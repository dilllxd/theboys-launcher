package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// -------------------- Downloads / Unzip --------------------

type progressWriter struct {
	total      int64
	downloaded int64
	filename   string
	startTime  time.Time
}

func (pw *progressWriter) Write(p []byte) (int, error) {
	n := len(p)
	err := error(nil)
	pw.downloaded += int64(n)

	// Update progress every 1MB or every second
	if pw.downloaded%1048576 == 0 || time.Since(pw.startTime) > time.Second {
		pw.updateProgress()
		pw.startTime = time.Now()
	}

	return n, err
}

func (pw *progressWriter) updateProgress() {
	if pw.total > 0 {
		percent := float64(pw.downloaded) / float64(pw.total) * 100

		// Calculate download speed
		elapsed := time.Since(pw.startTime).Seconds()
		if elapsed > 0 {
			speedMBps := (float64(pw.downloaded) / (1024 * 1024)) / elapsed
			fmt.Fprintf(out, "\rDownloading %s (%.1f MB/s, %d%%)", pw.filename, speedMBps, int(percent))
		} else {
			fmt.Fprintf(out, "\rDownloading %s (%d%%)", pw.filename, int(percent))
		}
	}
}

func downloadTo(url, path string, mode os.FileMode) error {
	b, err := downloadWithProgress(url)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, mode)
}

func downloadAndUnzipTo(url, dest string) error {
	b, err := download(url)
	if err != nil {
		return err
	}
	return extractBytesTo(b, dest, url)
}

// extractBytesTo extracts archive bytes to destination, detecting format
func extractBytesTo(b []byte, dest, url string) error {
	// Determine format based on file extension and platform
	if strings.HasSuffix(strings.ToLower(url), ".zip") {
		return unzipBytesTo(b, dest)
	} else if strings.HasSuffix(strings.ToLower(url), ".tar.gz") || strings.HasSuffix(strings.ToLower(url), ".tgz") {
		return untarBytesTo(b, dest)
	} else {
		// Fallback: try zip first (Windows), then tar.gz (macOS/Linux)
		if err := unzipBytesTo(b, dest); err != nil {
			if runtime.GOOS != "windows" {
				return untarBytesTo(b, dest)
			}
			return err
		}
		return nil
	}
}

// untarBytesTo extracts a .tar.gz archive to destination
func untarBytesTo(b []byte, dest string) error {
	// First decompress gzip
	gzReader, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Then extract tar
	tarReader := tar.NewReader(gzReader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		targetPath := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// Create directory
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}

		case tar.TypeReg:
			// Create file
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", targetPath, err)
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", targetPath, err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("failed to write file %s: %w", targetPath, err)
			}
			outFile.Close()

		default:
			// Skip other file types (symlinks, etc.)
			continue
		}
	}

	return nil
}

func download(url string) ([]byte, error) {
	return downloadWithProgress(url)
}

func downloadWithProgress(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", getUserAgent("General"))
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, url)
	}

	// Get file info for progress
	contentLength := resp.ContentLength
	filename := filepath.Base(url)

	// Create progress writer
	pw := &progressWriter{
		total:     contentLength,
		filename:  filename,
		startTime: time.Now(),
	}

	// If we don't know the content length, show indefinite progress
	if contentLength <= 0 {
		fmt.Fprintf(out, "Downloading %s...", filename)
		return io.ReadAll(resp.Body)
	}

	// Read with progress tracking
	body, err := io.ReadAll(io.TeeReader(resp.Body, pw))
	if err != nil {
		return nil, err
	}

	// Show completion
	fmt.Fprintf(out, "\nDownloaded %s (%.1f MB)\n", filename, float64(contentLength)/(1024*1024))

	return body, nil
}

func unzipBytesTo(b []byte, dest string) error {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return err
	}
	for _, f := range r.File {
		p := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		outf, err := os.Create(p)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(outf, rc)
		outf.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
