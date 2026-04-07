package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Default values for client configuration
const (
	DefaultTimeout   = 30 * time.Second
	DefaultUserAgent = "terraform-provider-link11waap"
	APIVersion       = "v4.3"
	DefaultRetryMax  = 5
	DefaultRetryWait = 1 * time.Second
	MaxRetryWait     = 60 * time.Second
)

// Config holds the client configuration
type Config struct {
	Domain    string        // Customer domain (e.g., "customer.app.reblaze.io")
	APIKey    string        // API key for Basic auth
	Timeout   time.Duration // HTTP request timeout
	UserAgent string        // User-Agent header value
	Debug     bool          // Enable debug logging
	RetryMax  int           // Maximum number of retries for 503 errors
	RetryWait time.Duration // Initial wait time between retries
}

// Client is the Link11 WAAP API client
type Client struct {
	baseURL    *url.URL
	apiKey     string
	httpClient *http.Client
	userAgent  string
	debug      bool
	retryMax   int
	retryWait  time.Duration
}

// New creates a new API client
func New(cfg Config) (*Client, error) {
	if cfg.Domain == "" {
		return nil, fmt.Errorf("domain is required")
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("api_key is required")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	userAgent := cfg.UserAgent
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	retryMax := cfg.RetryMax
	if retryMax == 0 {
		retryMax = DefaultRetryMax
	}

	retryWait := cfg.RetryWait
	if retryWait == 0 {
		retryWait = DefaultRetryWait
	}

	parsedURL, err := url.Parse(fmt.Sprintf("https://%s/api/%s", cfg.Domain, APIVersion))
	if err != nil {
		return nil, fmt.Errorf("Error parsing base URL: %w", err)
	}

	return &Client{
		baseURL:   parsedURL,
		apiKey:    cfg.APIKey,
		userAgent: userAgent,
		debug:     cfg.Debug,
		retryMax:  retryMax,
		retryWait: retryWait,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// doRequest performs an HTTP request with retry logic for 503 errors
func (c *Client) doRequest(ctx context.Context, method, path string, body []byte) (*http.Response, error) {
	url := c.baseURL.String() + path

	var lastErr error
	retryWait := c.retryWait

	for attempt := 0; attempt <= c.retryMax; attempt++ {
		var reqBody io.Reader
		if body != nil {
			reqBody = bytes.NewReader(body)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		// Set headers
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", c.apiKey))
		req.Header.Set("User-Agent", c.userAgent)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			continue
		}

		// Retry on 503 (publish in progress)
		if resp.StatusCode == http.StatusServiceUnavailable {
			resp.Body.Close()
			lastErr = fmt.Errorf("service unavailable (503): publish in progress")

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryWait):
				// Exponential backoff with cap
				retryWait = retryWait * 2
				if retryWait > MaxRetryWait {
					retryWait = MaxRetryWait
				}
				continue
			}
		}

		return resp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Get performs a GET request
func (c *Client) Get(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// Post performs a POST request
func (c *Client) Post(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPost, path, body)
}

// Put performs a PUT request
func (c *Client) Put(ctx context.Context, path string, body []byte) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodPut, path, body)
}

// Delete performs a DELETE request
func (c *Client) Delete(ctx context.Context, path string) (*http.Response, error) {
	return c.doRequest(ctx, http.MethodDelete, path, nil)
}

// DecodeResponse decodes a JSON response into the provided target
func DecodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
