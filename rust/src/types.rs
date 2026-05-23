use serde::Serialize;

/// Search backend selector.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Default)]
#[serde(rename_all = "lowercase")]
pub enum Scope {
    #[default]
    Leak,
    Service,
    Shodan,
}

/// Result severity filter.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum Severity {
    Critical,
    High,
    Medium,
    Low,
    Info,
    All,
}

/// Field searched by the ULP endpoint.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "lowercase")]
pub enum UlpType {
    Domain,
    Email,
    Username,
    Password,
}

/// Intelligence X data bucket.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
pub enum IntelXBucket {
    #[serde(rename = "leaks.private.general")]
    LeaksPrivateGeneral,
    #[serde(rename = "leaks.logs")]
    LeaksLogs,
    #[serde(rename = "whois")]
    Whois,
    #[serde(rename = "dns")]
    Dns,
    #[serde(rename = "darknet.tor")]
    DarknetTor,
}

/// Body for [`crate::Client::search`].
#[derive(Debug, Clone, Default, Serialize)]
pub struct SearchRequest {
    pub query: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub scope: Option<Scope>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub plugin: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub severity: Option<Severity>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub page: Option<u32>,
}

/// Body for the Shodan custom-scan endpoint.
#[derive(Debug, Clone, Serialize)]
pub struct ShodanCustomScanRequest {
    pub target: String,
}

/// Body for the ULP search endpoint.
#[derive(Debug, Clone, Serialize)]
pub struct UlpSearchRequest {
    #[serde(rename = "type")]
    pub kind: UlpType,
    pub value: String,
}

/// Body for the ULP download endpoint.
#[derive(Debug, Clone, Serialize)]
pub struct UlpDownloadRequest {
    pub token: String,
}

/// Body for the IntelX download endpoint.
#[derive(Debug, Clone, Serialize)]
pub struct IntelXDownloadRequest {
    #[serde(rename = "systemId")]
    pub system_id: String,
    pub bucket: IntelXBucket,
}

/// A downloaded binary file (IntelX `get_file`).
#[derive(Debug, Clone)]
pub struct BinaryFile {
    pub data: Vec<u8>,
    pub content_type: String,
    pub filename: Option<String>,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn enum_serialization() {
        assert_eq!(
            serde_json::to_string(&Scope::Shodan).unwrap(),
            r#""shodan""#
        );
        assert_eq!(serde_json::to_string(&Severity::All).unwrap(), r#""all""#);
        assert_eq!(
            serde_json::to_string(&UlpType::Domain).unwrap(),
            r#""domain""#
        );
        assert_eq!(
            serde_json::to_string(&IntelXBucket::LeaksPrivateGeneral).unwrap(),
            r#""leaks.private.general""#
        );
        assert_eq!(
            serde_json::to_string(&IntelXBucket::DarknetTor).unwrap(),
            r#""darknet.tor""#
        );
    }

    #[test]
    fn search_request_skips_none() {
        let b = serde_json::to_string(&SearchRequest {
            query: "x".into(),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(b, r#"{"query":"x"}"#);
        let b = serde_json::to_string(&SearchRequest {
            query: "x".into(),
            scope: Some(Scope::Leak),
            severity: Some(Severity::All),
            ..Default::default()
        })
        .unwrap();
        assert_eq!(b, r#"{"query":"x","scope":"leak","severity":"all"}"#);
    }

    #[test]
    fn ulp_and_intelx_request_json() {
        let b = serde_json::to_string(&UlpSearchRequest {
            kind: UlpType::Domain,
            value: "example.com".into(),
        })
        .unwrap();
        assert_eq!(b, r#"{"type":"domain","value":"example.com"}"#);
        let b = serde_json::to_string(&IntelXDownloadRequest {
            system_id: "abc".into(),
            bucket: IntelXBucket::LeaksLogs,
        })
        .unwrap();
        assert_eq!(b, r#"{"systemId":"abc","bucket":"leaks.logs"}"#);
    }
}
