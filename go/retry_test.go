package leaksnow

import (
	"testing"
	"time"
)

func TestBackoffDelay(t *testing.T) {
	c := RetryConfig{MaxRetries: 3, BaseDelay: 500 * time.Millisecond, MaxDelay: 8 * time.Second, RetryOn: []int{429, 500, 502, 503, 504}}

	if got := backoffDelay(0, c, 1*time.Second, 0.5); got != 1*time.Second {
		t.Fatalf("retryAfter: got %v, want 1s", got)
	}
	if got := backoffDelay(0, c, 1*time.Hour, 0.5); got != 8*time.Second {
		t.Fatalf("retryAfter cap: got %v, want 8s", got)
	}
	if got := backoffDelay(0, c, 0, 0.5); got != 250*time.Millisecond {
		t.Fatalf("jitter attempt0: got %v, want 250ms", got)
	}
	if got := backoffDelay(2, c, 0, 0.5); got != 1*time.Second {
		t.Fatalf("jitter attempt2: got %v, want 1s", got)
	}
	if got := backoffDelay(20, c, 0, 1.0); got > 8*time.Second {
		t.Fatalf("cap: got %v, want <= 8s", got)
	}
}

func TestShouldRetry(t *testing.T) {
	c := RetryConfig{MaxRetries: 2, RetryOn: []int{429, 503}}
	if !shouldRetry(429, 0, c) {
		t.Fatal("429 attempt 0 should retry")
	}
	if !shouldRetry(503, 1, c) {
		t.Fatal("503 attempt 1 should retry")
	}
	if shouldRetry(429, 2, c) {
		t.Fatal("attempts exhausted should not retry")
	}
	if shouldRetry(400, 0, c) {
		t.Fatal("unlisted status should not retry")
	}
}

func TestDefaultRetryDisabled(t *testing.T) {
	if DefaultRetry.MaxRetries != 0 {
		t.Fatalf("DefaultRetry.MaxRetries = %d, want 0 (disabled)", DefaultRetry.MaxRetries)
	}
}
