use std::time::{Duration, Instant};
use std::collections::HashMap;
use tokio::sync::RwLock;
use serde::{Serialize, Deserialize};
use tracing::{info, warn};

/// Performance metrics for monitoring
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PerformanceMetrics {
    pub startup_time_ms: u64,
    pub memory_usage_mb: u64,
    pub active_processes: usize,
    pub active_downloads: usize,
    pub cache_hit_rate: f64,
    pub average_response_time_ms: f64,
}

/// Simple in-memory cache for performance optimization
#[derive(Debug)]
pub struct Cache<K, V> {
    data: RwLock<HashMap<K, (V, Instant)>>,
    ttl: Duration,
}

impl<K: Clone + Eq + std::hash::Hash, V: Clone> Cache<K, V> {
    pub fn new(ttl: Duration) -> Self {
        Self {
            data: RwLock::new(HashMap::new()),
            ttl,
        }
    }

    pub async fn get(&self, key: &K) -> Option<V> {
        let data = self.data.read().await;
        if let Some((value, timestamp)) = data.get(key) {
            if timestamp.elapsed() < self.ttl {
                return Some(value.clone());
            }
        }
        None
    }

    pub async fn put(&self, key: K, value: V) {
        let mut data = self.data.write().await;
        data.insert(key, (value, Instant::now()));

        // Clean up expired entries periodically
        if data.len() > 1000 {
            data.retain(|_, (_, timestamp)| timestamp.elapsed() < self.ttl);
        }
    }

    pub async fn clear(&self) {
        let mut data = self.data.write().await;
        data.clear();
    }
}

/// Connection pool for HTTP requests
pub struct HttpPool {
    client: reqwest::Client,
}

impl HttpPool {
    pub fn new() -> Self {
        let client = reqwest::Client::builder()
            .timeout(Duration::from_secs(30))
            .pool_max_idle_per_host(10)
            .pool_idle_timeout(Duration::from_secs(90))
            .user_agent("TheBoys-Launcher/1.1.0")
            .build()
            .expect("Failed to create HTTP client");

        Self { client }
    }

    pub fn client(&self) -> &reqwest::Client {
        &self.client
    }
}

/// Global performance monitoring
pub struct PerformanceMonitor {
    startup_time: Instant,
    cache: Cache<String, f64>,
}

impl PerformanceMonitor {
    pub fn new() -> Self {
        Self {
            startup_time: Instant::now(),
            cache: Cache::new(Duration::from_secs(300)), // 5 minutes cache
        }
    }

    pub async fn record_operation_time(&self, operation: &str, duration_ms: f64) {
        self.cache.put(operation.to_string(), duration_ms).await;
    }

    pub async fn get_average_response_time(&self, operation: &str) -> Option<f64> {
        self.cache.get(&operation.to_string()).await
    }

    pub fn get_startup_time_ms(&self) -> u64 {
        self.startup_time.elapsed().as_millis() as u64
    }

    pub async fn get_memory_usage_mb(&self) -> u64 {
        // Get memory usage using platform-specific methods
        if cfg!(target_os = "windows") {
            self.get_windows_memory_usage().await
        } else if cfg!(target_os = "linux") {
            self.get_linux_memory_usage().await
        } else if cfg!(target_os = "macos") {
            self.get_macos_memory_usage().await
        } else {
            0
        }
    }

    async fn get_windows_memory_usage(&self) -> u64 {
        use std::process::Command;

        match Command::new("wmic")
            .args(&["process", "where", "processid=$PID", "get", "WorkingSetSize"])
            .output()
        {
            Ok(output) => {
                let output_str = String::from_utf8_lossy(&output.stdout);
                if let Some(line) = output_str.lines().nth(1) {
                    if let Ok(bytes) = line.trim().parse::<u64>() {
                        return bytes / 1024 / 1024; // Convert to MB
                    }
                }
            }
            Err(e) => warn!("Failed to get memory usage: {}", e),
        }
        0
    }

    async fn get_linux_memory_usage(&self) -> u64 {
        use std::fs;

        if let Ok(status) = fs::read_to_string("/proc/self/status") {
            for line in status.lines() {
                if line.starts_with("VmRSS:") {
                    if let Some(kb_str) = line.split_whitespace().nth(1) {
                        if let Ok(kb) = kb_str.parse::<u64>() {
                            return kb / 1024; // Convert KB to MB
                        }
                    }
                }
            }
        }
        0
    }

    async fn get_macos_memory_usage(&self) -> u64 {
        use std::process::Command;

        match Command::new("ps")
            .args(&["-o", "rss=", "-p", &std::process::id().to_string()])
            .output()
        {
            Ok(output) => {
                let output_str = String::from_utf8_lossy(&output.stdout);
                if let Ok(kb) = output_str.trim().parse::<u64>() {
                    return kb / 1024; // Convert KB to MB
                }
            }
            Err(e) => warn!("Failed to get memory usage: {}", e),
        }
        0
    }
}

/// Lazy static for global performance monitoring
use std::sync::OnceLock;
static PERFORMANCE_MONITOR: OnceLock<PerformanceMonitor> = OnceLock::new();

pub fn get_performance_monitor() -> &'static PerformanceMonitor {
    PERFORMANCE_MONITOR.get_or_init(|| {
        info!("Initializing performance monitor");
        PerformanceMonitor::new()
    })
}

/// Utility for measuring operation performance
pub struct PerformanceTimer {
    operation: String,
    start: Instant,
}

impl PerformanceTimer {
    pub fn new(operation: impl Into<String>) -> Self {
        Self {
            operation: operation.into(),
            start: Instant::now(),
        }
    }
}

impl Drop for PerformanceTimer {
    fn drop(&mut self) {
        let duration = self.start.elapsed().as_millis() as f64;
        let monitor = get_performance_monitor();

        // Record asynchronously to avoid blocking the drop
        let operation = self.operation.clone();
        tokio::spawn(async move {
            monitor.record_operation_time(&operation, duration).await;
        });

        if duration > 1000.0 { // Log slow operations
            warn!("Slow operation detected: {} took {}ms", self.operation, duration);
        }
    }
}

/// Macro for easy performance measurement
#[macro_export]
macro_rules! timed {
    ($operation:expr) => {{
        let _timer = $crate::utils::performance::PerformanceTimer::new(stringify!($operation));
        $operation
    }};
}