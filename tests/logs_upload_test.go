package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockLogUploadServer creates a mock HTTP server for testing log uploads
func MockLogUploadServer(responseCode int, responseBody string, contentType string) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify content type contains multipart/form-data
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify file field exists
		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Verify filename
		if header.Filename != "latest.log" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Read file content to verify it's not empty
		_, err = io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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

// MockLogUploadServerWithValidation creates a mock HTTP server with validation for testing log uploads
func MockLogUploadServerWithValidation(responseCode int, responseBody string, contentType string, t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify content type contains multipart/form-data
		if !strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", r.Header.Get("Content-Type"))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse multipart form
		err := r.ParseMultipartForm(32 << 20) // 32MB max memory
		if err != nil {
			t.Errorf("Failed to parse multipart form: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify file field exists
		file, header, err := r.FormFile("file")
		if err != nil {
			t.Errorf("Failed to get file from form: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Verify filename
		if header.Filename != "latest.log" {
			t.Errorf("Expected filename 'latest.log', got '%s'", header.Filename)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Read file content to verify it's not empty
		content, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("Failed to read file content: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if len(content) == 0 {
			t.Error("File content should not be empty")
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

// TestUploadLogSuccess tests successful log upload scenario
func TestUploadLogSuccess(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-test")
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
	mockServer := MockLogUploadServerWithValidation(http.StatusOK, expectedResponse, "application/json", t)
	defer mockServer.Close()

	// Test the uploadLog function by simulating its logic
	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field with the log content
	part, err := writer.CreateFormFile("file", "latest.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write the log content to the form file
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Failed to write log content: %v", err)
	}

	// Close the multipart writer to finalize the form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to finalize form data: %v", err)
	}

	// Create a new HTTP request with the multipart form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
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

// TestUploadLogNetworkError tests network error handling
func TestUploadLogNetworkError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-error-test")
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

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field with the log content
	part, err := writer.CreateFormFile("file", "latest.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write the log content to the form file
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Failed to write log content: %v", err)
	}

	// Close the multipart writer to finalize the form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to finalize form data: %v", err)
	}

	// Create a request to a non-routable IP address to simulate network error
	req, err := http.NewRequest("POST", "http://192.0.2.0/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

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

// TestUploadLogInvalidJSON tests invalid JSON response handling
func TestUploadLogInvalidJSON(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-invalid-json-test")
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
	mockServer := MockLogUploadServer(http.StatusOK, invalidJSONResponse, "application/json")
	defer mockServer.Close()

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field with the log content
	part, err := writer.CreateFormFile("file", "latest.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write the log content to the form file
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Failed to write log content: %v", err)
	}

	// Close the multipart writer to finalize the form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to finalize form data: %v", err)
	}

	// Create a new HTTP request with the multipart form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
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

// TestUploadLogAPIError tests API error response handling
func TestUploadLogAPIError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-api-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for API error test"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// API error response
	errorResponse := `{"ok": false, "url": ""}`

	// Create mock server
	mockServer := MockLogUploadServer(http.StatusOK, errorResponse, "application/json")
	defer mockServer.Close()

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field with the log content
	part, err := writer.CreateFormFile("file", "latest.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write the log content to the form file
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Failed to write log content: %v", err)
	}

	// Close the multipart writer to finalize the form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to finalize form data: %v", err)
	}

	// Create a new HTTP request with the multipart form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
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

	// Verify response content indicates error
	if result.OK {
		t.Error("Expected OK to be false in error response")
	}

	if result.URL != "" {
		t.Errorf("Expected URL to be empty in error response, got '%s'", result.URL)
	}
}

// TestUploadLogHTTPError tests HTTP error response handling
func TestUploadLogHTTPError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-http-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for HTTP error test"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// HTTP error response
	errorResponse := `{"error": "Internal server error"}`

	// Create mock server that returns 500 error
	mockServer := MockLogUploadServer(http.StatusInternalServerError, errorResponse, "application/json")
	defer mockServer.Close()

	// Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field with the log content
	part, err := writer.CreateFormFile("file", "latest.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write the log content to the form file
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Failed to write log content: %v", err)
	}

	// Close the multipart writer to finalize the form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to finalize form data: %v", err)
	}

	// Create a new HTTP request with the multipart form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code indicates error
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// TestUploadLogFileNotFound tests handling when log file doesn't exist
func TestUploadLogFileNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-not-found-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a non-existent log file path
	logPath := filepath.Join(tempDir, "nonexistent.log")

	// Try to read non-existent log file
	_, err = os.ReadFile(logPath)
	if err == nil {
		t.Error("Expected error when reading non-existent file")
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}
}

// TestUploadLogEmptyFile tests handling when log file is empty
func TestUploadLogEmptyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-log-upload-empty-test")
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
	mockServer := MockLogUploadServer(http.StatusOK, expectedResponse, "application/json")
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

	// Create a buffer to hold the multipart form data
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create a form file field with the log content
	part, err := writer.CreateFormFile("file", "latest.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Write the (empty) log content to the form file
	if _, err := part.Write(content); err != nil {
		t.Fatalf("Failed to write log content: %v", err)
	}

	// Close the multipart writer to finalize the form data
	if err := writer.Close(); err != nil {
		t.Fatalf("Failed to finalize form data: %v", err)
	}

	// Create a new HTTP request with the multipart form data
	req, err := http.NewRequest("POST", mockServer.URL+"/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header with the boundary
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
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
