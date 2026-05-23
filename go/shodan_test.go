package leaksnow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func recordingServer(t *testing.T) (*httptest.Server, *string, *string) {
	t.Helper()
	var path, method string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		method = r.Method
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	return srv, &path, &method
}

func TestShodan(t *testing.T) {
	srv, path, method := recordingServer(t)
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	ctx := context.Background()

	if _, err := c.Shodan.CustomScans(ctx); err != nil {
		t.Fatal(err)
	}
	if *path != "/api/v1/shodan/custom-scans" || *method != http.MethodGet {
		t.Fatalf("customScans: %s %s", *method, *path)
	}

	if _, err := c.Shodan.CustomScan(ctx, ShodanCustomScanRequest{Target: "scanme.nmap.org"}); err != nil {
		t.Fatal(err)
	}
	if *path != "/api/v1/shodan/custom-scan" || *method != http.MethodPost {
		t.Fatalf("customScan: %s %s", *method, *path)
	}

	if _, err := c.Shodan.GetScan(ctx, "42"); err != nil {
		t.Fatal(err)
	}
	if *path != "/api/v1/shodan/custom-scans/42" || *method != http.MethodGet {
		t.Fatalf("getScan: %s %s", *method, *path)
	}

	if _, err := c.Shodan.DeleteScan(ctx, "42"); err != nil {
		t.Fatal(err)
	}
	if *path != "/api/v1/shodan/custom-scans/42" || *method != http.MethodDelete {
		t.Fatalf("deleteScan: %s %s", *method, *path)
	}

	if _, err := c.Shodan.Host(ctx, "1.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if *path != "/api/v1/shodan/host/1.2.3.4" || *method != http.MethodGet {
		t.Fatalf("host: %s %s", *method, *path)
	}
}
