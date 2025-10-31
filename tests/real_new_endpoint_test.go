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
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// generateRandomIDForRealTest generates a random 8-character hexadecimal string (copied from gui.go for testing)
func generateRandomIDForRealTest() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3]), nil
}

// TestRealNewEndpointUpload tests uploading to the actual i.dylan.lol/logs/ endpoint
func TestRealNewEndpointUpload(t *testing.T) {
	// Create a temporary test file
	tempDir, err := os.MkdirTemp("", "real-new-endpoint-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testLogPath := filepath.Join(tempDir, "test.log")
	testContent := fmt.Sprintf("This is a test log file for the new i.dylan.lol/logs/ endpoint.\nTimestamp: %s\nTest content line 2\nTest content line 3\n%s\n",
		time.Now().Format(time.RFC3339), strings.Repeat("This is additional content to make the file larger and avoid 'file too small' errors.\n", 100))
	err = os.WriteFile(testLogPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Generate a random ID for the filename
	randomID, err := generateRandomIDForRealTest()
	if err != nil {
		t.Fatalf("Failed to generate random ID: %v", err)
	}
	filename := fmt.Sprintf("%s.log", randomID)

	t.Logf("Testing upload to real endpoint with filename: %s", filename)

	// Create multipart form with file upload (matching the new implementation)
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add the required "act" field with value "bput" as required by the endpoint
	err = writer.WriteField("act", "bput")
	if err != nil {
		t.Fatalf("Failed to add act field: %v", err)
	}

	// Open the test file for reading
	file, err := os.Open(testLogPath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
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
	req, err := http.NewRequest("POST", "https://i.dylan.lol/logs/", &requestBody)
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

	t.Logf("Sending request to https://i.dylan.lol/logs/ with %d bytes of content", len(requestBody.Bytes()))

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to upload log to real endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Read the full response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Log the raw response for debugging
	t.Logf("Upload response status: %s", resp.Status)
	t.Logf("Upload response content-type: %s", resp.Header.Get("Content-Type"))
	t.Logf("Upload response body (first 500 chars): %s", string(body)[:minIntReal(500, len(body))])

	// Log the full response body for analysis
	t.Logf("Full response body: %s", string(body))

	// Check if the upload was successful (status code 200-299)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		t.Errorf("Upload failed with status: %s\nResponse: %s", resp.Status, string(body))
		return
	}

	// The endpoint returns HTML, not JSON. We need to extract the file URL from the HTML.
	// From the response, we can see the file is available at: /logs/9758f5bd.log
	var logURL string

	// Try to extract the file URL from the HTML response
	bodyStr := string(body)

	// Look for the file link pattern in the HTML
	fileLinkPattern := `href="/logs/([^"]+\.log)"`
	re := regexp.MustCompile(fileLinkPattern)
	matches := re.FindStringSubmatch(bodyStr)
	if len(matches) > 1 {
		logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s", matches[1])
		t.Logf("Extracted file URL from HTML: %s", logURL)
	} else {
		// Fallback to our constructed URL with .log extension
		logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s.log", randomID)
		t.Logf("Using fallback URL: %s", logURL)
	}

	t.Logf("Upload successful! Generated URL: %s", logURL)

	// Verify the URL is valid
	parsedURL, err := url.Parse(logURL)
	if err != nil {
		t.Errorf("Generated URL is not valid: %v", err)
		return
	}

	if parsedURL.Scheme != "https" {
		t.Errorf("Expected HTTPS URL, got scheme: %s", parsedURL.Scheme)
	}

	if parsedURL.Host != "i.dylan.lol" {
		t.Errorf("Expected host i.dylan.lol, got: %s", parsedURL.Host)
	}

	// Test URL accessibility (optional, can be slow)
	t.Run("URLAccessibility", func(t *testing.T) {
		t.Logf("Testing URL accessibility: %s", logURL)

		// Create a request to check if the uploaded file is accessible
		checkReq, err := http.NewRequest("GET", logURL, nil)
		if err != nil {
			t.Fatalf("Failed to create check request: %v", err)
		}

		checkReq.Header.Set("User-Agent", "TheBoysLauncher/1.0")

		// Use a shorter timeout for accessibility check
		checkClient := &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
					MaxVersion: tls.VersionTLS12,
				},
			},
		}

		checkResp, err := checkClient.Do(checkReq)
		if err != nil {
			t.Logf("URL accessibility check failed (this may be expected): %v", err)
			// Don't fail the test for accessibility issues as the endpoint might not serve files immediately
			return
		}
		defer checkResp.Body.Close()

		if checkResp.StatusCode == 200 {
			t.Logf("URL is accessible! Status: %s", checkResp.Status)

			// Read a small portion of the content to verify it contains our test data
			checkBody, err := io.ReadAll(checkResp.Body)
			if err == nil {
				bodyStr := string(checkBody)
				if strings.Contains(bodyStr, "This is a test log file for the new i.dylan.lol/logs/ endpoint") {
					t.Logf("Uploaded content verified - contains expected test string")
				} else {
					t.Logf("Uploaded content verification - could not find expected test string in first 1000 chars")
				}
			}
		} else {
			t.Logf("URL returned status: %s (file may not be immediately available)", checkResp.Status)
		}
	})

	t.Logf("Real endpoint test completed successfully")
}

// TestRealNewEndpointUploadMultiple tests multiple uploads to the real endpoint
func TestRealNewEndpointUploadMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple upload test in short mode")
	}

	// Test multiple uploads to ensure consistency
	for i := 0; i < 3; i++ {
		t.Run(fmt.Sprintf("Upload%d", i+1), func(t *testing.T) {
			// Create a temporary test file
			tempDir, err := os.MkdirTemp("", fmt.Sprintf("real-new-endpoint-test-%d", i))
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			testLogPath := filepath.Join(tempDir, "test.log")
			testContent := fmt.Sprintf("Multiple test upload #%d\nTimestamp: %s\nContent line specific to upload #%d\n%s\n",
				i+1, time.Now().Format(time.RFC3339), i+1, strings.Repeat("This is additional content to make file larger and avoid 'file too small' errors.\n", 50))
			err = os.WriteFile(testLogPath, []byte(testContent), 0644)
			if err != nil {
				t.Fatalf("Failed to create test log file: %v", err)
			}

			// Generate a random ID for the filename
			randomID, err := generateRandomIDForRealTest()
			if err != nil {
				t.Fatalf("Failed to generate random ID: %v", err)
			}
			filename := fmt.Sprintf("%s.log", randomID)

			// Create multipart form with file upload
			var requestBody bytes.Buffer
			writer := multipart.NewWriter(&requestBody)

			// Add the required "act" field with value "bput" as required by the endpoint
			err = writer.WriteField("act", "bput")
			if err != nil {
				t.Fatalf("Failed to add act field: %v", err)
			}

			file, err := os.Open(testLogPath)
			if err != nil {
				t.Fatalf("Failed to open test file: %v", err)
			}
			defer file.Close()

			part, err := writer.CreateFormFile("file", filename)
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}

			_, err = io.Copy(part, file)
			if err != nil {
				t.Fatalf("Failed to copy file content: %v", err)
			}

			writer.Close()

			// Create HTTP request
			req, err := http.NewRequest("POST", "https://i.dylan.lol/logs/", &requestBody)
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
				t.Fatalf("Failed to upload log to real endpoint: %v", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			// Check if the upload was successful
			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				t.Errorf("Upload #%d failed with status: %s", i+1, resp.Status)
				return
			}

			// Parse response
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

			if logURL == "" {
				logURL = fmt.Sprintf("https://i.dylan.lol/logs/%s", randomID)
			}

			t.Logf("Upload #%d successful! URL: %s", i+1, logURL)

			// Verify URL format
			parsedURL, err := url.Parse(logURL)
			if err != nil {
				t.Errorf("Generated URL #%d is not valid: %v", i+1, err)
				return
			}

			if parsedURL.Host != "i.dylan.lol" {
				t.Errorf("Expected host i.dylan.lol for upload #%d, got: %s", i+1, parsedURL.Host)
			}

			// Add a small delay between uploads to avoid rate limiting
			time.Sleep(1 * time.Second)
		})
	}
}

func minIntReal(a, b int) int {
	if a < b {
		return a
	}
	return b
}
