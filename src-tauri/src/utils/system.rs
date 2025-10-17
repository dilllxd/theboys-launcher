use crate::models::{SystemInfo, JavaVersion, LauncherResult, LauncherError};
use std::process::Command;
use std::path::Path;

/// Get total system memory in MB
pub async fn get_total_memory_mb() -> LauncherResult<u64> {
    let (total_mb, _) = get_memory_info()?;
    Ok(total_mb)
}

/// Get system information
pub fn get_system_info() -> LauncherResult<SystemInfo> {
    let os = std::env::consts::OS.to_string();
    let arch = std::env::consts::ARCH.to_string();

    // Get memory information (platform specific)
    let (total_memory_mb, available_memory_mb) = get_memory_info()?;

    let cpu_cores = num_cpus::get();
    let java_versions = detect_java_installations()?;
    let java_installed = !java_versions.is_empty();

    Ok(SystemInfo {
        os,
        arch,
        total_memory_mb,
        available_memory_mb,
        cpu_cores,
        java_installed,
        java_versions,
    })
}

/// Detect Java installations on the system
pub fn detect_java_installations() -> LauncherResult<Vec<JavaVersion>> {
    let mut java_versions = Vec::new();

    // Common Java installation paths
    let common_paths = get_common_java_paths();

    for path in common_paths {
        if let Ok(java_exe) = find_java_executable(&path) {
            if let Ok(version) = get_java_version(&java_exe) {
                java_versions.push(version);
            }
        }
    }

    // Also check system PATH
    if let Ok(java_exe) = which::which("java") {
        if let Ok(version) = get_java_version(&java_exe.to_string_lossy()) {
            // Avoid duplicates
            if !java_versions.iter().any(|v| v.path == java_exe.to_string_lossy()) {
                java_versions.push(version);
            }
        }
    }

    // Sort by major version (descending)
    java_versions.sort_by(|a, b| b.major_version.cmp(&a.major_version));

    Ok(java_versions)
}

/// Get memory information for the current system
fn get_memory_info() -> LauncherResult<(u64, u64)> {
    #[cfg(target_os = "windows")]
    {
        use std::mem;
        use windows_sys::Win32::System::SystemInformation::{
            GlobalMemoryStatusEx, MEMORYSTATUSEX,
        };

        unsafe {
            let mut mem_info: MEMORYSTATUSEX = mem::zeroed();
            mem_info.dwLength = mem::size_of::<MEMORYSTATUSEX>() as u32;

            if GlobalMemoryStatusEx(&mut mem_info) != 0 {
                let total = mem_info.ullTotalPhys / (1024 * 1024); // Convert to MB
                let available = mem_info.ullAvailPhys / (1024 * 1024);
                return Ok((total, available));
            }
        }
    }

    #[cfg(target_os = "macos")]
    {
        use std::process::Command;

        if let Ok(output) = Command::new("sysctl")
            .args(&["-n", "hw.memsize"])
            .output()
        {
            let total_bytes: u64 = String::from_utf8_lossy(&output.stdout)
                .trim()
                .parse()
                .unwrap_or(0);

            let total_mb = total_bytes / (1024 * 1024);
            // For available memory on macOS, we'd need more complex logic
            // For now, estimate 80% of total as available
            let available_mb = (total_mb as f64 * 0.8) as u64;
            return Ok((total_mb, available_mb));
        }
    }

    #[cfg(target_os = "linux")]
    {
        if let Ok(meminfo) = fs::read_to_string("/proc/meminfo") {
            let mut total_kb = 0u64;
            let mut available_kb = 0u64;

            for line in meminfo.lines() {
                if line.starts_with("MemTotal:") {
                    total_kb = line.split_whitespace()
                        .nth(1)
                        .and_then(|s| s.parse().ok())
                        .unwrap_or(0);
                } else if line.starts_with("MemAvailable:") {
                    available_kb = line.split_whitespace()
                        .nth(1)
                        .and_then(|s| s.parse().ok())
                        .unwrap_or(0);
                }
            }

            if total_kb > 0 {
                return Ok((total_kb / 1024, available_kb / 1024));
            }
        }
    }

    // Fallback values
    Ok((8192, 4096)) // Assume 8GB total, 4GB available
}

/// Get common Java installation paths for the current platform
fn get_common_java_paths() -> Vec<String> {
    match std::env::consts::OS {
        "windows" => vec![
            r"C:\Program Files\Java".to_string(),
            r"C:\Program Files (x86)\Java".to_string(),
            r"C:\Program Files\Eclipse Adoptium".to_string(),
            r"C:\Program Files (x86)\Eclipse Adoptium".to_string(),
        ],
        "macos" => vec![
            "/Library/Java/JavaVirtualMachines".to_string(),
            "/System/Library/Java/JavaVirtualMachines".to_string(),
            "/usr/local/Cellar/openjdk".to_string(),
        ],
        "linux" => vec![
            "/usr/lib/jvm".to_string(),
            "/usr/lib64/jvm".to_string(),
            "/usr/local/lib/jvm".to_string(),
            "/opt/java".to_string(),
        ],
        _ => Vec::new(),
    }
}

/// Find Java executable in a given path
fn find_java_executable(base_path: &str) -> LauncherResult<String> {
    let java_exe = if cfg!(target_os = "windows") {
        "java.exe"
    } else {
        "java"
    };

    let path = Path::new(base_path);

    // Look for bin/java.exe or bin/java in subdirectories
    for entry in path.read_dir().unwrap_or_else(|_| {
        // Create empty iterator if read fails
        std::fs::read_dir(".").unwrap()
    }) {
        if let Ok(entry) = entry {
            let entry_path = entry.path();
            if entry_path.is_dir() {
                let java_path = entry_path.join("bin").join(java_exe);
                if java_path.exists() {
                    return Ok(java_path.to_string_lossy().to_string());
                }
            }
        }
    }

    Err(LauncherError::JavaNotFound)
}

/// Get Java version information
fn get_java_version(java_path: &str) -> LauncherResult<JavaVersion> {
    let output = Command::new(java_path)
        .arg("-version")
        .output();

    let output = match output {
        Ok(output) => {
            // Java writes version info to stderr
            String::from_utf8_lossy(&output.stderr).to_string()
        }
        Err(_) => return Err(LauncherError::JavaNotFound),
    };

    // Parse version string
    // Java 17+ output: openjdk version "17.0.2" 2022-01-18
    // Java 8 output: java version "1.8.0_322"
    let version_regex = regex::Regex::new(r#"version "(\d+)(?:\.(\d+))?"#).unwrap();

    let (version, major_version, is_64bit) = if let Some(captures) = version_regex.captures(&output) {
        let major = captures.get(1).unwrap().as_str().parse::<u32>().unwrap_or(8);
        let minor = captures.get(2).map(|m| m.as_str()).unwrap_or("0");

        let version_str = if major >= 9 {
            format!("{}.{}", major, minor)
        } else {
            // Java 8 and below use 1.x format
            format!("1.{}", major)
        };

        // Check if 64-bit
        let is_64bit = output.contains("64-Bit") || output.contains("64bit");

        (version_str, major, is_64bit)
    } else {
        // Fallback parsing
        let version = "Unknown".to_string();
        let major_version = 8;
        let is_64bit = output.contains("64-Bit") || output.contains("64bit");
        (version, major_version, is_64bit)
    };

    Ok(JavaVersion {
        version,
        path: java_path.to_string(),
        is_64bit,
        major_version,
    })
}