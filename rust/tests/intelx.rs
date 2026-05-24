use leaksnow::{Client, IntelXBucket, IntelXDownloadRequest};
use wiremock::matchers::{body_json, method, path};
use wiremock::{Mock, MockServer, ResponseTemplate};

#[tokio::test]
async fn downloads_and_download_and_delete() {
    let server = MockServer::start().await;
    Mock::given(method("GET"))
        .and(path("/api/v1/intelx/downloads"))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"items": []})))
        .mount(&server)
        .await;
    Mock::given(method("POST"))
        .and(path("/api/v1/intelx/download"))
        .and(body_json(
            serde_json::json!({"systemId": "uuid-1", "bucket": "leaks.private.general"}),
        ))
        .respond_with(ResponseTemplate::new(200).set_body_json(serde_json::json!({"id": "1"})))
        .mount(&server)
        .await;
    Mock::given(method("DELETE"))
        .and(path("/api/v1/intelx/downloads/7"))
        .respond_with(ResponseTemplate::new(200))
        .mount(&server)
        .await;

    let c = Client::builder("ms_key").base_url(server.uri()).build();
    c.intelx().downloads().await.unwrap();
    c.intelx()
        .download(IntelXDownloadRequest {
            system_id: "uuid-1".into(),
            bucket: IntelXBucket::LeaksPrivateGeneral,
        })
        .await
        .unwrap();
    c.intelx().delete_download("7").await.unwrap();
}

#[tokio::test]
async fn get_file_returns_binary() {
    let server = MockServer::start().await;
    Mock::given(method("GET"))
        .and(path("/api/v1/intelx/downloads/7/file"))
        .respond_with(
            ResponseTemplate::new(200)
                .insert_header("content-type", "application/octet-stream")
                .insert_header("content-disposition", "attachment; filename=\"dump.bin\"")
                .set_body_bytes(vec![1u8, 2, 3, 4]),
        )
        .mount(&server)
        .await;

    let c = Client::builder("ms_key").base_url(server.uri()).build();
    let f = c.intelx().get_file("7").await.unwrap();
    assert_eq!(f.data, vec![1u8, 2, 3, 4]);
    assert_eq!(f.content_type, "application/octet-stream");
    assert_eq!(f.filename.as_deref(), Some("dump.bin"));
}
