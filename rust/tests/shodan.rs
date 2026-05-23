use leaksnow::{Client, ShodanCustomScanRequest};
use wiremock::matchers::{method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

#[tokio::test]
async fn custom_scans_get() {
    let server = MockServer::start().await;
    Mock::given(method("GET")).and(path("/api/v1/shodan/custom-scans"))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!([])))
        .expect(1).mount(&server).await;
    let c = Client::builder("ms_key").base_url(server.uri()).build();
    c.shodan().custom_scans().await.unwrap();
}

#[tokio::test]
async fn custom_scan_post() {
    let server = MockServer::start().await;
    Mock::given(method("POST")).and(path("/api/v1/shodan/custom-scan"))
        .and(wiremock::matchers::body_json(serde_json::json!({"target": "scanme.nmap.org"})))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"id": 1})))
        .expect(1).mount(&server).await;
    let c = Client::builder("ms_key").base_url(server.uri()).build();
    c.shodan().custom_scan(ShodanCustomScanRequest { target: "scanme.nmap.org".into() }).await.unwrap();
}

#[tokio::test]
async fn get_delete_scan_and_host() {
    let server = MockServer::start().await;
    Mock::given(method("GET")).and(path("/api/v1/shodan/custom-scans/42"))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({}))).mount(&server).await;
    Mock::given(method("DELETE")).and(path("/api/v1/shodan/custom-scans/42"))
        .respond_with(ResponseTemplate::new(200)).mount(&server).await;
    Mock::given(method("GET")).and(path("/api/v1/shodan/host/1.2.3.4"))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({}))).mount(&server).await;

    let c = Client::builder("ms_key").base_url(server.uri()).build();
    c.shodan().get_scan("42").await.unwrap();
    c.shodan().delete_scan("42").await.unwrap();
    c.shodan().host("1.2.3.4").await.unwrap();
}
