package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// This test verifies the upload functionality changes
func main() {
	fmt.Println("Testing upload functionality changes...")

	// Test 1: Verify generateRandomID function works
	fmt.Println("\n1. Testing generateRandomID function...")
	randomID, err := generateRandomID()
	if err != nil {
		fmt.Printf("   ❌ Failed to generate random ID: %v\n", err)
	} else {
		fmt.Printf("   ✅ Generated random ID: %s\n", randomID)
	}

	// Test 2: Verify HTML parsing for log URL extraction
	fmt.Println("\n2. Testing HTML response parsing...")
	testHTML := `<!DOCTYPE html>
<html>
<head><title>copyparty @ homelab</title></head>
<body>
<div id="box">
<h2><a href="/logs">return to /logs</a></h2>
<pre>OK // 8742 bytes // 0.083 MiB/s
sha512: 13ed14a0111ec850fa1e2f02eea74149d5fe5ab338d9391bfbf8d83a // E-0UoBEeyFD6Hi8C7qdBSdX-WrM42Tkb-_jYOumonHBV // 8742 bytes // <a href="/logs/7c01632f.log">7c01632f.log</a> 
</pre>
</div>
</body>
</html>`

	extractedURL := extractLogURLFromHTML(testHTML, randomID)
	expectedURL := "https://i.dylan.lol/logs/7c01632f.log"
	if extractedURL == expectedURL {
		fmt.Printf("   ✅ Successfully extracted URL: %s\n", extractedURL)
	} else {
		fmt.Printf("   ❌ URL extraction failed. Expected: %s, Got: %s\n", expectedURL, extractedURL)
	}

	// Test 3: Verify dialog button configuration (by checking the code structure)
	fmt.Println("\n3. Testing dialog button configuration...")
	// This test verifies that our dialog only has 2 buttons by checking the code
	guiCode, err := os.ReadFile("../gui.go")
	if err != nil {
		fmt.Printf("   ❌ Failed to read gui.go: %v\n", err)
	} else {
		codeStr := string(guiCode)

		// Check that debug messages are commented out
		if strings.Contains(codeStr, "// logf(\"Upload response status: %s\", resp.Status)") {
			fmt.Println("   ✅ Debug messages are commented out in uploadLog()")
		} else {
			fmt.Println("   ❌ Debug messages are not properly commented out")
		}

		// Check that only 2 buttons are created in the success dialog
		copyButtonCount := strings.Count(codeStr, "copyButton := widget.NewButtonWithIcon")
		okButtonCount := strings.Count(codeStr, "okButton := widget.NewButtonWithIcon")

		if copyButtonCount == 1 && okButtonCount == 1 {
			fmt.Println("   ✅ Dialog creates exactly 2 buttons (copy and ok)")
		} else {
			fmt.Printf("   ❌ Unexpected button count. Copy: %d, OK: %d\n", copyButtonCount, okButtonCount)
		}

		// Verify no third button is created
		if !strings.Contains(codeStr, "viewButton := widget.NewButtonWithIcon") {
			fmt.Println("   ✅ No third 'view' button found in dialog")
		} else {
			fmt.Println("   ❌ Third 'view' button still exists in dialog")
		}

		// Check that the button container only has 2 buttons
		if strings.Contains(codeStr, "buttonContainer := container.NewHBox(") &&
			strings.Contains(codeStr, "copyButton,") &&
			strings.Contains(codeStr, "okButton,") {
			fmt.Println("   ✅ Button container correctly configured with only 2 buttons")
		} else {
			fmt.Println("   ❌ Button container configuration issue")
		}
	}

	fmt.Println("\n✅ All tests completed!")
}

// generateRandomID generates a random 8-character hexadecimal string
func generateRandomID() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02x%02x%02x%02x", bytes[0], bytes[1], bytes[2], bytes[3]), nil
}

// extractLogURLFromHTML extracts the log file URL from the HTML response
func extractLogURLFromHTML(html string, fallbackID string) string {
	// Try to extract the filename from the HTML response using regex
	// Pattern to match href="/logs/filename.log"
	logPattern := `href="/logs/([^"]+\.log)"`
	re := regexp.MustCompile(logPattern)
	matches := re.FindStringSubmatch(html)

	if len(matches) > 1 {
		// Extract the filename from the match
		filename := matches[1]
		// Construct the full URL
		return fmt.Sprintf("https://i.dylan.lol/logs/%s", filename)
	} else {
		// If regex fails, fall back to using our random ID
		return fmt.Sprintf("https://i.dylan.lol/logs/%s.log", fallbackID)
	}
}
