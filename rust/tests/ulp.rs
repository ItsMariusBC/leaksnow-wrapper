use leaksnow::{Client, UlpDownloadRequest, UlpSearchRequest, UlpType};
use wiremock::matchers::{body_json, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

#[tokio::test]
async fn search_returns_token() {
    let server = MockServer::start().await;
    Mock::given(method("POST")).and(path("/api/v1/ulp/search"))
        .and(body_json(serde_json::json!({"type": "domain", "value": "example.com"})))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"downloadToken": "tok123"})))
        .expect(1).mount(&server).await;

    let c = Client::builder("ms_key").base_url(server.uri()).build();
    let v = c.ulp().search(UlpSearchRequest { kind: UlpType::Domain, value: "example.com".into() }).await.unwrap();
    assert_eq!(v["downloadToken"], "tok123");
}

#[tokio::test]
async fn download_posts_token() {
    let server = MockServer::start().await;
    Mock::given(method("POST")).and(path("/api/v1/ulp/download"))
        .and(body_json(serde_json::json!({"token": "tok123"})))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"lines": []})))
        .expect(1).mount(&server).await;

    let c = Client::builder("ms_key").base_url(server.uri()).build();
    c.ulp().download(UlpDownloadRequest { token: "tok123".into() }).await.unwrap();
}
