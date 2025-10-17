use crate::models::{LauncherResult, DownloadProgress, DownloadStatus};
use std::sync::Arc;
use tokio::sync::{mpsc, RwLock};
use tokio::time::{Duration, Instant};
use futures_util::StreamExt;
use uuid::Uuid;
use std::path::Path;
use tokio::fs::File;
use tokio::io::AsyncWriteExt;

/// Individual download task
#[derive(Debug, Clone)]
pub struct DownloadTask {
    pub id: String,
    pub name: String,
    pub url: String,
    pub destination: String,
    pub total_bytes: u64,
    pub downloaded_bytes: u64,
    pub status: DownloadStatus,
    pub created_at: Instant,
    pub started_at: Option<Instant>,
    pub completed_at: Option<Instant>,
    pub retry_count: u32,
    pub max_retries: u32,
}

impl DownloadTask {
    pub fn new(name: String, url: String, destination: String) -> Self {
        let id = format!("download-{}", Uuid::new_v4());
        Self {
            id,
            name,
            url,
            destination,
            total_bytes: 0,
            downloaded_bytes: 0,
            status: DownloadStatus::Pending,
            created_at: Instant::now(),
            started_at: None,
            completed_at: None,
            retry_count: 0,
            max_retries: 3,
        }
    }

    pub fn progress_percent(&self) -> f64 {
        if self.total_bytes == 0 {
            0.0
        } else {
            (self.downloaded_bytes as f64 / self.total_bytes as f64) * 100.0
        }
    }

    pub fn elapsed_time(&self) -> Duration {
        self.started_at
            .map(|started| Instant::now() - started)
            .unwrap_or(Duration::ZERO)
    }

    pub fn speed_bps(&self) -> u64 {
        let elapsed = self.elapsed_time().as_secs_f64();
        if elapsed > 0.0 {
            (self.downloaded_bytes as f64 / elapsed) as u64
        } else {
            0
        }
    }
}

/// Download manager for handling multiple downloads
pub struct DownloadManager {
    active_downloads: Arc<RwLock<std::collections::HashMap<String, Arc<tokio::sync::Mutex<DownloadTask>>>>>,
    download_queue: Arc<RwLock<Vec<DownloadTask>>>,
    max_concurrent: usize,
    download_tx: mpsc::UnboundedSender<String>,
    active_downloads_count: Arc<RwLock<usize>>,
}

impl DownloadManager {
    pub fn new() -> Self {
        let (download_tx, _download_rx) = mpsc::unbounded_channel();
        Self {
            active_downloads: Arc::new(RwLock::new(std::collections::HashMap::new())),
            download_queue: Arc::new(RwLock::new(Vec::new())),
            max_concurrent: 3, // Default to 3 concurrent downloads
            download_tx,
            active_downloads_count: Arc::new(RwLock::new(0)),
        }
    }

    /// Set maximum concurrent downloads
    pub fn set_max_concurrent(&mut self, max: usize) {
        self.max_concurrent = max;
    }

    /// Start a new download
    pub async fn start_download(
        &self,
        name: String,
        url: String,
        destination: String,
    ) -> LauncherResult<String> {
        let task = DownloadTask::new(name, url, destination);
        let task_id = task.id.clone();

        // Add to queue
        {
            let mut queue = self.download_queue.write().await;
            queue.push(task);
        }

        // Try to start the download if there's capacity
        self.try_start_next_download().await?;

        Ok(task_id)
    }

    /// Try to start the next download in queue if there's capacity
    async fn try_start_next_download(&self) -> LauncherResult<()> {
        loop {
            let active_count = *self.active_downloads_count.read().await;

            if active_count >= self.max_concurrent {
                break Ok(());
            }

            // Get next task from queue
            let task = {
                let mut queue = self.download_queue.write().await;
                if queue.is_empty() {
                    break Ok(());
                }
                queue.remove(0)
            };

            let task_id = task.id.clone();
            let task_arc = Arc::new(tokio::sync::Mutex::new(task));

            // Add to active downloads
            {
                let mut active = self.active_downloads.write().await;
                active.insert(task_id.clone(), task_arc.clone());
            }

            // Increment active count
            {
                let mut count = self.active_downloads_count.write().await;
                *count += 1;
            }

            // Execute the download
            let result = self.execute_download(task_arc.clone()).await;

            // Handle completion
            {
                let mut active = self.active_downloads.write().await;
                active.remove(&task_id);
            }

            {
                let mut count = self.active_downloads_count.write().await;
                *count = count.saturating_sub(1);
            }

            // Notify completion
            if let Err(e) = self.download_tx.send(task_id.clone()) {
                tracing::error!("Failed to send download completion notification: {}", e);
            }

            if let Err(e) = result {
                tracing::error!("Download {} failed: {}", task_id, e);
            }

            // Continue loop to try next download
        }
    }

    /// Execute a download
    async fn execute_download(&self, task_arc: Arc<tokio::sync::Mutex<DownloadTask>>) -> LauncherResult<()> {
        let (url, destination) = {
            let task = task_arc.lock().await;
            (task.url.clone(), task.destination.clone())
        };

        // Update status to downloading
        {
            let mut task = task_arc.lock().await;
            task.status = DownloadStatus::Downloading;
            task.started_at = Some(Instant::now());
        }

        // Create parent directories if they don't exist
        if let Some(parent) = Path::new(&destination).parent() {
            tokio::fs::create_dir_all(parent).await
                .map_err(|e| crate::models::LauncherError::FileSystem(
                    format!("Failed to create directory {}: {}", parent.display(), e)
                ))?;
        }

        // Perform download with retry logic
        let mut retry_count = 0;
        loop {
            match self.download_with_progress(&url, &destination, task_arc.clone()).await {
                Ok(_) => {
                    // Mark as completed
                    let mut task = task_arc.lock().await;
                    task.status = DownloadStatus::Completed;
                    task.completed_at = Some(Instant::now());
                    break Ok(());
                }
                Err(e) => {
                    retry_count += 1;
                    let mut task = task_arc.lock().await;
                    task.retry_count = retry_count;

                    if retry_count >= task.max_retries {
                        task.status = DownloadStatus::Failed(e.to_string());
                        break Err(e);
                    }

                    // Exponential backoff
                    let delay = Duration::from_secs(2u64.pow(retry_count - 1));
                    tokio::time::sleep(delay).await;
                }
            }
        }
    }

    /// Download with progress tracking
    async fn download_with_progress(
        &self,
        url: &str,
        destination: &str,
        task_arc: Arc<tokio::sync::Mutex<DownloadTask>>,
    ) -> LauncherResult<()> {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(30))
            .user_agent("TheBoys-Launcher/1.1.0")
            .build()?;

        let response = client.get(url).send().await?;

        if !response.status().is_success() {
            return Err(crate::models::LauncherError::DownloadFailed(
                format!("HTTP {}: {}", response.status(), url)
            ));
        }

        let total_bytes = response.content_length().unwrap_or(0);

        // Update task with total bytes
        {
            let mut task = task_arc.lock().await;
            task.total_bytes = total_bytes;
        }

        let mut file = File::create(destination).await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to create file {}: {}", destination, e)
            ))?;

        let mut downloaded_bytes: u64 = 0;
        let mut last_update = Instant::now();

        let mut stream = response.bytes_stream();
        while let Some(chunk) = stream.next().await {
            let chunk = chunk.map_err(|e| crate::models::LauncherError::DownloadFailed(
                format!("Download stream error: {}", e)
            ))?;

            file.write_all(&chunk).await
                .map_err(|e| crate::models::LauncherError::FileSystem(
                    format!("Failed to write to file {}: {}", destination, e)
                ))?;

            downloaded_bytes += chunk.len() as u64;

            // Update progress (but not too frequently)
            if last_update.elapsed() >= Duration::from_millis(100) {
                {
                    let mut task = task_arc.lock().await;
                    task.downloaded_bytes = downloaded_bytes;
                }
                last_update = Instant::now();
            }
        }

        // Final update
        {
            let mut task = task_arc.lock().await;
            task.downloaded_bytes = downloaded_bytes;
        }

        file.flush().await
            .map_err(|e| crate::models::LauncherError::FileSystem(
                format!("Failed to flush file {}: {}", destination, e)
            ))?;

        Ok(())
    }

    /// Get download progress
    pub async fn get_progress(&self, id: &str) -> Option<DownloadProgress> {
        let active = self.active_downloads.read().await;
        if let Some(task_arc) = active.get(id) {
            let task = task_arc.lock().await;
            return Some(DownloadProgress {
                id: task.id.clone(),
                name: task.name.clone(),
                downloaded_bytes: task.downloaded_bytes,
                total_bytes: task.total_bytes,
                progress_percent: task.progress_percent(),
                speed_bps: task.speed_bps(),
                status: task.status.clone(),
            });
        }

        // Check queue for pending downloads
        let queue = self.download_queue.read().await;
        if let Some(task) = queue.iter().find(|t| t.id == id) {
            return Some(DownloadProgress {
                id: task.id.clone(),
                name: task.name.clone(),
                downloaded_bytes: task.downloaded_bytes,
                total_bytes: task.total_bytes,
                progress_percent: task.progress_percent(),
                speed_bps: task.speed_bps(),
                status: task.status.clone(),
            });
        }

        None
    }

    /// Cancel a download
    pub async fn cancel_download(&self, id: &str) -> LauncherResult<()> {
        // Remove from active downloads
        {
            let mut active = self.active_downloads.write().await;
            if let Some(task_arc) = active.remove(id) {
                let mut task = task_arc.lock().await;
                task.status = DownloadStatus::Cancelled;

                // Decrement active count
                let mut count = self.active_downloads_count.write().await;
                *count = count.saturating_sub(1);
            }
        }

        // Remove from queue if pending
        {
            let mut queue = self.download_queue.write().await;
            queue.retain(|task| task.id != id);
        }

        // Delete partial file
        if let Some(progress) = self.get_progress(id).await {
            if progress.status != DownloadStatus::Completed {
                if let Err(e) = tokio::fs::remove_file(&progress.name).await {
                    tracing::warn!("Failed to remove partial download file: {}", e);
                }
            }
        }

        Ok(())
    }

    /// Pause a download (remove from active and add back to queue)
    pub async fn pause_download(&self, id: &str) -> LauncherResult<()> {
        let task_arc = {
            let mut active = self.active_downloads.write().await;
            active.remove(id)
        };

        if let Some(task_arc) = task_arc {
            let mut task = task_arc.lock().await;
            task.status = DownloadStatus::Pending;

            // Add back to queue
            let mut queue = self.download_queue.write().await;
            queue.push(task.clone());

            // Decrement active count
            let mut count = self.active_downloads_count.write().await;
            *count = count.saturating_sub(1);
        }

        Ok(())
    }

    /// Resume a paused download
    pub async fn resume_download(&self, id: &str) -> LauncherResult<()> {
        // Find in queue and move to active
        {
            let mut queue = self.download_queue.write().await;
            if let Some(pos) = queue.iter().position(|t| t.id == id) {
                let mut task = queue.remove(pos);
                task.status = DownloadStatus::Pending;
                queue.push(task);
            }
        }

        // Try to start downloads
        self.try_start_next_download().await?;

        Ok(())
    }

    /// Remove completed download from tracking
    pub async fn remove_download(&self, id: &str) -> LauncherResult<()> {
        {
            let mut active = self.active_downloads.write().await;
            active.remove(id);
        }

        {
            let mut queue = self.download_queue.write().await;
            queue.retain(|task| task.id != id);
        }

        Ok(())
    }

    /// Get all active downloads
    pub async fn get_all_downloads(&self) -> Vec<DownloadProgress> {
        let mut downloads = Vec::new();

        // Get active downloads
        {
            let active = self.active_downloads.read().await;
            for task_arc in active.values() {
                let task = task_arc.lock().await;
                downloads.push(DownloadProgress {
                    id: task.id.clone(),
                    name: task.name.clone(),
                    downloaded_bytes: task.downloaded_bytes,
                    total_bytes: task.total_bytes,
                    progress_percent: task.progress_percent(),
                    speed_bps: task.speed_bps(),
                    status: task.status.clone(),
                });
            }
        }

        // Get queued downloads
        {
            let queue = self.download_queue.read().await;
            for task in queue.iter() {
                downloads.push(DownloadProgress {
                    id: task.id.clone(),
                    name: task.name.clone(),
                    downloaded_bytes: task.downloaded_bytes,
                    total_bytes: task.total_bytes,
                    progress_percent: task.progress_percent(),
                    speed_bps: task.speed_bps(),
                    status: task.status.clone(),
                });
            }
        }

        downloads
    }

    /// Get download completion notifications
    pub async fn get_completion_receiver(&self) -> Option<mpsc::UnboundedReceiver<String>> {
        // Since we can't store the receiver in the struct due to Send issues,
        // this method is not implemented. For completion notifications,
        // use a different approach like polling or events.
        None
    }
}

impl Clone for DownloadManager {
    fn clone(&self) -> Self {
        Self {
            active_downloads: self.active_downloads.clone(),
            download_queue: self.download_queue.clone(),
            max_concurrent: self.max_concurrent,
            download_tx: self.download_tx.clone(),
            active_downloads_count: self.active_downloads_count.clone(),
        }
    }
}

impl Default for DownloadManager {
    fn default() -> Self {
        Self::new()
    }
}

/// Global download manager instance
pub static DOWNLOAD_MANAGER: std::sync::LazyLock<DownloadManager> =
    std::sync::LazyLock::new(DownloadManager::new);

/// Get the global download manager
pub fn download_manager() -> &'static DownloadManager {
    &DOWNLOAD_MANAGER
}