use crate::models::LauncherResult;
use std::path::Path;
use tokio::fs as async_fs;
use tokio::io::{AsyncReadExt, AsyncWriteExt};

/// Remove a directory and all its contents
pub async fn remove_dir_all<P: AsRef<Path>>(path: P) -> LauncherResult<()> {
    let path = path.as_ref();
    if path.exists() {
        async_fs::remove_dir_all(path).await?;
    }
    Ok(())
}

/// Copy a file with progress callback
pub async fn copy_file_with_progress<P: AsRef<Path>, Q: AsRef<Path>>(
    from: P,
    to: Q,
    mut progress_callback: Option<Box<dyn Fn(u64, u64) + Send>>,
) -> LauncherResult<u64> {
    let from = from.as_ref();
    let to = to.as_ref();

    // Ensure parent directory exists
    if let Some(parent) = to.parent() {
        async_fs::create_dir_all(parent).await?;
    }

    let from_file = async_fs::File::open(from).await?;
    let to_file = async_fs::File::create(to).await?;

    let mut reader = tokio::io::BufReader::new(from_file);
    let mut writer = tokio::io::BufWriter::new(to_file);

    let file_size = async_fs::metadata(from).await?.len();
    let mut copied = 0u64;

    const BUFFER_SIZE: usize = 8192;
    let mut buffer = vec![0u8; BUFFER_SIZE];

    loop {
        let bytes_read = reader.read(&mut buffer).await?;
        if bytes_read == 0 {
            break;
        }

        writer.write_all(&buffer[..bytes_read]).await?;
        copied += bytes_read as u64;

        if let Some(ref callback) = progress_callback {
            callback(copied, file_size);
        }
    }

    writer.flush().await?;
    Ok(copied)
}

/// Get directory size recursively
pub async fn get_dir_size<P: AsRef<Path>>(path: P) -> LauncherResult<u64> {
    let path = path.as_ref();
    let mut total_size = 0u64;

    let mut entries = async_fs::read_dir(path).await?;
    while let Some(entry) = entries.next_entry().await? {
        let metadata = entry.metadata().await?;
        if metadata.is_file() {
            total_size += metadata.len();
        } else if metadata.is_dir() {
            total_size += get_dir_size(entry.path()).await?;
        }
    }

    Ok(total_size)
}

/// Ensure directory exists
pub async fn ensure_dir<P: AsRef<Path>>(path: P) -> LauncherResult<()> {
    let path = path.as_ref();
    if !path.exists() {
        async_fs::create_dir_all(path).await?;
    }
    Ok(())
}

/// Check if path exists
pub async fn path_exists<P: AsRef<Path>>(path: P) -> bool {
    path.as_ref().exists()
}

/// Get file extension
pub fn get_file_extension<P: AsRef<Path>>(path: P) -> Option<String> {
    path.as_ref()
        .extension()
        .and_then(|ext| ext.to_str())
        .map(|s| s.to_lowercase())
}

/// Validate file name
pub fn is_valid_filename(name: &str) -> bool {
    !name.is_empty()
        && !name.contains("..")
        && !name.contains('/')
        && !name.contains('\\')
        && !name.contains(':')
        && !name.contains('*')
        && !name.contains('?')
        && !name.contains('"')
        && !name.contains('<')
        && !name.contains('>')
        && !name.contains('|')
}