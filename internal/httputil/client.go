package httputil

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// Client wraps http.Client with exponential backoff, jitter, and 429 rate-limit handling
type Client struct {
	httpClient *http.Client
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

// ClientOption allows customizing Client behavior
type ClientOption func(*Client)

// WithMaxRetries sets the maximum number of retries
func WithMaxRetries(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

// WithBaseDelay sets the initial backoff delay
func WithBaseDelay(delay time.Duration) ClientOption {
	return func(c *Client) {
		c.baseDelay = delay
	}
}

// WithHTTPClient sets a custom underlying http.Client
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// NewClient creates an initialized resilient HTTP client
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		maxRetries: 3,
		baseDelay:  500 * time.Millisecond,
		maxDelay:   10 * time.Second,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Do executes an HTTP request with automatic retries for transient errors and rate limits
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// Clone request body if reusable reader is available
		resp, err = c.httpClient.Do(req)
		if err != nil {
			// If context was cancelled, return immediately
			if req.Context().Err() != nil {
				return nil, req.Context().Err()
			}

			// Transient network error, retry if attempts remain
			if attempt == c.maxRetries {
				return nil, fmt.Errorf("http request failed after %d attempts: %w", attempt+1, err)
			}

			sleepDuration := c.calculateDelay(attempt, nil)
			if err := sleepWithContext(req.Context(), sleepDuration); err != nil {
				return nil, err
			}
			continue
		}

		// Handle 429 Too Many Requests or 5xx Server Errors
		if resp.StatusCode == http.StatusTooManyRequests || (resp.StatusCode >= 500 && resp.StatusCode <= 504) {
			if attempt == c.maxRetries {
				// Return response on final attempt so caller can inspect error body
				return resp, nil
			}

			delay := c.calculateDelay(attempt, resp)
			// Drain and close body before retrying
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
			resp.Body.Close()

			if err := sleepWithContext(req.Context(), delay); err != nil {
				return nil, err
			}
			continue
		}

		// Non-retryable status (2xx, 3xx, 4xx except 429)
		return resp, nil
	}

	return resp, err
}

func (c *Client) calculateDelay(attempt int, resp *http.Response) time.Duration {
	if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
		// Check Retry-After header
		if retryAfterStr := resp.Header.Get("Retry-After"); retryAfterStr != "" {
			if seconds, err := strconv.Atoi(retryAfterStr); err == nil && seconds > 0 {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	// Exponential backoff: baseDelay * 2^attempt
	multiplier := 1 << attempt
	backoff := c.baseDelay * time.Duration(multiplier)
	if backoff > c.maxDelay {
		backoff = c.maxDelay
	}

	// Add jitter (10% - 30%)
	jitter := time.Duration(rand.Int63n(int64(backoff / 3)))
	return backoff + jitter
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
