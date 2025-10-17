use crate::models::{LauncherResult, LauncherError};
use reqwest::Client;
use std::time::Duration;
use url::Url;
use tokio::io::AsyncWriteExt;

/// Create a HTTP client with appropriate settings
pub fn create_http_client() -> LauncherResult<Client> {
    let client = Client::builder()
        .timeout(Duration::from_secs(30))
        .user_agent(get_user_agent())
        .build()?;

    Ok(client)
}

/// Get user agent string for HTTP requests
fn get_user_agent() -> String {
    format!("TheBoys-Launcher/{}", env!("CARGO_PKG_VERSION"))
}

/// Validate a URL
pub fn validate_url(url: &str) -> LauncherResult<()> {
    Url::parse(url)
        .map_err(|e| LauncherError::InvalidConfig(format!("Invalid URL: {}", e)))?;
    Ok(())
}

/// Check if URL is reachable
pub async fn check_url_reachable(url: &str) -> LauncherResult<bool> {
    let client = create_http_client()?;
    let response = client.head(url).send().await;

    match response {
        Ok(resp) => Ok(resp.status().is_success()),
        Err(_) => Ok(false),
    }
}

/// Download a file with progress reporting
pub async fn download_file_with_progress<F>(
    url: &str,
    destination: &str,
    progress_callback: F,
) -> LauncherResult<()>
where
    F: Fn(u64, u64) + Send + Sync + 'static,
{
    let client = create_http_client()?;
    validate_url(url)?;

    // Create parent directories
    if let Some(parent) = std::path::Path::new(destination).parent() {
        tokio::fs::create_dir_all(parent).await?;
    }

    // Start download
    let response = client.get(url).send().await?;
    let total_size = response.content_length().unwrap_or(0);

    let mut downloaded = 0u64;
    let mut file = tokio::fs::File::create(destination).await?;
    let mut stream = response.bytes_stream();

    use futures_util::StreamExt;
    while let Some(chunk) = stream.next().await {
        let chunk = chunk?;
        file.write_all(&chunk).await?;
        downloaded += chunk.len() as u64;
        progress_callback(downloaded, total_size);
    }

    Ok(())
}