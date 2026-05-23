use reqwest::Method;
use serde_json::Value;

use crate::types::{UlpDownloadRequest, UlpSearchRequest};
use crate::Error;

/// ULP (Leak Zero) endpoints. Obtain via [`crate::Client::ulp`].
pub struct Ulp<'a> {
    pub(crate) client: &'a crate::Client,
}

impl Ulp<'_> {
    /// Runs a ULP search; the response includes a downloadToken (~15 min). Cost: 1.
    pub async fn search(&self, req: UlpSearchRequest) -> Result<Value, Error> {
        self.client
            .request_value(Method::POST, "/api/v1/ulp/search", Some(serde_json::to_value(req)?))
            .await
    }

    /// Exports the JSON rows for a previously returned token. Cost: 10.
    pub async fn download(&self, req: UlpDownloadRequest) -> Result<Value, Error> {
        self.client
            .request_value(Method::POST, "/api/v1/ulp/download", Some(serde_json::to_value(req)?))
            .await
    }
}
