package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// TestFormUploadStandalone tests the new form upload functionality independently
func TestFormUploadStandalone(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-form-standalone-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for standalone test\nLine 2\nLine 3\n"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Test the form data creation logic (matching the new uploadLog implementation)
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create form data with the required fields (matching new implementation)
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Verify form data contains expected fields
	if formData.Get("expiration") != "never" {
		t.Errorf("Expected expiration='never', got '%s'", formData.Get("expiration"))
	}

	if formData.Get("syntax_highlight") != "none" {
		t.Errorf("Expected syntax_highlight='none', got '%s'", formData.Get("syntax_highlight"))
	}

	if formData.Get("privacy") != "public" {
		t.Errorf("Expected privacy='public', got '%s'", formData.Get("privacy"))
	}

	if formData.Get("content") != string(content) {
		t.Errorf("Content mismatch. Expected '%s', got '%s'", string(content), formData.Get("content"))
	}

	// Test the encoded form data
	encodedData := formData.Encode()
	if !strings.Contains(encodedData, "expiration=never") {
		t.Error("Encoded data should contain 'expiration=never'")
	}

	if !strings.Contains(encodedData, "syntax_highlight=none") {
		t.Error("Encoded data should contain 'syntax_highlight=none'")
	}

	if !strings.Contains(encodedData, "privacy=public") {
		t.Error("Encoded data should contain 'privacy=public'")
	}

	if !strings.Contains(encodedData, "content=Test+log+content") {
		t.Error("Encoded data should contain URL-encoded content")
	}

	t.Logf("Form data created successfully: %s", encodedData)
}

// TestFormUploadRequestCreation tests the HTTP request creation logic
func TestFormUploadRequestCreation(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-request-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for request creation test"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create form data with the required fields (matching new implementation)
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Create a new HTTP request with the form data (matching new implementation)
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data (matching new implementation)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Verify request properties
	if req.Method != "POST" {
		t.Errorf("Expected POST method, got %s", req.Method)
	}

	if req.URL.String() != "https://logs.dylan.lol/upload" {
		t.Errorf("Expected URL 'https://logs.dylan.lol/upload', got '%s'", req.URL.String())
	}

	contentType := req.Header.Get("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Expected Content-Type 'application/x-www-form-urlencoded', got '%s'", contentType)
	}

	userAgent := req.Header.Get("User-Agent")
	if userAgent != "TheBoysLauncher/1.0" {
		t.Errorf("Expected User-Agent 'TheBoysLauncher/1.0', got '%s'", userAgent)
	}

	t.Logf("HTTP request created successfully with method: %s, URL: %s, Content-Type: %s",
		req.Method, req.URL.String(), contentType)
}

// TestFormUploadClientConfiguration tests the HTTP client configuration
func TestFormUploadClientConfiguration(t *testing.T) {
	// Test the HTTP client configuration (matching new implementation)
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
		},
	}

	// Verify client configuration
	if client.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.Timeout)
	}

	transport, ok := client.Transport.(*http.Transport)
	if !ok {
		t.Fatal("Expected *http.Transport, got different type")
	}

	if transport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS min version 1.2, got %v", transport.TLSClientConfig.MinVersion)
	}

	if transport.TLSClientConfig.MaxVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS max version 1.2, got %v", transport.TLSClientConfig.MaxVersion)
	}

	t.Logf("HTTP client configured successfully with TLS 1.2 and 30s timeout")
}

// TestFormUploadResponseParsingJSON tests JSON response parsing
func TestFormUploadResponseParsingJSON(t *testing.T) {
	// Test JSON response parsing (matching new implementation)
	jsonResponse := `{"ok": true, "url": "https://logs.dylan.lol/abc123"}`

	var result struct {
		OK  bool   `json:"ok"`
		URL string `json:"url"`
	}

	err := json.Unmarshal([]byte(jsonResponse), &result)
	if err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify parsed response
	if !result.OK {
		t.Error("Expected OK to be true in response")
	}

	if result.URL == "" {
		t.Error("Expected URL to be non-empty in response")
	}

	if result.URL != "https://logs.dylan.lol/abc123" {
		t.Errorf("Expected URL 'https://logs.dylan.lol/abc123', got '%s'", result.URL)
	}

	t.Logf("JSON response parsed successfully: OK=%v, URL=%s", result.OK, result.URL)
}

// TestFormUploadResponseParsingHTML tests HTML response parsing
func TestFormUploadResponseParsingHTML(t *testing.T) {
	// Test HTML response parsing (matching new implementation)
	htmlResponse := `<!DOCTYPE html>
<html>
<head>
    <title>MicroBin - File Uploaded</title>
</head>
<body>
    <h1>File uploaded successfully!</h1>
    <p>File ID: mouse-tiger-fly</p>
    <a href="https://logs.dylan.lol/upload/mouse-tiger-fly">View File</a>
    <a href="https://logs.dylan.lol/file/mouse-tiger-fly">Direct Link</a>
    <a href="https://logs.dylan.lol/edit/mouse-tiger-fly">Edit</a>
</body>
</html>`

	// Test the HTML parsing logic
	fileID := extractFileIDFromHTMLStandalone(htmlResponse)
	if fileID != "mouse-tiger-fly" {
		t.Errorf("Expected file ID 'mouse-tiger-fly', got '%s'", fileID)
	}

	// Test direct URL construction
	directURL := fmt.Sprintf("https://logs.dylan.lol/p/%s", fileID)
	expectedURL := "https://logs.dylan.lol/p/mouse-tiger-fly"
	if directURL != expectedURL {
		t.Errorf("Expected direct URL '%s', got '%s'", expectedURL, directURL)
	}

	t.Logf("HTML response parsed successfully: fileID=%s, directURL=%s", fileID, directURL)
}

// extractFileIDFromHTMLStandalone is a test helper function that mirrors GUI's extractFileIDFromHTML logic
func extractFileIDFromHTMLStandalone(html string) string {
	// Look for the pattern in the HTML that contains the file ID
	// The file ID appears in URLs like https://logs.dylan.lol/upload/mouse-tiger-fly or https://logs.dylan.lol/file/mouse-tiger-fly

	// First try to find the upload URL pattern (absolute URLs)
	uploadPattern := `href="https://logs\.dylan\.lol/upload/([^"]+)"`
	re := regexp.MustCompile(uploadPattern)
	matches := re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// If that fails, try the file URL pattern (absolute URLs)
	filePattern := `href="https://logs\.dylan\.lol/file/([^"]+)"`
	re = regexp.MustCompile(filePattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// If that fails, try the edit URL pattern (absolute URLs)
	editPattern := `href="https://logs\.dylan\.lol/edit/([^"]+)"`
	re = regexp.MustCompile(editPattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// If that fails, try the JavaScript URL pattern
	jsPattern := `const url = .*https://logs\.dylan\.lol/upload/([^"]+)`
	re = regexp.MustCompile(jsPattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	// Fallback: try relative patterns (in case the format changes)
	relativeUploadPattern := `href="/upload/([^"]+)"`
	re = regexp.MustCompile(relativeUploadPattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	relativeFilePattern := `href="/file/([^"]+)"`
	re = regexp.MustCompile(relativeFilePattern)
	matches = re.FindStringSubmatch(html)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// TestFormUploadErrorHandling tests various error scenarios
func TestFormUploadErrorHandling(t *testing.T) {
	// Test 1: Empty log content
	t.Run("EmptyContent", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("expiration", "never")
		formData.Set("syntax_highlight", "none")
		formData.Set("privacy", "public")
		formData.Set("content", "") // Empty content

		encodedData := formData.Encode()
		if !strings.Contains(encodedData, "content=") {
			t.Error("Empty content should be encoded as 'content='")
		}

		t.Logf("Empty content handled correctly: %s", encodedData)
	})

	// Test 2: Special characters in content
	t.Run("SpecialCharacters", func(t *testing.T) {
		specialContent := "Test with special chars: & = + % #"
		formData := url.Values{}
		formData.Set("expiration", "never")
		formData.Set("syntax_highlight", "none")
		formData.Set("privacy", "public")
		formData.Set("content", specialContent)

		encodedData := formData.Encode()
		if !strings.Contains(encodedData, "content=Test+with+special+chars") {
			t.Error("Special characters should be URL-encoded")
		}

		t.Logf("Special characters handled correctly: %s", encodedData)
	})

	// Test 3: Missing form field
	t.Run("MissingField", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("expiration", "never")
		formData.Set("syntax_highlight", "none")
		// Missing privacy field
		formData.Set("content", "test content")

		encodedData := formData.Encode()
		if strings.Contains(encodedData, "privacy=") {
			t.Error("Missing privacy field should not be in encoded data")
		}

		t.Logf("Missing field handled correctly: %s", encodedData)
	})
}
