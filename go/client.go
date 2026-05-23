package leaksnow

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultBaseURL = "https://leaks.now"

// Client is a leaks.now /api/v1 API client. Create one with NewClient.
type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	retry      RetryConfig
	timeout    time.Duration // pending timeout from WithTimeout; applied after options

	// Service namespaces.
	Shodan *ShodanService
	ULP    *ULPService
	IntelX *IntelXService
}

// Option configures a Client in NewClient.
type Option func(*Client)

// WithBaseURL overrides the API base URL (default https://leaks.now).
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(u, "/") }
}

// WithHTTPClient sets a custom *http.Client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.httpClient = h
		}
	}
}

// WithTimeout sets the per-request timeout on the underlying http.Client,
// applied after all options so it is honored regardless of option order.
// It does not bound total time across retries.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.timeout = d }
}

// WithRetry enables and configures automatic retries (disabled by default).
func WithRetry(r RetryConfig) Option {
	return func(c *Client) { c.retry = r }
}

// NewClient creates a Client. The apiKey is sent as "Authorization: Bearer <key>".
func NewClient(apiKey string, opts ...Option) *Client {
	c := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		retry:      DefaultRetry,
	}
	for _, o := range opts {
		o(c)
	}
	if c.timeout > 0 {
		c.httpClient.Timeout = c.timeout
	}
	c.Shodan = &ShodanService{client: c}
	c.ULP = &ULPService{client: c}
	c.IntelX = &IntelXService{client: c}
	return c
}

func isTimeout(err error) bool {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded)
}

func parseRetryAfter(resp *http.Response) time.Duration {
	raw := resp.Header.Get("Retry-After")
	if raw == "" {
		return 0
	}
	// Only the numeric delta-seconds form is honored; HTTP-date form is ignored.
	secs, err := strconv.Atoi(raw)
	if err != nil || secs < 0 {
		return 0
	}
	return time.Duration(secs) * time.Second
}

// doRequest performs the HTTP request with auth, optional JSON body, and retries.
// The caller owns resp.Body and must close it on success.
func (c *Client) doRequest(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyBytes []byte
	hasBody := body != nil && method != http.MethodGet && method != http.MethodDelete
	if hasBody {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, &TransportError{Op: method + " " + path, Err: err}
		}
		bodyBytes = b
	}

	op := method + " " + path
	url := c.baseURL + path
	attempt := 0
	for {
		var reader io.Reader
		if hasBody {
			reader = bytes.NewReader(bodyBytes)
		}
		req, err := http.NewRequestWithContext(ctx, method, url, reader)
		if err != nil {
			return nil, &TransportError{Op: op, Err: err}
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		if hasBody {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, &TransportError{Op: op, Err: err, Timeout: isTimeout(err)}
		}

		if resp.StatusCode < 300 {
			return resp, nil
		}

		retryAfter := parseRetryAfter(resp)
		if shouldRetry(resp.StatusCode, attempt, c.retry) {
			_ = resp.Body.Close()
			delay := backoffDelay(attempt, c.retry, retryAfter, rand.Float64())
			t := time.NewTimer(delay)
			select {
			case <-t.C:
			case <-ctx.Done():
				t.Stop()
				return nil, &TransportError{Op: op, Err: ctx.Err(), Timeout: errors.Is(ctx.Err(), context.DeadlineExceeded)}
			}
			attempt++
			continue
		}

		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, errorFromResponse(resp.StatusCode, statusText(resp), b, retryAfter)
	}
}

func statusText(resp *http.Response) string {
	// resp.Status is like "401 Unauthorized"; strip the leading code if present.
	s := resp.Status
	if i := strings.IndexByte(s, ' '); i >= 0 {
		return s[i+1:]
	}
	return s
}

// doJSON performs a request and returns the raw JSON response body.
func (c *Client) doJSON(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	resp, err := c.doRequest(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &TransportError{Op: method + " " + path, Err: err, Timeout: isTimeout(err)}
	}
	return data, nil
}

// Search runs an OSINT search (leak/service/shodan). Cost: 1 credit.
func (c *Client) Search(ctx context.Context, req SearchRequest) (json.RawMessage, error) {
	return c.doJSON(ctx, http.MethodPost, "/api/v1/search", req)
}
