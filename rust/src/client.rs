use std::time::Duration;

use reqwest::header::{AUTHORIZATION, RETRY_AFTER};
use reqwest::{Method, Response};
use serde_json::Value;

use crate::error::error_from_response;
use crate::retry::{backoff_delay, should_retry, RetryConfig};
use crate::types::SearchRequest;
use crate::Error;

const DEFAULT_BASE_URL: &str = "https://leaks.now";

/// Async client for the leaks.now `/api/v1` API. Build with [`Client::new`] or
/// [`Client::builder`].
#[derive(Debug, Clone)]
pub struct Client {
    api_key: String,
    base_url: String,
    http: reqwest::Client,
    retry: RetryConfig,
}

/// Builder for [`Client`].
#[derive(Debug)]
pub struct ClientBuilder {
    api_key: String,
    base_url: String,
    timeout: Duration,
    retry: RetryConfig,
    http: Option<reqwest::Client>,
}

impl Client {
    /// Creates a client with default settings (30s timeout, retries disabled).
    pub fn new(api_key: impl Into<String>) -> Self {
        Self::builder(api_key).build()
    }

    /// Starts a builder for custom base URL, timeout, retry, or HTTP client.
    pub fn builder(api_key: impl Into<String>) -> ClientBuilder {
        ClientBuilder {
            api_key: api_key.into(),
            base_url: DEFAULT_BASE_URL.to_string(),
            timeout: Duration::from_secs(30),
            retry: RetryConfig::default(),
            http: None,
        }
    }

    /// Runs an OSINT search (leak/service/shodan). Cost: 1 credit.
    pub async fn search(&self, req: SearchRequest) -> Result<Value, Error> {
        self.request_value(
            Method::POST,
            "/api/v1/search",
            Some(serde_json::to_value(req)?),
        )
        .await
    }

    pub(crate) async fn request_value(
        &self,
        method: Method,
        path: &str,
        body: Option<Value>,
    ) -> Result<Value, Error> {
        let resp = self.execute(method, path, body).await?;
        let bytes = resp.bytes().await?;
        if bytes.is_empty() {
            return Ok(Value::Null);
        }
        Ok(serde_json::from_slice(&bytes)?)
    }

    pub(crate) async fn execute(
        &self,
        method: Method,
        path: &str,
        body: Option<Value>,
    ) -> Result<Response, Error> {
        let url = format!("{}{}", self.base_url, path);
        let mut attempt = 0u32;
        loop {
            let mut builder = self
                .http
                .request(method.clone(), &url)
                .header(AUTHORIZATION, format!("Bearer {}", self.api_key));
            if let Some(ref b) = body {
                builder = builder.json(b);
            }

            let resp = builder.send().await?;
            if resp.status().is_success() {
                return Ok(resp);
            }

            let status = resp.status().as_u16();
            let retry_after = parse_retry_after(&resp);
            if should_retry(status, attempt, &self.retry) {
                let delay = backoff_delay(attempt, &self.retry, retry_after, fastrand::f64());
                tokio::time::sleep(delay).await;
                attempt += 1;
                continue;
            }

            let text = resp.text().await.unwrap_or_default();
            return Err(error_from_response(status, text, retry_after));
        }
    }

    /// Shodan endpoints.
    pub fn shodan(&self) -> crate::shodan::Shodan<'_> {
        crate::shodan::Shodan { client: self }
    }

    /// ULP endpoints.
    pub fn ulp(&self) -> crate::ulp::Ulp<'_> {
        crate::ulp::Ulp { client: self }
    }

    /// Intelligence X endpoints.
    pub fn intelx(&self) -> crate::intelx::IntelX<'_> {
        crate::intelx::IntelX { client: self }
    }
}

impl ClientBuilder {
    /// Overrides the API base URL (default `https://leaks.now`).
    pub fn base_url(mut self, url: impl Into<String>) -> Self {
        let mut u = url.into();
        while u.ends_with('/') {
            u.pop();
        }
        self.base_url = u;
        self
    }

    /// Sets the per-request timeout (default 30s). Ignored if `http_client` is set.
    pub fn timeout(mut self, d: Duration) -> Self {
        self.timeout = d;
        self
    }

    /// Enables and configures automatic retries (disabled by default).
    pub fn retry(mut self, r: RetryConfig) -> Self {
        self.retry = r;
        self
    }

    /// Supplies a custom `reqwest::Client` (overrides `timeout`).
    pub fn http_client(mut self, client: reqwest::Client) -> Self {
        self.http = Some(client);
        self
    }

    /// Builds the [`Client`].
    pub fn build(self) -> Client {
        let http = self.http.unwrap_or_else(|| {
            reqwest::Client::builder()
                .timeout(self.timeout)
                .build()
                .expect("default reqwest client")
        });
        Client {
            api_key: self.api_key,
            base_url: self.base_url,
            http,
            retry: self.retry,
        }
    }
}

pub(crate) fn parse_retry_after(resp: &Response) -> Option<Duration> {
    let raw = resp.headers().get(RETRY_AFTER)?.to_str().ok()?;
    let secs: u64 = raw.trim().parse().ok()?;
    Some(Duration::from_secs(secs))
}
