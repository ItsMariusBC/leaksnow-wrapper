use std::time::Duration;

use leaksnow::{Client, RetryConfig, SearchRequest};
use wiremock::matchers::{body_json, header, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

#[tokio::test]
async fn sends_bearer_and_parses() {
    let server = MockServer::start().await;
    Mock::given(method("POST"))
        .and(path("/api/v1/search"))
        .and(header("authorization", "Bearer ms_key"))
        .and(body_json(serde_json::json!({"query": "x"})))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"ok": true})))
        .mount(&server)
        .await;

    let client = Client::builder("ms_key").base_url(server.uri()).build();
    let v = client
        .search(SearchRequest {
            query: "x".into(),
            ..Default::default()
        })
        .await
        .unwrap();
    assert_eq!(v, serde_json::json!({"ok": true}));
}

#[tokio::test]
async fn maps_auth_error() {
    let server = MockServer::start().await;
    Mock::given(method("POST"))
        .and(path("/api/v1/search"))
        .respond_with(
            ResponseTemplate::new(401).set_body_json(serde_json::json!({"error": "bad key"})),
        )
        .mount(&server)
        .await;

    let client = Client::builder("ms_key").base_url(server.uri()).build();
    let err = client.search(SearchRequest::default()).await.unwrap_err();
    assert!(matches!(err, leaksnow::Error::Auth { .. }));
}

#[tokio::test]
async fn retries_then_succeeds() {
    let server = MockServer::start().await;
    Mock::given(method("POST"))
        .and(path("/api/v1/search"))
        .respond_with(ResponseTemplate::new(429).insert_header("retry-after", "0"))
        .up_to_n_times(1)
        .mount(&server)
        .await;
    Mock::given(method("POST"))
        .and(path("/api/v1/search"))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"ok": 1})))
        .mount(&server)
        .await;

    let client = Client::builder("ms_key")
        .base_url(server.uri())
        .retry(RetryConfig {
            max_retries: 2,
            base_delay: Duration::from_millis(1),
            max_delay: Duration::from_millis(5),
            retry_on: vec![429],
        })
        .build();
    let v = client.search(SearchRequest::default()).await.unwrap();
    assert_eq!(v, serde_json::json!({"ok": 1}));
}

#[tokio::test]
async fn quota_error_after_exhaustion() {
    let server = MockServer::start().await;
    Mock::given(method("POST"))
        .and(path("/api/v1/search"))
        .respond_with(ResponseTemplate::new(429).insert_header("retry-after", "0"))
        .mount(&server)
        .await;

    let client = Client::builder("ms_key")
        .base_url(server.uri())
        .retry(RetryConfig {
            max_retries: 1,
            base_delay: Duration::from_millis(1),
            max_delay: Duration::from_millis(5),
            retry_on: vec![429],
        })
        .build();
    let err = client.search(SearchRequest::default()).await.unwrap_err();
    assert!(matches!(err, leaksnow::Error::Quota { .. }));
}
