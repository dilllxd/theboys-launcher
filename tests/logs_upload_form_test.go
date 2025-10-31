package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// MockFormUploadServer creates a mock HTTP server for testing the new form-based log uploads
func MockFormUploadServer(responseCode int, responseBody string, contentType string) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify content type is application/x-www-form-urlencoded
		if !strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse form data
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify required form fields exist
		if r.FormValue("expiration") == "" || r.FormValue("syntax_highlight") == "" ||
			r.FormValue("privacy") == "" || r.FormValue("content") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Set response headers and body
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		} else {
			w.Header().Set("Content-Type", "application/json")
		}

		w.WriteHeader(responseCode)
		w.Write([]byte(responseBody))
	})

	return httptest.NewServer(mux)
}

// MockFormUploadServerWithValidation creates a mock HTTP server with detailed validation for testing form uploads
func MockFormUploadServerWithValidation(responseCode int, responseBody string, contentType string, t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify content type is application/x-www-form-urlencoded
		if !strings.Contains(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			t.Errorf("Expected application/x-www-form-urlencoded content type, got %s", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse form data
		err := r.ParseForm()
		if err != nil {
			t.Errorf("Failed to parse form data: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify required form fields exist and have correct values
		expiration := r.FormValue("expiration")
		syntaxHighlight := r.FormValue("syntax_highlight")
		privacy := r.FormValue("privacy")
		content := r.FormValue("content")

		if expiration != "never" {
			t.Errorf("Expected expiration='never', got '%s'", expiration)
		}

		if syntaxHighlight != "none" {
			t.Errorf("Expected syntax_highlight='none', got '%s'", syntaxHighlight)
		}

		if privacy != "public" {
			t.Errorf("Expected privacy='public', got '%s'", privacy)
		}

		if content == "" {
			t.Error("Content field should not be empty")
		}

		// Set response headers and body
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		} else {
			w.Header().Set("Content-Type", "application/json")
		}

		w.WriteHeader(responseCode)
		w.Write([]byte(responseBody))
	})

	return httptest.NewServer(mux)
}

// TestFormUploadLogSuccess tests successful log upload with the new form format
func TestFormUploadLogSuccess(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-form-upload-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content\nLine 2\nLine 3\n"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Expected successful response
	expectedResponse := `{"ok": true, "url": "https://logs.dylan.lol/abc123"}`

	// Create mock server with validation
	mockServer := MockFormUploadServerWithValidation(http.StatusOK, expectedResponse, "application/json", t)
	defer mockServer.Close()

	// Test the new upload format by simulating the uploadLog logic
	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create form data with the required fields (matching the new implementation)
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Create a new HTTP request with the form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: &http.Transport{
			// Match the TLS configuration from the actual implementation
			// Note: httptest.Server doesn't use TLS, so this is just for completeness
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse response
	var result struct {
		OK  bool   `json:"ok"`
		URL string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	// Verify response content
	if !result.OK {
		t.Error("Expected OK to be true in response")
	}

	if result.URL == "" {
		t.Error("Expected URL to be non-empty in response")
	}

	if result.URL != "https://logs.dylan.lol/abc123" {
		t.Errorf("Expected URL 'https://logs.dylan.lol/abc123', got '%s'", result.URL)
	}
}

// TestFormUploadLogNetworkError tests network error handling with the new form format
func TestFormUploadLogNetworkError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-form-upload-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for network error test"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create form data with the required fields
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Create a request to a non-routable IP address to simulate network error
	req, err := http.NewRequest("POST", "http://192.0.2.0/upload", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request with a short timeout to avoid hanging
	client := &http.Client{
		Timeout: 1 * time.Second,
	}

	resp, err := client.Do(req)
	if err == nil {
		if resp != nil {
			resp.Body.Close()
		}
		t.Error("Expected network error, but request succeeded")
		return
	}

	// We got a timeout error, which is a valid network error
	if strings.Contains(strings.ToLower(err.Error()), "timeout") {
		t.Logf("Successfully detected network timeout error: %v", err)
		return // Test passes, we got the expected network error
	}

	// Verify it's a network error
	if !strings.Contains(strings.ToLower(err.Error()), "connection refused") && !strings.Contains(strings.ToLower(err.Error()), "no route to host") {
		t.Errorf("Expected network error, got: %v", err)
	}
}

// TestFormUploadLogInvalidJSON tests invalid JSON response handling with the new form format
func TestFormUploadLogInvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-form-upload-invalid-json-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for invalid JSON test"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Invalid JSON response
	invalidJSONResponse := `{"ok": true, "url": "https://logs.dylan.lol/abc123", invalid}`

	// Create mock server
	mockServer := MockFormUploadServer(http.StatusOK, invalidJSONResponse, "application/json")
	defer mockServer.Close()

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create form data with the required fields
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Create a new HTTP request with the form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Try to parse response (should fail)
	var result struct {
		OK  bool   `json:"ok"`
		URL string `json:"url"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err == nil {
		t.Error("Expected JSON parsing error, but parsing succeeded")
	}
}

// TestFormUploadLogHTMLResponse tests HTML response handling with the new form format
func TestFormUploadLogHTMLResponse(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-form-upload-html-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for HTML response test"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// HTML response that simulates MicroBin's response format
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

	// Create mock server
	mockServer := MockFormUploadServer(http.StatusOK, htmlResponse, "text/html")
	defer mockServer.Close()

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create form data with the required fields
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Create a new HTTP request with the form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Verify content type indicates HTML
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(strings.ToLower(contentType), "text/html") {
		t.Errorf("Expected text/html content type, got %s", contentType)
	}

	// Verify response starts with HTML
	if len(body) == 0 || body[0] != '<' {
		t.Error("Expected HTML response (starting with '<')")
	}

	// Test the HTML parsing logic
	htmlContent := string(body)
	fileID := extractFileIDFromHTMLForTesting(htmlContent)
	if fileID != "mouse-tiger-fly" {
		t.Errorf("Expected file ID 'mouse-tiger-fly', got '%s'", fileID)
	}
}

// TestFormUploadLogEmptyFile tests handling when log file is empty with the new form format
func TestFormUploadLogEmptyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-form-upload-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create an empty mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	err = os.WriteFile(logPath, []byte{}, 0644)
	if err != nil {
		t.Fatalf("Failed to create empty mock log file: %v", err)
	}

	// Expected successful response (even for empty files)
	expectedResponse := `{"ok": true, "url": "https://logs.dylan.lol/empty123"}`

	// Create mock server
	mockServer := MockFormUploadServer(http.StatusOK, expectedResponse, "application/json")
	defer mockServer.Close()

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Verify content is empty
	if len(content) != 0 {
		t.Errorf("Expected empty file content, got %d bytes", len(content))
	}

	// Create form data with the required fields (content will be empty)
	formData := url.Values{}
	formData.Set("expiration", "never")
	formData.Set("syntax_highlight", "none")
	formData.Set("privacy", "public")
	formData.Set("content", string(content))

	// Create a new HTTP request with the form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", strings.NewReader(formData.Encode()))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Parse response
	var result struct {
		OK  bool   `json:"ok"`
		URL string `json:"url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	// Verify response content
	if !result.OK {
		t.Error("Expected OK to be true in response for empty file")
	}

	if result.URL == "" {
		t.Error("Expected URL to be non-empty in response for empty file")
	}
}

// extractFileIDFromHTMLForTesting is a test helper function that mirrors the GUI's extractFileIDFromHTML logic
func extractFileIDFromHTMLForTesting(html string) string {
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
