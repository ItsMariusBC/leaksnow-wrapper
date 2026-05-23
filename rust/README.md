# leaksnow

Async Rust client for the [leaks.now](https://leaks.now) `/api/v1` OSINT API.

```toml
[dependencies]
leaksnow = "0.1"
tokio = { version = "1", features = ["macros", "rt-multi-thread"] }
```

```rust
use leaksnow::{Client, SearchRequest, Scope};

#[tokio::main]
async fn main() -> Result<(), leaksnow::Error> {
    let client = Client::new(std::env::var("LEAKSNOW_API_KEY").unwrap());

    let leaks = client.search(SearchRequest {
        query: "host:example.com".into(),
        scope: Some(Scope::Leak),
        ..Default::default()
    }).await?;

    let scan = client.shodan().custom_scan(
        leaksnow::ShodanCustomScanRequest { target: "scanme.nmap.org".into() }
    ).await?;

    let file = client.intelx().get_file("7").await?; // BinaryFile { data, content_type, filename }
    let _ = (leaks, scan, file);
    Ok(())
}
```

JSON responses are returned as `serde_json::Value` (the provider documents request
bodies only); deserialize into your own structs as needed. Errors are the
`leaksnow::Error` enum (`Auth`, `Quota`, `Validation`, `Server`, `Api`,
`Transport`, `Decode`). Retries are off by default; enable with
`Client::builder(key).retry(RetryConfig { .. }).build()`.

## License

MIT
