package leaksnow

import "time"

// RetryConfig controls automatic retries. Retries are disabled by default
// (MaxRetries == 0); set a positive MaxRetries to enable.
type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
	RetryOn    []int
}

// DefaultRetry has retries disabled but carries sensible values used once enabled.
var DefaultRetry = RetryConfig{
	MaxRetries: 0,
	BaseDelay:  500 * time.Millisecond,
	MaxDelay:   8 * time.Second,
	RetryOn:    []int{429, 500, 502, 503, 504},
}

// backoffDelay computes a full-jitter exponential backoff. rnd must be in [0,1).
// When retryAfter > 0 it takes precedence, capped at MaxDelay.
func backoffDelay(attempt int, c RetryConfig, retryAfter time.Duration, rnd float64) time.Duration {
	if retryAfter > 0 {
		if retryAfter > c.MaxDelay {
			return c.MaxDelay
		}
		return retryAfter
	}
	exp := c.BaseDelay * time.Duration(1<<attempt)
	if exp > c.MaxDelay || exp <= 0 { // exp <= 0 guards against shift overflow
		exp = c.MaxDelay
	}
	return time.Duration(rnd * float64(exp))
}

func shouldRetry(status, attempt int, c RetryConfig) bool {
	if attempt >= c.MaxRetries {
		return false
	}
	for _, s := range c.RetryOn {
		if s == status {
			return true
		}
	}
	return false
}
