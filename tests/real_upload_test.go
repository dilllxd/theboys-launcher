package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRealUpload tests uploading to the actual logs.dylan.lol endpoint
func TestRealUpload(t *testing.T) {
	// Create a temporary test file
	tempDir, err := os.MkdirTemp("", "real-upload-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testLogPath := filepath.Join(tempDir, "test.log")
	testContent := "This is a test log file for upload testing.\nTimestamp: " + time.Now().Format(time.RFC3339) + "\n"
	err = os.WriteFile(testLogPath, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Read the test file
	content, err := os.ReadFile(testLogPath)
	if err != nil {
		t.Fatalf("Failed to read test file: %v", err)
	}

	// Test with different approaches to see what works

	// Approach 1: Basic multipart upload (current implementation)
	t.Run("BasicMultipart", func(t *testing.T) {
		testBasicMultipartUpload(t, content)
	})

	// Approach 2: With TLS 1.2 explicitly set
	t.Run("TLS12Multipart", func(t *testing.T) {
		testTLS12MultipartUpload(t, content)
	})

	// Approach 3: With explicit content type for file part
	t.Run("ExplicitContentType", func(t *testing.T) {
		testExplicitContentTypeUpload(t, content)
	})

	// Approach 4: Combined approach (TLS 1.2 + explicit content type)
	t.Run("CombinedApproach", func(t *testing.T) {
		testCombinedApproachUpload(t, content)
	})
}

func testBasicMultipartUpload(t *testing.T, content []byte) {
	fmt.Println("Testing basic multipart upload...")

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", "test.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
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
	fmt.Printf("Response Body (first 500 chars): %s\n", string(body)[:min(500, len(body))])

	// Check if response is HTML or JSON
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		t.Logf("Received HTML response instead of JSON. Content-Type: %s", contentType)
	}

	if len(body) > 0 && body[0] == '<' {
		t.Logf("Response appears to be HTML (starts with '<')")
	}
}

func testTLS12MultipartUpload(t *testing.T, content []byte) {
	fmt.Println("Testing TLS 1.2 multipart upload...")

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	part, err := writer.CreateFormFile("file", "test.log")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Create client with TLS 1.2
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
	fmt.Printf("Response Body (first 500 chars): %s\n", string(body)[:min(500, len(body))])

	// Check if response is HTML or JSON
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		t.Logf("Received HTML response instead of JSON. Content-Type: %s", contentType)
	}

	if len(body) > 0 && body[0] == '<' {
		t.Logf("Response appears to be HTML (starts with '<')")
	}
}

func testExplicitContentTypeUpload(t *testing.T, content []byte) {
	fmt.Println("Testing explicit content type upload...")

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create form file with explicit content type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.log"`)
	h.Set("Content-Type", "text/plain")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("Failed to create form part: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send request
	client := &http.Client{
		Timeout: 30 * time.Second,
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
	fmt.Printf("Response Body (first 500 chars): %s\n", string(body)[:min(500, len(body))])

	// Check if response is HTML or JSON
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		t.Logf("Received HTML response instead of JSON. Content-Type: %s", contentType)
	}

	if len(body) > 0 && body[0] == '<' {
		t.Logf("Response appears to be HTML (starts with '<')")
	}
}

func testCombinedApproachUpload(t *testing.T, content []byte) {
	fmt.Println("Testing combined approach (TLS 1.2 + explicit content type)...")

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create form file with explicit content type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.log"`)
	h.Set("Content-Type", "text/plain")

	part, err := writer.CreatePart(h)
	if err != nil {
		t.Fatalf("Failed to create form part: %v", err)
	}

	_, err = part.Write(content)
	if err != nil {
		t.Fatalf("Failed to write content: %v", err)
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("User-Agent", "TheBoysLauncher/1.0")

	// Create client with TLS 1.2
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
	fmt.Printf("Response Body (first 500 chars): %s\n", string(body)[:min(500, len(body))])

	// Try to parse as JSON
	var result struct {
		OK  bool   `json:"ok"`
		URL string `json:"url"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		t.Logf("Failed to parse as JSON: %v", err)
		if len(body) > 0 && body[0] == '<' {
			t.Logf("Response appears to be HTML (starts with '<')")
			t.Logf("Full HTML response: %s", string(body))
		}
	} else {
		t.Logf("Successfully parsed JSON response: %+v", result)
		if result.OK {
			t.Logf("Upload successful! URL: %s", result.URL)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
