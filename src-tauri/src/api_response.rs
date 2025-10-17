use serde::{Deserialize, Serialize};
use crate::models::LauncherError;

/// Standard API response structure that matches frontend expectations
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApiResponse<T> {
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
}

impl<T> ApiResponse<T> {
    pub fn success(data: T) -> Self {
        Self {
            success: true,
            data: Some(data),
            error: None,
        }
    }

    pub fn error(error: String) -> Self {
        Self {
            success: false,
            data: None,
            error: Some(error),
        }
    }
}

/// Convert LauncherResult to ApiResponse
impl<T: Serialize> From<Result<T, LauncherError>> for ApiResponse<T> {
    fn from(result: Result<T, LauncherError>) -> Self {
        match result {
            Ok(data) => ApiResponse::success(data),
            Err(error) => ApiResponse::error(error.to_string()),
        }
    }
}

/// Common response types for API calls
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HealthResponse {
    pub status: String,
    pub timestamp: String,
    pub version: String,
}