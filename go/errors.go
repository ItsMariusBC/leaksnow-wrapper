package leaksnow

import (
	"fmt"
	"time"
)

// APIError is the base API error returned for any non-2xx HTTP response.
// It is exported as Error via a type alias so callers use *Error.
type APIError struct {
	StatusCode int
	Status     string
	Body       []byte
	RetryAfter time.Duration
}

func (e *APIError) Error() string {
	return fmt.Sprintf("leaksnow: request failed: %d %s", e.StatusCode, e.Status)
}

// Error is the public name for the base API error type.
type Error = APIError

// Typed error kinds. Each embeds *APIError so fields are promoted and
// errors.As(err, &target) works for the specific kind. The base type is
// named APIError internally to avoid the field/method name collision that
// arises when embedding a type whose name matches its own method.
type (
	// AuthError is returned for 401 and 403 responses.
	AuthError struct{ *APIError }
	// QuotaError is returned for 429 responses; RetryAfter may be set.
	QuotaError struct{ *APIError }
	// ValidationError is returned for 400 and 422 responses.
	ValidationError struct{ *APIError }
	// ServerError is returned for 5xx responses.
	ServerError struct{ *APIError }
)

func errorFromResponse(status int, statusText string, body []byte, retryAfter time.Duration) error {
	base := &APIError{StatusCode: status, Status: statusText, Body: body, RetryAfter: retryAfter}
	switch {
	case status == 401 || status == 403:
		return &AuthError{base}
	case status == 429:
		return &QuotaError{base}
	case status == 400 || status == 422:
		return &ValidationError{base}
	case status >= 500:
		return &ServerError{base}
	default:
		return base
	}
}

// TransportError wraps a network/timeout failure (no HTTP response received).
type TransportError struct {
	Op      string
	Err     error
	Timeout bool
}

func (e *TransportError) Error() string {
	if e.Timeout {
		return fmt.Sprintf("leaksnow: %s: timed out: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("leaksnow: %s: %v", e.Op, e.Err)
}

func (e *TransportError) Unwrap() error { return e.Err }
