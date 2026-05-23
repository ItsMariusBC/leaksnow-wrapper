use reqwest::header::{CONTENT_DISPOSITION, CONTENT_TYPE};
use reqwest::Method;
use serde_json::Value;

use crate::shodan::urlencode;
use crate::types::{BinaryFile, IntelXDownloadRequest};
use crate::Error;

/// Intelligence X endpoints. Obtain via [`crate::Client::intelx`].
pub struct IntelX<'a> {
    pub(crate) client: &'a crate::Client,
}

impl IntelX<'_> {
    /// Returns your IntelX download history. Cost: 0.
    pub async fn downloads(&self) -> Result<Value, Error> {
        self.client
            .request_value(Method::GET, "/api/v1/intelx/downloads", None)
            .await
    }

    /// Requests an IntelX file (server-cached). Cost: 5.
    pub async fn download(&self, req: IntelXDownloadRequest) -> Result<Value, Error> {
        self.client
            .request_value(
                Method::POST,
                "/api/v1/intelx/download",
                Some(serde_json::to_value(req)?),
            )
            .await
    }

    /// Hides an entry in the IntelX history. Cost: 0.
    pub async fn delete_download(&self, id: &str) -> Result<Value, Error> {
        let path = format!("/api/v1/intelx/downloads/{}", urlencode(id));
        self.client.request_value(Method::DELETE, &path, None).await
    }

    /// Downloads the cached binary for a download id (no re-billing). Cost: 0.
    pub async fn get_file(&self, id: &str) -> Result<BinaryFile, Error> {
        let path = format!("/api/v1/intelx/downloads/{}/file", urlencode(id));
        let resp = self.client.execute(Method::GET, &path, None).await?;
        let content_type = resp
            .headers()
            .get(CONTENT_TYPE)
            .and_then(|v| v.to_str().ok())
            .unwrap_or("application/octet-stream")
            .to_string();
        let filename = resp
            .headers()
            .get(CONTENT_DISPOSITION)
            .and_then(|v| v.to_str().ok())
            .and_then(parse_filename);
        let data = resp.bytes().await?.to_vec();
        Ok(BinaryFile {
            data,
            content_type,
            filename,
        })
    }
}

/// Extracts a filename from a Content-Disposition header. The RFC 5987 extended
/// form (`filename*`) takes precedence and is percent-decoded.
pub(crate) fn parse_filename(disposition: &str) -> Option<String> {
    if let Some(rest) = find_ci(disposition, "filename*") {
        let after_eq = rest.trim_start_matches('=').trim();
        if let Some(idx) = after_eq.find('\'') {
            let tail = &after_eq[idx + 1..];
            if let Some(idx2) = tail.find('\'') {
                let value = &tail[idx2 + 1..];
                let value = value.split(';').next().unwrap_or(value).trim();
                return Some(percent_decode(value));
            }
        }
    }
    if let Some(rest) = find_ci(disposition, "filename") {
        let after_eq = rest.trim_start_matches('=').trim();
        let value = after_eq.split(';').next().unwrap_or(after_eq).trim();
        let value = value.trim_matches('"');
        if !value.is_empty() {
            return Some(value.to_string());
        }
    }
    None
}

fn find_ci<'a>(haystack: &'a str, token: &str) -> Option<&'a str> {
    let lower = haystack.to_ascii_lowercase();
    let idx = lower.find(token)?;
    let after = &haystack[idx + token.len()..];
    if token == "filename" && after.starts_with('*') {
        return None;
    }
    Some(after)
}

fn percent_decode(s: &str) -> String {
    let bytes = s.as_bytes();
    let mut out: Vec<u8> = Vec::with_capacity(bytes.len());
    let mut i = 0;
    while i < bytes.len() {
        if bytes[i] == b'%' && i + 2 < bytes.len() {
            let hi = (bytes[i + 1] as char).to_digit(16);
            let lo = (bytes[i + 2] as char).to_digit(16);
            if let (Some(h), Some(l)) = (hi, lo) {
                out.push((h * 16 + l) as u8);
                i += 3;
                continue;
            }
        }
        out.push(bytes[i]);
        i += 1;
    }
    String::from_utf8_lossy(&out).into_owned()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parses_filenames() {
        assert_eq!(
            parse_filename(r#"attachment; filename="dump.bin""#).as_deref(),
            Some("dump.bin")
        );
        assert_eq!(
            parse_filename("attachment; filename=plain.bin").as_deref(),
            Some("plain.bin")
        );
        assert_eq!(
            parse_filename("attachment; filename*=UTF-8''r%C3%A9sum%C3%A9.bin").as_deref(),
            Some("résumé.bin")
        );
        assert_eq!(
            parse_filename("attachment; filename*=UTF-8''y%20.bin").as_deref(),
            Some("y .bin")
        );
        assert_eq!(
            parse_filename(r#"attachment; filename="x.bin"; filename*=UTF-8''real%20name.bin"#)
                .as_deref(),
            Some("real name.bin")
        );
        assert_eq!(parse_filename(""), None);
    }
}
