package leaksnow

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDoJSONSendsBearerAndParses(t *testing.T) {
	var gotAuth, gotCT, gotBody, gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCT = r.Header.Get("Content-Type")
		gotMethod = r.Method
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer srv.Close()

	c := NewClient("ms_key", WithBaseURL(srv.URL))
	raw, err := c.doJSON(context.Background(), http.MethodPost, "/api/v1/search", SearchRequest{Query: "x"})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("raw = %s", raw)
	}
	if gotAuth != "Bearer ms_key" {
		t.Fatalf("auth = %q", gotAuth)
	}
	if gotCT != "application/json" {
		t.Fatalf("content-type = %q", gotCT)
	}
	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q", gotMethod)
	}
	if gotBody != `{"query":"x"}` {
		t.Fatalf("body = %q", gotBody)
	}
}

func TestDoJSONOmitsBodyForGET(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if len(b) != 0 {
			t.Errorf("GET should have no body, got %q", b)
		}
		if r.Header.Get("Content-Type") != "" {
			t.Errorf("GET should not set Content-Type")
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	if _, err := c.doJSON(context.Background(), http.MethodGet, "/api/v1/shodan/custom-scans", nil); err != nil {
		t.Fatal(err)
	}
}

func TestDoJSONMapsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad key"}`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	_, err := c.doJSON(context.Background(), http.MethodPost, "/api/v1/search", SearchRequest{})
	if !errors.As(err, new(*AuthError)) {
		t.Fatalf("want *AuthError, got %T", err)
	}
}

func TestDoJSONRetriesThenSucceeds(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			w.Header().Set("Retry-After", "0")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"ok":1}`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL),
		WithRetry(RetryConfig{MaxRetries: 2, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, RetryOn: []int{429}}))
	raw, err := c.doJSON(context.Background(), http.MethodPost, "/api/v1/search", SearchRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != `{"ok":1}` {
		t.Fatalf("raw = %s", raw)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestDoJSONQuotaErrorAfterExhaustion(t *testing.T) {
	var calls int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL),
		WithRetry(RetryConfig{MaxRetries: 1, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, RetryOn: []int{429}}))
	_, err := c.doJSON(context.Background(), http.MethodPost, "/api/v1/search", SearchRequest{})
	if !errors.As(err, new(*QuotaError)) {
		t.Fatalf("want *QuotaError, got %T", err)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2 (initial + 1 retry)", calls)
	}
}

func TestDoRequestTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL), WithTimeout(5*time.Millisecond))
	_, err := c.doJSON(context.Background(), http.MethodGet, "/slow", nil)
	var te *TransportError
	if !errors.As(err, &te) {
		t.Fatalf("want *TransportError, got %T", err)
	}
	if !te.Timeout {
		t.Fatal("TransportError.Timeout should be true")
	}
}

func TestSearchHitsEndpoint(t *testing.T) {
	var path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, _ = w.Write([]byte(`{"results":[]}`))
	}))
	defer srv.Close()
	c := NewClient("ms_key", WithBaseURL(srv.URL))
	raw, err := c.Search(context.Background(), SearchRequest{Query: "host:example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/search" {
		t.Fatalf("path = %q", path)
	}
	var out map[string]json.RawMessage
	if json.Unmarshal(raw, &out); out == nil {
		t.Fatal("response not decodable")
	}
}
