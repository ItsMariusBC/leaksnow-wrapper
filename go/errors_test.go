package leaksnow

import (
	"errors"
	"testing"
	"time"
)

func TestErrorFromResponse(t *testing.T) {
	if err := errorFromResponse(401, "Unauthorized", nil, 0); !errors.As(err, new(*AuthError)) {
		t.Fatalf("401 should be *AuthError, got %T", err)
	}
	if err := errorFromResponse(403, "Forbidden", nil, 0); !errors.As(err, new(*AuthError)) {
		t.Fatalf("403 should be *AuthError, got %T", err)
	}
	var qe *QuotaError
	err := errorFromResponse(429, "Too Many Requests", nil, 2*time.Second)
	if !errors.As(err, &qe) {
		t.Fatalf("429 should be *QuotaError, got %T", err)
	}
	if qe.RetryAfter != 2*time.Second {
		t.Fatalf("RetryAfter = %v, want 2s", qe.RetryAfter)
	}
	if err := errorFromResponse(422, "Unprocessable", nil, 0); !errors.As(err, new(*ValidationError)) {
		t.Fatalf("422 should be *ValidationError, got %T", err)
	}
	if err := errorFromResponse(503, "Unavailable", nil, 0); !errors.As(err, new(*ServerError)) {
		t.Fatalf("503 should be *ServerError, got %T", err)
	}
	err = errorFromResponse(418, "Teapot", nil, 0)
	var base *Error
	if !errors.As(err, &base) || base.StatusCode != 418 {
		t.Fatalf("418 should be base *Error with status 418, got %T", err)
	}
	if errors.As(err, new(*AuthError)) {
		t.Fatalf("418 must not match *AuthError")
	}
}

func TestErrorMessage(t *testing.T) {
	e := &Error{StatusCode: 500, Status: "Server Error"}
	if e.Error() == "" {
		t.Fatal("Error() must not be empty")
	}
}
