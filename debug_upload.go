package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"regexp"
	"strings"
	"time"
)

func main() {
	// Create a small test file content
	testContent := "This is a test log file for debugging upload.\nTimestamp: " + time.Now().Format(time.RFC3339) + "\n"

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Create form file with explicit content type
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="test.log"`)
	h.Set("Content-Type", "text/plain")

	part, err := writer.CreatePart(h)
	if err != nil {
		fmt.Printf("Failed to create form part: %v\n", err)
		return
	}

	_, err = part.Write([]byte(testContent))
	if err != nil {
		fmt.Printf("Failed to write content: %v\n", err)
		return
	}

	writer.Close()

	// Create request
	req, err := http.NewRequest("POST", "https://logs.dylan.lol/upload", &requestBody)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
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

	fmt.Println("Sending request to https://logs.dylan.lol/upload...")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to send request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Failed to read response body: %v\n", err)
		return
	}

	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Headers: %+v\n", resp.Header)
	fmt.Printf("Content-Type: %s\n", resp.Header.Get("Content-Type"))
	fmt.Printf("Response Body (first 1000 chars):\n%s\n", string(body)[:min(1000, len(body))])

	// Check if response is HTML or JSON
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(strings.ToLower(contentType), "text/html") {
		fmt.Println("\n*** Response is HTML ***")
	} else if strings.Contains(strings.ToLower(contentType), "application/json") {
		fmt.Println("\n*** Response is JSON ***")
	}

	if len(body) > 0 && body[0] == '<' {
		fmt.Println("*** Response starts with '<' - likely HTML ***")
	}

	// Try to extract file ID using current patterns
	html := string(body)
	fmt.Println("\n*** Testing extraction patterns ***")

	// Pattern 1: href="/upload/([^"]+)"
	uploadPattern := `href="/upload/([^"]+)"`
	if matches := regexp.MustCompile(uploadPattern).FindStringSubmatch(html); len(matches) > 1 {
		fmt.Printf("Pattern 1 matched: %s\n", matches[1])
	} else {
		fmt.Println("Pattern 1: No match")
	}

	// Pattern 2: href="/file/([^"]+)"
	filePattern := `href="/file/([^"]+)"`
	if matches := regexp.MustCompile(filePattern).FindStringSubmatch(html); len(matches) > 1 {
		fmt.Printf("Pattern 2 matched: %s\n", matches[1])
	} else {
		fmt.Println("Pattern 2: No match")
	}

	// Pattern 3: const url = \(.*logs\.dylan\.lol.*\)/([^"]+)"
	copyPattern := `const url = \(.*logs\.dylan\.lol.*\)/([^"]+)"`
	if matches := regexp.MustCompile(copyPattern).FindStringSubmatch(html); len(matches) > 1 {
		fmt.Printf("Pattern 3 matched: %s\n", matches[1])
	} else {
		fmt.Println("Pattern 3: No match")
	}

	// Save full response to file for analysis
	err = os.WriteFile("debug_response.html", body, 0644)
	if err != nil {
		fmt.Printf("Failed to save response to file: %v\n", err)
	} else {
		fmt.Println("\nFull response saved to debug_response.html")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
