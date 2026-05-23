use std::time::Duration;

use thiserror::Error;

/// Errors returned by the leaks.now client.
#[derive(Debug, Error)]
pub enum Error {
    /// 401 / 403.
    #[error("leaksnow: authentication failed ({status})")]
    Auth { status: u16, body: String },
    /// 429. `retry_after` is set when the server sent a numeric Retry-After.
    #[error("leaksnow: quota exceeded (429)")]
    Quota {
        retry_after: Option<Duration>,
        body: String,
    },
    /// 400 / 422.
    #[error("leaksnow: validation error ({status})")]
    Validation { status: u16, body: String },
    /// 5xx.
    #[error("leaksnow: server error ({status})")]
    Server { status: u16, body: String },
    /// Any other non-success status.
    #[error("leaksnow: request failed ({status})")]
    Api { status: u16, body: String },
    /// Network / timeout / connection failure.
    #[error("leaksnow: transport error: {0}")]
    Transport(#[from] reqwest::Error),
    /// Response or request body (de)serialization failure.
    #[error("leaksnow: decode error: {0}")]
    Decode(#[from] serde_json::Error),
}

/// Maps an HTTP status + body to the matching [`Error`] variant.
pub(crate) fn error_from_response(
    status: u16,
    body: String,
    retry_after: Option<Duration>,
) -> Error {
    match status {
        401 | 403 => Error::Auth { status, body },
        429 => Error::Quota { retry_after, body },
        400 | 422 => Error::Validation { status, body },
        s if s >= 500 => Error::Server { status, body },
        _ => Error::Api { status, body },
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn maps_status_to_variant() {
        assert!(matches!(
            error_from_response(401, String::new(), None),
            Error::Auth { .. }
        ));
        assert!(matches!(
            error_from_response(403, String::new(), None),
            Error::Auth { .. }
        ));
        assert!(matches!(
            error_from_response(422, String::new(), None),
            Error::Validation { .. }
        ));
        assert!(matches!(
            error_from_response(503, String::new(), None),
            Error::Server { .. }
        ));
        assert!(matches!(
            error_from_response(418, String::new(), None),
            Error::Api { status: 418, .. }
        ));
    }

    #[test]
    fn quota_keeps_retry_after() {
        let e = error_from_response(429, "x".into(), Some(Duration::from_secs(2)));
        match e {
            Error::Quota { retry_after, body } => {
                assert_eq!(retry_after, Some(Duration::from_secs(2)));
                assert_eq!(body, "x");
            }
            _ => panic!("expected Quota"),
        }
    }
}
