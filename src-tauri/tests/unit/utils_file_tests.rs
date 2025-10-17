#[cfg(test)]
mod tests {
    use std::path::Path;
    use tempfile::TempDir;
    use tokio::fs;
    use tokio::io::AsyncWriteExt;
    use theboys_launcher::utils::file::*;

    #[tokio::test]
    async fn test_remove_dir_all() {
        let temp_dir = TempDir::new().unwrap();
        let test_path = temp_dir.path().join("test_dir");

        // Create test directory with files
        fs::create_dir_all(&test_path).await.unwrap();
        fs::write(test_path.join("test.txt"), "test content").await.unwrap();

        // Verify directory exists
        assert!(test_path.exists());

        // Remove directory
        remove_dir_all(&test_path).await.unwrap();

        // Verify directory is removed
        assert!(!test_path.exists());
    }

    #[tokio::test]
    async fn test_remove_dir_all_nonexistent() {
        let nonexistent_path = "/nonexistent/path/that/should/not/exist";

        // Should not panic on nonexistent directory
        let result = remove_dir_all(nonexistent_path).await;
        assert!(result.is_ok());
    }

    #[tokio::test]
    async fn test_copy_file_with_progress() {
        let temp_dir = TempDir::new().unwrap();
        let source_path = temp_dir.path().join("source.txt");
        let dest_path = temp_dir.path().join("dest.txt");

        // Create source file with content
        let source_content = "Hello, World! This is test content.";
        let mut source_file = fs::File::create(&source_path).await.unwrap();
        source_file.write_all(source_content.as_bytes()).await.unwrap();
        source_file.flush().await.unwrap();

        // Track progress
        let mut progress_calls = Vec::new();
        let progress_callback = Box::new(|copied: u64, total: u64| {
            progress_calls.push((copied, total));
        });

        // Copy file with progress
        let copied_bytes = copy_file_with_progress(&source_path, &dest_path, Some(progress_callback)).await.unwrap();

        // Verify file was copied
        assert!(dest_path.exists());
        let dest_content = fs::read_to_string(&dest_path).await.unwrap();
        assert_eq!(dest_content, source_content);

        // Verify progress was reported
        assert!(!progress_calls.is_empty());
        assert_eq!(progress_calls.last().unwrap().0, source_content.len() as u64);
        assert_eq!(copied_bytes, source_content.len() as u64);
    }

    #[tokio::test]
    async fn test_copy_file_with_progress_no_callback() {
        let temp_dir = TempDir::new().unwrap();
        let source_path = temp_dir.path().join("source.txt");
        let dest_path = temp_dir.path().join("dest.txt");

        // Create source file
        let source_content = "Test content";
        fs::write(&source_path, source_content).await.unwrap();

        // Copy file without progress callback
        let copied_bytes = copy_file_with_progress(&source_path, &dest_path, None).await.unwrap();

        // Verify file was copied
        assert!(dest_path.exists());
        assert_eq!(copied_bytes, source_content.len() as u64);
    }

    #[tokio::test]
    async fn test_get_dir_size() {
        let temp_dir = TempDir::new().unwrap();
        let test_dir = temp_dir.path().join("test_dir");

        // Create directory structure
        fs::create_dir_all(&test_dir).await.unwrap();

        // Create files with known sizes
        let content1 = "Content 1";
        let content2 = "Content 2 with more text";
        let content3 = "Content 3";

        fs::write(test_dir.join("file1.txt"), content1).await.unwrap();
        fs::write(test_dir.join("file2.txt"), content2).await.unwrap();

        // Create subdirectory with file
        let sub_dir = test_dir.join("subdir");
        fs::create_dir_all(&sub_dir).await.unwrap();
        fs::write(sub_dir.join("file3.txt"), content3).await.unwrap();

        // Calculate directory size
        let size = get_dir_size(&test_dir).await.unwrap();

        let expected_size = content1.len() + content2.len() + content3.len();
        assert_eq!(size, expected_size as u64);
    }

    #[tokio::test]
    async fn test_get_dir_size_empty() {
        let temp_dir = TempDir::new().unwrap();
        let empty_dir = temp_dir.path().join("empty");

        // Create empty directory
        fs::create_dir_all(&empty_dir).await.unwrap();

        // Size should be 0
        let size = get_dir_size(&empty_dir).await.unwrap();
        assert_eq!(size, 0);
    }

    #[tokio::test]
    async fn test_get_dir_size_nonexistent() {
        let nonexistent_path = "/nonexistent/directory";

        // Should return error for nonexistent directory
        let result = get_dir_size(nonexistent_path).await;
        assert!(result.is_err());
    }

    #[tokio::test]
    async fn test_ensure_dir() {
        let temp_dir = TempDir::new().unwrap();
        let new_dir = temp_dir.path().join("new_dir").join("nested");

        // Directory should not exist initially
        assert!(!new_dir.exists());

        // Ensure directory exists
        ensure_dir(&new_dir).await.unwrap();

        // Directory should now exist
        assert!(new_dir.exists());
    }

    #[tokio::test]
    async fn test_ensure_dir_already_exists() {
        let temp_dir = TempDir::new().unwrap();
        let existing_dir = temp_dir.path().join("existing");

        // Create directory
        fs::create_dir_all(&existing_dir).await.unwrap();
        assert!(existing_dir.exists());

        // Ensure directory exists (should not error)
        ensure_dir(&existing_dir).await.unwrap();
        assert!(existing_dir.exists());
    }

    #[tokio::test]
    fn test_path_exists() {
        let temp_dir = TempDir::new().unwrap();
        let existing_path = temp_dir.path();
        let nonexistent_path = Path::new("/nonexistent/path");

        // Test existing path
        assert!(path_exists(existing_path));

        // Test nonexistent path
        assert!(!path_exists(nonexistent_path));
    }

    #[tokio::test]
    fn test_get_file_extension() {
        let test_cases = vec![
            ("file.txt", Some("txt")),
            ("document.pdf", Some("pdf")),
            ("archive.tar.gz", Some("gz")),
            ("no_extension", None),
            (".hidden", None),
            ("multiple.dots.tar.gz", Some("gz")),
            ("UPPERCASE.TXT", Some("txt")), // Should be lowercase
            ("", None),
        ];

        for (filename, expected) in test_cases {
            let result = get_file_extension(filename);
            assert_eq!(result, expected.map(|s| s.to_string()),
                      "Failed for filename: {}", filename);
        }
    }

    #[tokio::test]
    fn test_is_valid_filename() {
        let valid_names = vec![
            "normal_file.txt",
            "file-with-dashes",
            "file_with_underscores",
            "File With Spaces.txt",
            "123456789",
            "a",
            "CamelCase.txt",
            "UPPERCASE.TXT",
        ];

        let invalid_names = vec![
            "", // Empty
            "..", // Directory traversal
            "file/with/slashes",
            "file\\with\\backslashes",
            "file:with:colons",
            "file*with*asterisks",
            "file?with?questions",
            "file\"with\"quotes",
            "file<with>brackets",
            "file|with|pipes",
            "con", // Windows reserved name
            "prn", // Windows reserved name
            "aux", // Windows reserved name
            "nul", // Windows reserved name
            "file\twith\ttabs",
            "file\nwith\nnewlines",
        ];

        for name in valid_names {
            assert!(is_valid_filename(name), "Should be valid: {}", name);
        }

        for name in invalid_names {
            assert!(!is_valid_filename(name), "Should be invalid: {}", name);
        }
    }

    #[tokio::test]
    async fn test_copy_file_creates_parent_directories() {
        let temp_dir = TempDir::new().unwrap();
        let source_path = temp_dir.path().join("source.txt");
        let dest_path = temp_dir.path().join("nested").join("dest").join("file.txt");

        // Create source file
        let content = "Test content";
        fs::write(&source_path, content).await.unwrap();

        // Copy file (should create parent directories)
        copy_file_with_progress(&source_path, &dest_path, None).await.unwrap();

        // Verify file was copied and parent directories were created
        assert!(dest_path.exists());
        let dest_content = fs::read_to_string(&dest_path).await.unwrap();
        assert_eq!(dest_content, content);
    }
}