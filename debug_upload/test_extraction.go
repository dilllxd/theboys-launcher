package main

import (
	"fmt"
	"regexp"
)

// extractFileIDFromHTML extracts the file ID from MicroBin's HTML response
func extractFileIDFromHTML(html string) string {
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

func main() {
	// Sample HTML response from logs.dylan.lol
	html := `<!DOCTYPE html>
<html>

<head>
    
    <title>MicroBin</title>

    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link rel="icon" type="image/svg+xml" href="https://logs.dylan.lol/static/favicon.ico">

    <script type="text/javascript" src="https://logs.dylan.lol/static/aes.js"></script>
     
    <link rel="stylesheet" href="https://logs.dylan.lol/static/water.css">

</head>


    <body style=" max-width: 800px; margin: auto; padding-left:0.5rem;
    padding-right:0.5rem; padding-top: 2rem; line-height: 1.5; font-size: 1.1em; ">
        <br>
        

        <div id="nav" style="margin-bottom: 1rem;">
            <b style="margin-right: 0.5rem">
                
                <!-- <i><span style="font-size:2.2rem;
                margin-right:1rem">μ</span></i> -->
                <a href="/"><img width=100 style="margin-bottom: -6px; margin-right:
                0.5rem;" src="https://logs.dylan.lol/static/logo.png"></a> 
            </b>

            <a href="https://logs.dylan.lol/" style="margin-right: 0.5rem;
            margin-left: 0.5rem">New</a>

            
            <a href="https://logs.dylan.lol/list" style="margin-right: 0.5rem; margin-left: 0.5rem">List</a>

            <a href="https://logs.dylan.lol/guide" style="margin-right: 0.5rem;
            margin-left: 0.5rem">Guide</a>


        </div>

        <!-- <hr> -->
<div style="float: left">
    
  <a style="margin-right: 1rem" href="https://logs.dylan.lol/edit/monkey-turkey-duck">Edit</a>
  <a style="margin-right: 1rem" href="https://logs.dylan.lol/remove/monkey-turkey-duck">Remove</a>
</div>
<div style="float: right">
  <a style="margin-right: 0.5rem"
    href="https://logs.dylan.lol/upload/monkey-turkey-duck"><i>monkey-turkey-duck</i></a>
  
  <button id="copy-url-button" class="small-button" style="margin-right: 0">
    Copy URL
  </button>
</div>

<br>
<br>




<br>




<span style="margin-left: auto; margin-right: auto; display: flex;
    justify-content: center; align-items: center;">
  <p style="font-size: small;">test.log
    [83 B]</p>
  <a href="https://logs.dylan.lol/file/monkey-turkey-duck" id="download-link">
    <button class="download-button" autofocus>
      Download
    </button>
  </a>
</span>




<div>
  
</div>

<br>

<script type="text/javascript" src="https://logs.dylan.lol/static/highlight/highlight.min.js"></script>
<link rel="stylesheet" href="https://logs.dylan.lol/static/highlight/highlight.min.css">

<script>
  const copyURLBtn = document.getElementById("copy-url-button")
  const copyTextBtn = document.getElementById("copy-text-button")
  const copyRedirectBtn = document.getElementById("copy-redirect-button")
  var content = ""
  const contentElement = document.getElementById("code");
  const url = ("https://logs.dylan.lol" === "") ? "https://logs.dylan.lol/upload/monkey-turkey-duck" : "https://logs.dylan.lol/p/monkey-turkey-duck"
  const redirect_url = ("https://logs.dylan.lol" === "") ? "https://logs.dylan.lol/url/monkey-turkey-duck" : "https://logs.dylan.lol/u/monkey-turkey-duck"

  const te = new TextEncoder();

  //  

  function escapeHtml(unsafe) {
    return unsafe
      .replace(/&/g, "&")
      .replace(/</g, "<")
      .replace(/>/g, ">")
      .replace(/"/g, """)
      .replace(/'/g, "&#039;");
  }

  function wrapStringInCodeLines(str) {
    const lines = str.split(/\r?\n/); // split string into an array of lines
    const wrappedLines = lines.map((line) => "<code-line>" + line + "</code-line>"); // wrap each line in a "code-line" tag
    return wrappedLines.join("\n"); // join the wrapped lines back into a single string with line breaks
  }

  const decodeEntity = (inputStr) => {
    var textarea = document.createElement("textarea");
    textarea.innerHTML = inputStr;
    return textarea.value;
  }

  if (copyURLBtn) {
    copyURLBtn.addEventListener("click", () => {
      navigator.clipboard.writeText(url)
      copyURLBtn.innerHTML = "Copied"
      setTimeout(() => {
        copyURLBtn.innerHTML = "Copy URL"
      }, 1000)
    })
  }

  // it will be undefined when the element does not exist on non-url pastas
  if (copyRedirectBtn) {
    copyRedirectBtn.addEventListener("click", () => {
      navigator.clipboard.writeText(redirect_url)
      copyRedirectBtn.innerHTML = "Copied"
      setTimeout(() => {
        copyRedirectBtn.innerHTML = "Copy Redirect"
      }, 1000)
    })
  }

  if (copyTextBtn) {
    copyTextBtn.addEventListener("click", () => {
      const decodeContent = decodeEntity(content)
      navigator.clipboard.writeText(decodeContent)
      copyTextBtn.innerHTML = "Copied"
      setTimeout(() => {
        copyTextBtn.innerHTML = "Copy Text"
      }, 1000)
    })
  }

  //  
</script>`

	// Test the extraction
	fileID := extractFileIDFromHTML(html)
	if fileID != "" {
		fmt.Printf("✅ Successfully extracted file ID: %s\n", fileID)
		fmt.Printf("🔗 Direct URL: https://logs.dylan.lol/p/%s\n", fileID)
		fmt.Printf("📄 File URL: https://logs.dylan.lol/file/%s\n", fileID)
	} else {
		fmt.Println("❌ Failed to extract file ID")
	}
}
