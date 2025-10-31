package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestMultipartUpload tests uploading to the actual logs.dylan.lol endpoint using multipart format
func TestMultipartUpload(t *testing.T) {
	// Create a temporary test file
	tempDir, err := os.MkdirTemp("", "multipart-upload-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testLogPath := filepath.Join(tempDir, "test.log")
	testContent := "This is a test log file for multipart upload testing.\nTimestamp: " + time.Now().Format(time.RFC3339) + "\n"
	err = os.WriteFile(testLogPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Read the test file
	content, err := os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Test with multipart format (matching Chrome request)
	t.Run("MultipartForm", func(t *testing.T) {
		testMultipartFormUpload(t, content)
	})

	// Test with different expiration values
	t.Run("MultipartFormNeverExpiration", func(t *testing.T) {
		testMultipartFormUploadWithExpiration(t, content, "never")
	})

	t.Run("MultipartForm24HourExpiration", func(t *testing.T) {
		testMultipartFormUploadWithExpiration(t, content, "24hour")
	})
}

func testMultipartFormUpload(t *testing.T, content []byte) {
	fmt.Println("Testing multipart form upload...")

	// Create multipart form with the required fields (matching Chrome request)
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add form fields (matching the working Chrome request)
	writer.WriteField("expiration", "24hour")
	writer.WriteField("syntax_highlight", "none")
	writer.WriteField("privacy", "public")
	writer.WriteField("content", string(content))

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Headers: %+v\n", resp.Header)
	fmt.Printf("Response Body (first 500 chars): %s\n", string(body)[:minInt(500, len(body))])

	// Check if we got a successful response
	if resp.StatusCode == 302 {
		t.Logf("Upload successful! Got 302 redirect")
		location := resp.Header.Get("Location")
		if location != "" {
			t.Logf("Redirect location: %s", location)
		}
	} else if resp.StatusCode == 200 {
		t.Logf("Upload successful! Got 200 OK")

		// Try to parse as JSON
		var result struct {
			OK  bool   `json:"ok"`
			URL string `json:"url"`
		}

		if err := json.Unmarshal(body, &result); err == nil && result.OK {
			t.Logf("Successfully parsed JSON response: %+v", result)
			t.Logf("Upload successful! URL: %s", result.URL)
		} else {
			t.Logf("Response is not JSON or parsing failed")
		}
	} else {
		t.Logf("Upload failed with status: %s", resp.Status)
		t.Logf("Response body: %s", string(body))
	}
}

func testMultipartFormUploadWithExpiration(t *testing.T, content []byte, expiration string) {
	fmt.Printf("Testing multipart form upload with expiration='%s'...\n", expiration)

	// Create multipart form with the required fields
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add form fields
	writer.WriteField("expiration", expiration)
	writer.WriteField("syntax_highlight", "none")
	writer.WriteField("privacy", "public")
	writer.WriteField("content", string(content))

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body (first 200 chars): %s\n", string(body)[:min(200, len(body))])

	// Check if we got a successful response
	if resp.StatusCode == 302 || resp.StatusCode == 200 {
		t.Logf("Upload successful with expiration='%s'!", expiration)
	} else {
		t.Logf("Upload failed with expiration='%s': %s", expiration, resp.Status)
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
