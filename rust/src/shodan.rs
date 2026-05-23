use reqwest::Method;
use serde_json::Value;

use crate::types::ShodanCustomScanRequest;
use crate::Error;

/// Shodan endpoints. Obtain via [`crate::Client::shodan`].
pub struct Shodan<'a> {
    pub(crate) client: &'a crate::Client,
}

impl Shodan<'_> {
    /// Lists your custom Shodan scans. Cost: 0.
    pub async fn custom_scans(&self) -> Result<Value, Error> {
        self.client
            .request_value(Method::GET, "/api/v1/shodan/custom-scans", None)
            .await
    }

    /// Launches a Shodan scan on the target. Cost: 1.
    pub async fn custom_scan(&self, req: ShodanCustomScanRequest) -> Result<Value, Error> {
        self.client
            .request_value(
                Method::POST,
                "/api/v1/shodan/custom-scan",
                Some(serde_json::to_value(req)?),
            )
            .await
    }

    /// Returns the status of a scan by id. Cost: 0.
    pub async fn get_scan(&self, id: &str) -> Result<Value, Error> {
        let path = format!("/api/v1/shodan/custom-scans/{}", urlencode(id));
        self.client.request_value(Method::GET, &path, None).await
    }

    /// Removes a scan entry by id. Cost: 0.
    pub async fn delete_scan(&self, id: &str) -> Result<Value, Error> {
        let path = format!("/api/v1/shodan/custom-scans/{}", urlencode(id));
        self.client.request_value(Method::DELETE, &path, None).await
    }

    /// Aggregated Shodan view for an IPv4. Cost: 0.
    pub async fn host(&self, ip: &str) -> Result<Value, Error> {
        let path = format!("/api/v1/shodan/host/{}", urlencode(ip));
        self.client.request_value(Method::GET, &path, None).await
    }
}

/// Minimal percent-encoding for a single path segment (RFC 3986 unreserved kept).
pub(crate) fn urlencode(seg: &str) -> String {
    let mut out = String::with_capacity(seg.len());
    for b in seg.bytes() {
        match b {
            b'A'..=b'Z' | b'a'..=b'z' | b'0'..=b'9' | b'-' | b'.' | b'_' | b'~' => {
                out.push(b as char)
            }
            _ => out.push_str(&format!("%{:02X}", b)),
        }
    }
    out
}
