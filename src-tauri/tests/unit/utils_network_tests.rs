#[cfg(test)]
mod tests {
    use std::time::Duration;
    use tempfile::TempDir;
    use tokio::fs;
    use mockito::Server;
    use theboys_launcher::utils::network::*;
    use theboys_launcher::models::{LauncherError, LauncherResult};

    #[tokio::test]
    fn test_create_http_client() {
        let client = create_http_client();
        assert!(client.is_ok());
    }

    #[tokio::test]
    fn test_get_user_agent() {
        let user_agent = get_user_agent();
        assert!(user_agent.starts_with("TheBoys-Launcher/"));
        assert!(user_agent.contains("1.1.0")); // Default version from Cargo.toml
    }

    #[tokio::test]
    fn test_validate_url_valid_urls() {
        let valid_urls = vec![
            "https://www.example.com",
            "http://localhost:8080",
            "https://api.github.com/repos/user/repo",
            "ftp://files.example.com/path",
            "https://example.com:443/path?query=value#fragment",
        ];

        for url in valid_urls {
            let result = validate_url(url);
            assert!(result.is_ok(), "URL should be valid: {}", url);
        }
    }

    #[tokio::test]
    fn test_validate_url_invalid_urls() {
        let invalid_urls = vec![
            "not-a-url",
            "://missing-protocol",
            "http://",
            "https://",
            "javascript:alert('xss')",
            "",
            " ",
            "ht tp://invalid-spaces.com",
        ];

        for url in invalid_urls {
            let result = validate_url(url);
            assert!(result.is_err(), "URL should be invalid: {}", url);

            // Check that it returns the correct error type
            match result.unwrap_err() {
                LauncherError::InvalidConfig(_) => {}, // Expected
                other => panic!("Expected InvalidConfig error for URL: {}, got: {:?}", url, other),
            }
        }
    }

    #[tokio::test]
    async fn test_check_url_reachable_success() {
        let mut server = Server::new();

        // Mock endpoint that returns 200 OK
        let mock = server.mock("HEAD", "/test")
            .with_status(200)
            .create();

        let url = format!("{}/test", server.url());
        let is_reachable = check_url_reachable(&url).await.unwrap();

        assert!(is_reachable);
        mock.assert();
    }

    #[tokio::test]
    async fn test_check_url_reachable_not_found() {
        let mut server = Server::new();

        // Mock endpoint that returns 404 Not Found
        let mock = server.mock("HEAD", "/notfound")
            .with_status(404)
            .create();

        let url = format!("{}/notfound", server.url());
        let is_reachable = check_url_reachable(&url).await.unwrap();

        assert!(!is_reachable);
        mock.assert();
    }

    #[tokio::test]
    async fn test_check_url_reachable_server_error() {
        let mut server = Server::new();

        // Mock endpoint that returns 500 Internal Server Error
        let mock = server.mock("HEAD", "/error")
            .with_status(500)
            .create();

        let url = format!("{}/error", server.url());
        let is_reachable = check_url_reachable(&url).await.unwrap();

        assert!(!is_reachable);
        mock.assert();
    }

    #[tokio::test]
    async fn test_check_url_reachable_invalid_server() {
        let url = "http://localhost:99999/nonexistent"; // Non-existent server

        let is_reachable = check_url_reachable(url).await.unwrap();

        assert!(!is_reachable);
    }

    #[tokio::test]
    async fn test_download_file_with_progress_success() {
        let mut server = Server::new();
        let test_content = "This is test file content for download testing.";

        // Mock endpoint with file content
        let mock = server.mock("GET", "/download")
            .with_status(200)
            .with_header("content-type", "text/plain")
            .with_body(test_content)
            .create();

        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("downloaded_file.txt");

        let url = format!("{}/download", server.url());

        // Track progress calls
        let mut progress_calls = Vec::new();
        let progress_callback = |downloaded: u64, total: u64| {
            progress_calls.push((downloaded, total));
        };

        // Download file
        let result = download_file_with_progress(&url, destination.to_str().unwrap(), progress_callback).await;

        assert!(result.is_ok());

        // Verify file was downloaded
        assert!(destination.exists());
        let downloaded_content = fs::read_to_string(&destination).await.unwrap();
        assert_eq!(downloaded_content, test_content);

        // Verify progress was reported
        assert!(!progress_calls.is_empty());
        assert_eq!(progress_calls.last().unwrap().0, test_content.len() as u64);

        mock.assert();
    }

    #[tokio::test]
    async fn test_download_file_with_progress_content_length() {
        let mut server = Server::new();
        let test_content = "Content with known length";

        // Mock endpoint with content-length header
        let mock = server.mock("GET", "/download-length")
            .with_status(200)
            .with_header("content-length", &test_content.len().to_string())
            .with_body(test_content)
            .create();

        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("file_with_length.txt");

        let url = format!("{}/download-length", server.url());

        let mut total_size_reported = None;
        let progress_callback = |downloaded: u64, total: u64| {
            total_size_reported = Some(total);
        };

        // Download file
        download_file_with_progress(&url, destination.to_str().unwrap(), progress_callback).await.unwrap();

        // Verify content length was reported
        assert_eq!(total_size_reported, Some(test_content.len() as u64));

        mock.assert();
    }

    #[tokio::test]
    async fn test_download_file_creates_parent_directories() {
        let mut server = Server::new();
        let test_content = "Test content for directory creation";

        let mock = server.mock("GET", "/test")
            .with_status(200)
            .with_body(test_content)
            .create();

        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("nested").join("directory").join("file.txt");

        let url = format!("{}/test", server.url());

        // Download file (should create parent directories)
        download_file_with_progress(&url, destination.to_str().unwrap(), |_, _| {}).await.unwrap();

        // Verify file was downloaded and directories were created
        assert!(destination.exists());
        let downloaded_content = fs::read_to_string(&destination).await.unwrap();
        assert_eq!(downloaded_content, test_content);

        mock.assert();
    }

    #[tokio::test]
    async fn test_download_file_invalid_url() {
        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("file.txt");

        let invalid_url = "not-a-valid-url";

        let result = download_file_with_progress(invalid_url, destination.to_str().unwrap(), |_, _| {}).await;

        assert!(result.is_err());
        match result.unwrap_err() {
            LauncherError::InvalidConfig(_) => {}, // Expected
            other => panic!("Expected InvalidConfig error, got: {:?}", other),
        }
    }

    #[tokio::test]
    async fn test_download_file_not_found() {
        let mut server = Server::new();

        // Mock endpoint that returns 404
        let mock = server.mock("GET", "/notfound")
            .with_status(404)
            .create();

        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("not_found.txt");

        let url = format!("{}/notfound", server.url());

        let result = download_file_with_progress(&url, destination.to_str().unwrap(), |_, _| {}).await;

        assert!(result.is_err());
        assert!(!destination.exists());

        mock.assert();
    }

    #[tokio::test]
    async fn test_download_file_server_error() {
        let mut server = Server::new();

        // Mock endpoint that returns 500
        let mock = server.mock("GET", "/error")
            .with_status(500)
            .create();

        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("error.txt");

        let url = format!("{}/error", server.url());

        let result = download_file_with_progress(&url, destination.to_str().unwrap(), |_, _| {}).await;

        assert!(result.is_err());
        assert!(!destination.exists());

        mock.assert();
    }

    #[tokio::test]
    async fn test_http_client_timeout() {
        // This test verifies that the HTTP client has a timeout configured
        // We can't easily test actual timeout behavior without slowing down tests
        let client = create_http_client().unwrap();

        // Get the timeout from the client (this is a bit of a hack since reqwest doesn't expose it directly)
        // For now, we just verify the client was created successfully
        assert!(client.timeout().is_some());
    }

    #[tokio::test]
    async fn test_download_large_file_progress_tracking() {
        let mut server = Server::new();

        // Create larger content to test multiple progress updates
        let large_content = "A".repeat(8192 * 10); // 80KB of content

        let mock = server.mock("GET", "/large")
            .with_status(200)
            .with_header("content-length", &large_content.len().to_string())
            .with_body(&large_content)
            .create();

        let temp_dir = TempDir::new().unwrap();
        let destination = temp_dir.path().join("large_file.txt");

        let url = format!("{}/large", server.url());

        let mut progress_updates = 0;
        let mut last_downloaded = 0u64;
        let progress_callback = |downloaded: u64, total: u64| {
            progress_updates += 1;
            assert!(downloaded <= total);
            assert!(downloaded >= last_downloaded);
            last_downloaded = downloaded;
        };

        // Download large file
        download_file_with_progress(&url, destination.to_str().unwrap(), progress_callback).await.unwrap();

        // Verify we received multiple progress updates
        assert!(progress_updates > 1);

        // Verify final state
        assert_eq!(last_downloaded, large_content.len() as u64);

        mock.assert();
    }
}