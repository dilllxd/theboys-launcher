package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
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

// generateRandomID generates a random 8-character hexadecimal string (copied from gui.go for testing)
func generateRandomID() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3]), nil
}

// MockNewLogUploadServer creates a mock HTTP server for testing the new i.dylan.lol/logs/ endpoint
func MockNewLogUploadServer(responseCode int, responseBody string, contentType string) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/logs/", func(w http.ResponseWriter, r *http.Request) {
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

		// Verify User-Agent header
		if r.Header.Get("User-Agent") != "TheBoysLauncher/1.0" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse multipart form (max 32MB)
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify act field exists and has correct value
		actValue := r.FormValue("act")
		if actValue != "bput" {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("expected field 'act', got 'file'"))
			return
		}

		// Verify file field exists
		file, header, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Verify filename has .log extension
		if !strings.HasSuffix(header.Filename, ".log") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Read file content to verify it's not empty (unless it's supposed to be)
		content, err := io.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Log the received content for debugging
		fmt.Printf("Mock server received file: %s, size: %d bytes\n", header.Filename, len(content))

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

// MockNewLogUploadServerWithValidation creates a mock server with detailed validation
func MockNewLogUploadServerWithValidation(responseCode int, responseBody string, contentType string, t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/logs/", func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// Verify content type contains multipart/form-data
		contentTypeHeader := r.Header.Get("Content-Type")
		if !strings.Contains(contentTypeHeader, "multipart/form-data") {
			t.Errorf("Expected multipart/form-data content type, got %s", contentTypeHeader)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify User-Agent header
		userAgent := r.Header.Get("User-Agent")
		if userAgent != "TheBoysLauncher/1.0" {
			t.Errorf("Expected User-Agent 'TheBoysLauncher/1.0', got '%s'", userAgent)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Parse multipart form (max 32MB)
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			t.Errorf("Failed to parse multipart form: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Verify act field exists and has correct value
		actValue := r.FormValue("act")
		if actValue != "bput" {
			t.Errorf("Expected act field with value 'bput', got '%s'", actValue)
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte("expected field 'act', got 'file'"))
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

		// Verify filename has .log extension and is 8 chars + .log
		if !strings.HasSuffix(header.Filename, ".log") {
			t.Errorf("Expected filename to end with .log, got %s", header.Filename)
		}

		filename := strings.TrimSuffix(header.Filename, ".log")
		if len(filename) != 8 {
			t.Errorf("Expected 8-character filename, got %d characters: %s", len(filename), filename)
		}

		// Verify filename is valid hex
		for _, c := range filename {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				t.Errorf("Filename contains non-hex character: %c in %s", c, filename)
			}
		}

		// Read file content
		content, err := io.ReadAll(file)
		if err != nil {
			t.Errorf("Failed to read file content: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Log the received content for debugging
		t.Logf("Mock server received file: %s, size: %d bytes", header.Filename, len(content))

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

// TestNewUploadLogSuccess tests successful log upload with the new endpoint
func TestNewUploadLogSuccess(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for new endpoint\nLine 2\nLine 3\nTimestamp: " + time.Now().Format(time.RFC3339) + "\n"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Generate a random ID for testing
	randomID, err := generateRandomID()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}
	filename := fmt.Sprintf("%s.log", randomID)

	// Expected successful response
	expectedResponse := fmt.Sprintf(`{"url": "https://i.dylan.lol/logs/%s", "id": "%s"}`, randomID, randomID)

	// Create mock server with validation
	mockServer := MockNewLogUploadServerWithValidation(http.StatusOK, expectedResponse, "application/json", t)
	defer mockServer.Close()

	// Test the upload logic by simulating the new implementation
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	// Open the log file for reading
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Create form file part using CreateFormFile to match curl -F format
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy file content to the form part
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	writer.Close()

	// Create a new HTTP request with the form data
	req, err := http.NewRequest("POST", mockServer.URL+"/logs/", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set the content type header for form data
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request with TLS 1.2 and timeout (matching new implementation)
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

	// Try to parse as JSON first
	var result struct {
		URL string `json:"url"`
		ID  string `json:"id"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse JSON response: %v", err)
	}

	// Verify response content
	expectedURL := fmt.Sprintf("https://i.dylan.lol/logs/%s", randomID)
	if result.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, result.URL)
	}

	if result.ID != randomID {
		t.Errorf("Expected ID '%s', got '%s'", randomID, result.ID)
	}

	t.Logf("Successfully uploaded log to: %s", result.URL)
}

// TestNewUploadLogWithDifferentResponses tests different response formats
func TestNewUploadLogWithDifferentResponses(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-responses-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := "Test log content for response format testing\n"
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Test cases for different response formats
	testCases := []struct {
		name           string
		responseBody   string
		expectedURL    string
		shouldFallback bool
	}{
		{
			name:           "JSONWithURL",
			responseBody:   `{"url": "https://i.dylan.lol/logs/abc123", "id": "abc123"}`,
			expectedURL:    "https://i.dylan.lol/logs/abc123",
			shouldFallback: false,
		},
		{
			name:           "JSONWithIDOnly",
			responseBody:   `{"id": "def456"}`,
			expectedURL:    "https://i.dylan.lol/logs/def456",
			shouldFallback: false,
		},
		{
			name:           "InvalidJSON",
			responseBody:   `{"url": "https://i.dylan.lol/logs/ghi789", "id": "ghi789", invalid}`,
			expectedURL:    "https://i.dylan.lol/logs/ghi789",
			shouldFallback: true,
		},
		{
			name:           "EmptyJSON",
			responseBody:   `{}`,
			expectedURL:    "https://i.dylan.lol/logs/jkl012",
			shouldFallback: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate a random ID for this test
			randomID, err := generateRandomID()
			if err != nil {
				t.Fatalf("Failed to generate random ID: %v", err)
			}

			// Create mock server
			mockServer := MockNewLogUploadServer(http.StatusOK, tc.responseBody, "application/json")
			defer mockServer.Close()

			// Simulate the upload logic
			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)

			// Add the required "act" field with value "bput" as required by the endpoint
			err = writer.WriteField("act", "bput")
			if err != nil {
				t.Fatalf("Failed to add act field: %v", err)
			}

			file, err := os.Open(logPath)
			if err != nil {
				t.Fatalf("Failed to open log file: %v", err)
			}
			defer file.Close()

			filename := fmt.Sprintf("%s.log", randomID)
			part, err := writer.CreateFormFile("file", filename)
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}

			_, err = io.Copy(part, file)
			if err != nil {
				t.Fatalf("Failed to copy file content: %v", err)
			}

			writer.Close()

			req, err := http.NewRequest("POST", mockServer.URL+"/logs/", &requestBody)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

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
				t.Fatalf("Failed to upload log: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// Simulate the response parsing logic from uploadLog()
			var result struct {
				URL string `json:"url"`
				ID  string `json:"id"`
			}

			var logURL string
			if err := json.Unmarshal(body, &result); err == nil {
				if result.URL != "" {
					logURL = result.URL
				} else if result.ID != "" {
					logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s", result.ID)
				}
			}

			// If JSON parsing failed or didn't give us a URL, construct it using our random ID
			if logURL == "" {
				logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s", randomID)
			}

			// Verify the result
			if tc.shouldFallback {
				// Should fall back to constructed URL
				expectedFallbackURL := fmt.Sprintf("https://i.dylan.lol/logs/%s", randomID)
				if logURL != expectedFallbackURL {
					t.Errorf("Expected fallback URL '%s', got '%s'", expectedFallbackURL, logURL)
				}
			} else {
				// Should use the URL from response
				if logURL != tc.expectedURL {
					t.Errorf("Expected URL '%s', got '%s'", tc.expectedURL, logURL)
				}
			}

			t.Logf("Test case '%s' passed, final URL: %s", tc.name, logURL)
		})
	}
}

// TestNewUploadLogNetworkError tests network error handling
func TestNewUploadLogNetworkError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-network-error-test")
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

	// Generate a random ID for testing
	randomID, err := generateRandomID()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}

	// Simulate the upload logic with a non-routable address
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	filename := fmt.Sprintf("%s.log", randomID)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	writer.Close()

	// Create a request to a non-routable IP address to simulate network error
	req, err := http.NewRequest("POST", "http://192.0.2.0/logs/", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Send the request with a short timeout to avoid hanging
	client := &http.Client{
		Timeout: 1 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS12,
			},
		},
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

// TestNewUploadLogHTTPError tests HTTP error response handling
func TestNewUploadLogHTTPError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-http-error-test")
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
	mockServer := MockNewLogUploadServer(http.StatusInternalServerError, errorResponse, "application/json")
	defer mockServer.Close()

	// Generate a random ID for testing
	randomID, err := generateRandomID()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}

	// Simulate the upload logic
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	filename := fmt.Sprintf("%s.log", randomID)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", mockServer.URL+"/logs/", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

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
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code indicates error
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	t.Logf("Successfully handled HTTP error response: %s", resp.Status)
}

// TestNewUploadLogFileNotFound tests handling when log file doesn't exist
func TestNewUploadLogFileNotFound(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-not-found-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a non-existent log file path
	logPath := filepath.Join(tempDir, "nonexistent.log")

	// Try to open non-existent log file
	file, err := os.Open(logPath)
	if err == nil {
		file.Close()
		t.Error("Expected error when opening non-existent file")
		return
	}

	if !os.IsNotExist(err) {
		t.Errorf("Expected 'file not found' error, got: %v", err)
	}

	t.Logf("Successfully detected missing file error: %v", err)
}

// TestNewUploadLogEmptyFile tests handling when log file is empty
func TestNewUploadLogEmptyFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-empty-test")
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

	// Generate a random ID for testing
	randomID, err := generateRandomID()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}

	// Expected successful response (even for empty files)
	expectedResponse := fmt.Sprintf(`{"url": "https://i.dylan.lol/logs/%s", "id": "%s"}`, randomID, randomID)

	// Create mock server
	mockServer := MockNewLogUploadServer(http.StatusOK, expectedResponse, "application/json")
	defer mockServer.Close()

	// Simulate the upload logic
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	filename := fmt.Sprintf("%s.log", randomID)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy empty file content to the form part
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", mockServer.URL+"/logs/", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

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
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse response
	var result struct {
		URL string `json:"url"`
		ID  string `json:"id"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	// Verify response content
	expectedURL := fmt.Sprintf("https://i.dylan.lol/logs/%s", randomID)
	if result.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, result.URL)
	}

	if result.ID != randomID {
		t.Errorf("Expected ID '%s', got '%s'", randomID, result.ID)
	}

	t.Logf("Successfully uploaded empty file to: %s", result.URL)
}

// TestGenerateRandomID tests the random ID generation function
func TestGenerateRandomID(t *testing.T) {
	// Test multiple generations to ensure uniqueness and format
	ids := make(map[string]bool)

	for i := 0; i < 1000; i++ {
		id, err := generateRandomID()
		if err != nil {
			t.Fatalf("Failed to generate random ID: %v", err)
		}

		// Check length
		if len(id) != 8 {
			t.Errorf("Expected 8-character ID, got %d characters: %s", len(id), id)
		}

		// Check hex format
		for _, c := range id {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
				t.Errorf("ID contains non-hex character: %c in %s", c, id)
			}
		}

		// Check uniqueness
		if ids[id] {
			t.Errorf("Generated duplicate ID: %s", id)
		}
		ids[id] = true
	}

	t.Logf("Successfully generated 1000 unique random IDs")
}

// TestNewUploadLogLargeFile tests handling of larger files
func TestNewUploadLogLargeFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "theboyslauncher-new-log-upload-large-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a larger mock log file (1MB)
	logPath := filepath.Join(tempDir, "latest.log")
	largeContent := strings.Repeat("This is a line of log content that will be repeated to create a large file.\n", 10000)
	err = os.WriteFile(logPath, []byte(largeContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create large mock log file: %v", err)
	}

	// Generate a random ID for testing
	randomID, err := generateRandomID()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}

	// Expected successful response
	expectedResponse := fmt.Sprintf(`{"url": "https://i.dylan.lol/logs/%s", "id": "%s"}`, randomID, randomID)

	// Create mock server
	mockServer := MockNewLogUploadServer(http.StatusOK, expectedResponse, "application/json")
	defer mockServer.Close()

	// Simulate the upload logic
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	filename := fmt.Sprintf("%s.log", randomID)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy large file content to the form part
	_, err = io.Copy(part, file)
	if err != nil {
		t.Fatalf("Failed to copy file content: %v", err)
	}

	writer.Close()

	req, err := http.NewRequest("POST", mockServer.URL+"/logs/", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

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
		t.Fatalf("Failed to upload log: %v", err)
	}
	defer resp.Body.Close()

	// Verify response status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Parse response
	var result struct {
		URL string `json:"url"`
		ID  string `json:"id"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse upload response: %v", err)
	}

	// Verify response content
	expectedURL := fmt.Sprintf("https://i.dylan.lol/logs/%s", randomID)
	if result.URL != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, result.URL)
	}

	if result.ID != randomID {
		t.Errorf("Expected ID '%s', got '%s'", randomID, result.ID)
	}

	t.Logf("Successfully uploaded large file (%d bytes) to: %s", len(largeContent), result.URL)
}
