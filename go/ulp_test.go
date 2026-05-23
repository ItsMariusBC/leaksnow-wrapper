package leaksnow

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestULP(t *testing.T) {
	var path, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		_, _ = w.Write([]byte(`{"downloadToken":"tok123"}`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	ctx := context.Background()

	raw, err := c.ULP.Search(ctx, ULPSearchRequest{Type: ULPTypeDomain, Value: "example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/ulp/search" {
		t.Fatalf("search path = %q", path)
	}
	if body != `{"type":"domain","value":"example.com"}` {
		t.Fatalf("search body = %q", body)
	}
	if string(raw) != `{"downloadToken":"tok123"}` {
		t.Fatalf("raw = %s", raw)
	}

	if _, err := c.ULP.Download(ctx, ULPDownloadRequest{Token: "tok123"}); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/ulp/download" {
		t.Fatalf("download path = %q", path)
	}
	if body != `{"token":"tok123"}` {
		t.Fatalf("download body = %q", body)
	}
}
