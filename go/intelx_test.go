package leaksnow

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntelX(t *testing.T) {
	var path, body, method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		method = r.Method
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	ctx := context.Background()

	if _, err := c.IntelX.Downloads(ctx); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/intelx/downloads" || method != http.MethodGet {
		t.Fatalf("downloads: %s %s", method, path)
	}

	if _, err := c.IntelX.Download(ctx, IntelXDownloadRequest{SystemID: "uuid-1", Bucket: BucketLeaksPrivateGeneral}); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/intelx/download" {
		t.Fatalf("download path = %q", path)
	}
	if body != `{"systemId":"uuid-1","bucket":"leaks.private.general"}` {
		t.Fatalf("download body = %q", body)
	}

	if _, err := c.IntelX.DeleteDownload(ctx, "7"); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/intelx/downloads/7" || method != http.MethodDelete {
		t.Fatalf("deleteDownload: %s %s", method, path)
	}
}

func TestIntelXGetFile(t *testing.T) {
	want := []byte{1, 2, 3, 4}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/intelx/downloads/7/file" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", `attachment; filename="dump.bin"`)
		_, _ = w.Write(want)
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	f, err := c.IntelX.GetFile(context.Background(), "7")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(f.Data, want) {
		t.Fatalf("data = %v, want %v", f.Data, want)
	}
	if f.ContentType != "application/octet-stream" {
		t.Fatalf("contentType = %q", f.ContentType)
	}
	if f.Filename != "dump.bin" {
		t.Fatalf("filename = %q", f.Filename)
	}
}

func TestParseFilename(t *testing.T) {
	cases := map[string]string{
		`attachment; filename="dump.bin"`:                         "dump.bin",
		`attachment; filename=plain.bin`:                          "plain.bin",
		`attachment; filename*=UTF-8''r%C3%A9sum%C3%A9.bin`:       "résumé.bin",
		`attachment; filename="x.bin"; filename*=UTF-8''y%20.bin`: "y .bin",
		`attachment; filename*=UTF-8'en'file.bin`:                 "file.bin",
		``: "",
	}
	for in, want := range cases {
		if got := parseFilename(in); got != want {
			t.Errorf("parseFilename(%q) = %q, want %q", in, got, want)
		}
	}
}
