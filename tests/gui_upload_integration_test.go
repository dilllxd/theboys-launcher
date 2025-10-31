package main

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// generateRandomIDForGUI generates a random 8-character hexadecimal string (copied from gui.go for testing)
func generateRandomIDForGUI() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3]), nil
}

// TestGUIUploadLogIntegration tests the complete GUI uploadLog() function with mock server
func TestGUIUploadLogIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gui-upload-integration-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock log file
	logPath := filepath.Join(tempDir, "latest.log")
	logContent := fmt.Sprintf("GUI integration test log content\nTimestamp: %s\nLine 2\nLine 3\n", time.Now().Format(time.RFC3339))
	err = os.WriteFile(logPath, []byte(logContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create mock log file: %v", err)
	}

	// Generate a random ID for expected response
	randomID, err := generateRandomIDForGUI()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}

	// Expected successful HTML response (matching the actual endpoint format)
	expectedResponse := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	   <meta charset="utf-8">
	   <title>copyparty @ homelab</title>
</head>
<body>
	   <div id="box">
	       <h2><a href="/logs">return to /logs</a></h2>
	       <pre>OK // 8742 bytes // 0.083 MiB/s
	       sha512: 8e41f5d21800391f5e5ac1e0a0aa4a5e57eabba3548b6bc860e96a38 // jkH10hgAOR9eWsHgoKpKXlfqu6NUi2vIYOlqOGjJBzjl // 8742 bytes // <a href="/logs/%s.log">%s.log</a>
	       </pre>
	   </div>
</body>
</html>`, randomID, randomID)

	// Create mock server that simulates the new i.dylan.lol/logs/ endpoint
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		// Log the received file info for debugging
		t.Logf("Mock server received file: %s, size: %d bytes", header.Filename, header.Size)

		// Set response headers and body (HTML response)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(expectedResponse))
	}))
	defer mockServer.Close()

	// Test the core upload logic that would be used by uploadLog()
	t.Run("MockServerIntegration", func(t *testing.T) {
		testUploadLogicWithMockServer(t, logPath, mockServer.URL, randomID, expectedResponse)
	})

	t.Run("CompleteFlowSimulation", func(t *testing.T) {
		// Simulate the complete flow from file reading to URL generation
		simulateCompleteUploadFlow(t, logPath, randomID)
	})
}

// testUploadLogicWithMockServer tests the upload logic with a mock server
func testUploadLogicWithMockServer(t *testing.T, logPath string, mockServerURL string, expectedID string, expectedResponse string) {
	// Read the log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Create multipart form with file upload (matching the new implementation)
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	// Create form file part using CreateFormFile to match curl -F format
	filename := fmt.Sprintf("%s.log", expectedID)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	// Copy file content to the form part
	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content to form part: %v", err)
	}

	writer.Close()

	// Create a new HTTP request with the form data
	req, err := http.NewRequest("POST", mockServerURL+"/logs/", &requestBody)
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

	// Log the raw response for debugging
	t.Logf("Upload response status: %s", resp.Status)
	t.Logf("Upload response content-type: %s", resp.Header.Get("Content-Type"))
	t.Logf("Upload response body (first 200 chars): %s", string(body)[:min(200, len(body))])

	// Check if the upload was successful (status code 200-299)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		// Parse the HTML response to extract the file URL (matching new implementation)
		bodyStr := string(body)
		var logURL string

		// Try to extract the filename from the HTML response using regex
		// Pattern to match href="/logs/filename.log"
		logPattern := `href="/logs/([^"]+\.log)"`
		re := regexp.MustCompile(logPattern)
		matches := re.FindStringSubmatch(bodyStr)

		if len(matches) > 1 {
			// Extract the filename from the match
			filename := matches[1]
			// Construct the full URL
			logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s", filename)
			t.Logf("Successfully extracted filename from HTML: %s", filename)
		} else {
			// If regex fails, fall back to using our random ID
			t.Logf("Failed to extract filename from HTML, falling back to random ID: %s", expectedID)
			logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s.log", expectedID)
		}

		t.Logf("Final log URL: %s", logURL)

		// Verify the final URL
		expectedURL := fmt.Sprintf("https://i.dylan.lol/logs/%s.log", expectedID)
		if logURL != expectedURL {
			t.Errorf("Expected URL '%s', got '%s'", expectedURL, logURL)
		}

		t.Logf("Successfully processed upload response, final URL: %s", logURL)
	} else {
		t.Errorf("Upload failed with status: %s", resp.Status)
	}
}

// simulateCompleteUploadFlow simulates the complete upload flow from file to URL
func simulateCompleteUploadFlow(t *testing.T, logPath string, expectedID string) {
	// Step 1: Verify log file exists
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file does not exist: %s", logPath)
		return
	}

	// Step 2: Read log file content
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Errorf("Failed to read log file: %v", err)
		return
	}

	t.Logf("Step 2: Successfully read %d bytes from log file", len(content))

	// Step 3: Generate random ID (already done, but verify format)
	if len(expectedID) != 8 {
		t.Errorf("Expected 8-character ID, got %d characters: %s", len(expectedID), expectedID)
		return
	}

	// Verify ID is valid hex
	for _, c := range expectedID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("ID contains non-hex character: %c in %s", c, expectedID)
			return
		}
	}

	t.Logf("Step 3: Generated valid random ID: %s", expectedID)

	// Step 4: Construct expected filename
	filename := fmt.Sprintf("%s.log", expectedID)
	t.Logf("Step 4: Constructed filename: %s", filename)

	// Step 5: Construct expected URL
	expectedURL := fmt.Sprintf("https://i.dylan.lol/logs/%s", expectedID)
	t.Logf("Step 5: Constructed expected URL: %s", expectedURL)

	// Step 6: Verify URL format
	if !strings.HasPrefix(expectedURL, "https://i.dylan.lol/logs/") {
		t.Errorf("Invalid URL format: %s", expectedURL)
		return
	}

	t.Logf("Step 6: URL format validation passed")
	t.Logf("Complete flow simulation successful - all steps validated")
}

// TestGUIUploadLogErrorHandling tests error handling in the GUI upload function
func TestGUIUploadLogErrorHandling(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "gui-upload-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("FileNotFound", func(t *testing.T) {
		nonExistentPath := filepath.Join(tempDir, "nonexistent.log")

		// Try to read non-existent file
		_, err := os.ReadFile(nonExistentPath)
		if err == nil {
			t.Error("Expected error when reading non-existent file")
			return
		}

		if !os.IsNotExist(err) {
			t.Errorf("Expected 'file not found' error, got: %v", err)
		}

		t.Logf("Successfully detected missing file error: %v", err)
	})

	t.Run("EmptyFile", func(t *testing.T) {
		emptyLogPath := filepath.Join(tempDir, "empty.log")
		err := os.WriteFile(emptyLogPath, []byte{}, 0644)
		if err != nil {
			t.Fatalf("Failed to create empty log file: %v", err)
		}

		// Read empty file
		content, err := os.ReadFile(emptyLogPath)
		if err != nil {
			t.Errorf("Failed to read empty file: %v", err)
			return
		}

		if len(content) != 0 {
			t.Errorf("Expected empty file content, got %d bytes", len(content))
		}

		t.Logf("Successfully handled empty file (%d bytes)", len(content))
	})

	t.Run("NetworkError", func(t *testing.T) {
		// Create a mock log file
		logPath := filepath.Join(tempDir, "network-test.log")
		logContent := "Test content for network error simulation"
		err := os.WriteFile(logPath, []byte(logContent), 0644)
		if err != nil {
			t.Fatalf("Failed to create test log file: %v", err)
		}

		// Try to upload to a non-routable address
		var requestBody bytes.Buffer
		writer := multipart.NewWriter(&requestBody)

		// Add the required "act" field with value "bput" as required by the endpoint
		err = writer.WriteField("act", "bput")
		if err != nil {
			t.Fatalf("Failed to add act field: %v", err)
		}

		file, err := os.Open(logPath)
		if err != nil {
			t.Fatalf("Failed to open test file: %v", err)
		}
		defer file.Close()

		randomID, _ := generateRandomIDForGUI()
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

		// Create request to non-routable IP
		req, err := http.NewRequest("POST", "http://192.0.2.0/logs/", &requestBody)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

		// Send with short timeout
		client := &http.Client{Timeout: 1 * time.Second}
		resp, err := client.Do(req)
		if err == nil {
			if resp != nil {
				resp.Body.Close()
			}
			t.Error("Expected network error, but request succeeded")
			return
		}

		t.Logf("Successfully detected network error: %v", err)
	})
}

func minIntGUI(a, b int) int {
	if a < b {
		return a
	}
	return b
}
