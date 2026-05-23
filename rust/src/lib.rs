//! Async Rust client for the leaks.now `/api/v1` OSINT API
//! (Leak, Service, Shodan, ULP, IntelX).
//!
//! ```no_run
//! # async fn run() -> Result<(), leaksnow::Error> {
//! let client = leaksnow::Client::new("ms_xxx");
//! let results = client.search(leaksnow::SearchRequest {
//!     query: "host:example.com".into(),
//!     ..Default::default()
//! }).await?;
//! # Ok(()) }
//! ```
//!
//! Authentication uses an `Authorization: Bearer ms_...` header. Each successful
//! call consumes the credits documented in `docs/openapi.yaml` at the repo root.

mod error;
pub use error::Error;

mod retry;
pub use retry::RetryConfig;

mod types;
pub use types::{
    BinaryFile, IntelXBucket, IntelXDownloadRequest, Scope, SearchRequest, Severity,
    ShodanCustomScanRequest, UlpDownloadRequest, UlpSearchRequest, UlpType,
};
