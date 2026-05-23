//! Minimal runnable example.
//!
//! Run with:
//! ```bash
//! LEAKSNOW_API_KEY=ms_xxx cargo run --example quickstart
//! ```

use std::time::Duration;

use leaksnow::{Client, RetryConfig, Scope, SearchRequest};

#[tokio::main]
async fn main() -> Result<(), leaksnow::Error> {
    let api_key =
        std::env::var("LEAKSNOW_API_KEY").expect("set LEAKSNOW_API_KEY in the environment");

    let client = Client::builder(api_key)
        .timeout(Duration::from_secs(15))
        .retry(RetryConfig {
            max_retries: 3,
            base_delay: Duration::from_millis(500),
            max_delay: Duration::from_secs(8),
            retry_on: vec![429, 500, 502, 503, 504],
        })
        .build();

    let results = client
        .search(SearchRequest {
            query: "host:example.com".into(),
            scope: Some(Scope::Leak),
            ..Default::default()
        })
        .await?;

    println!("{}", serde_json::to_string_pretty(&results).unwrap_or_default());
    Ok(())
}
