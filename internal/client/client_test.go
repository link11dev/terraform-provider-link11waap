package client

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newTestClient creates a Client pointing at the given httptest.Server.
func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	return &Client{
		baseURL:    parsedURL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		userAgent:  "test-agent",
		debug:      false,
		retryMax:   0,
		retryWait:  1 * time.Millisecond,
	}
}

// newTestClientWithRetry creates a Client with retry support.
func newTestClientWithRetry(t *testing.T, server *httptest.Server, retryMax int) *Client {
	t.Helper()
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	return &Client{
		baseURL:    parsedURL,
		apiKey:     "test-api-key",
		httpClient: server.Client(),
		userAgent:  "test-agent",
		debug:      false,
		retryMax:   retryMax,
		retryWait:  1 * time.Millisecond,
	}
}

// --- Tests for New() ---

func TestNew_Success(t *testing.T) {
	c, err := New(Config{
		Domain: "example.reblaze.io",
		APIKey: "my-key",
	})
	require.NoError(t, err)
	assert.Equal(t, "https://example.reblaze.io/api/v4.3", c.baseURL.String())
	assert.Equal(t, "my-key", c.apiKey)
	assert.Equal(t, DefaultUserAgent, c.userAgent)
	assert.Equal(t, DefaultRetryMax, c.retryMax)
	assert.Equal(t, DefaultRetryWait, c.retryWait)
	assert.False(t, c.debug)
}

func TestNew_CustomValues(t *testing.T) {
	c, err := New(Config{
		Domain:    "custom.io",
		APIKey:    "key",
		Timeout:   10 * time.Second,
		UserAgent: "custom-agent",
		Debug:     true,
		RetryMax:  3,
		RetryWait: 5 * time.Second,
	})
	require.NoError(t, err)
	assert.Equal(t, "https://custom.io/api/v4.3", c.baseURL.String())
	assert.Equal(t, "custom-agent", c.userAgent)
	assert.True(t, c.debug)
	assert.Equal(t, 3, c.retryMax)
	assert.Equal(t, 5*time.Second, c.retryWait)
}

func TestNew_ValidURL(t *testing.T) {
	c, err := New(Config{
		Domain: "valid-domain.io",
		APIKey: "key",
	})
	require.NoError(t, err)
	require.NotNil(t, c.baseURL)
	assert.Equal(t, "https", c.baseURL.Scheme)
	assert.Equal(t, "valid-domain.io", c.baseURL.Host)
	assert.Equal(t, "/api/v4.3", c.baseURL.Path)
}

func TestNew_InvalidURL(t *testing.T) {
	_, err := New(Config{
		Domain: "invalid domain with spaces",
		APIKey: "key",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Error parsing base URL")
}

func TestNew_MissingDomain(t *testing.T) {
	_, err := New(Config{APIKey: "key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "domain is required")
}

func TestNew_MissingAPIKey(t *testing.T) {
	_, err := New(Config{Domain: "example.io"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "api_key is required")
}

// --- Tests for doRequest / Get / Post / Put / Delete ---

func TestGet_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test-path", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Accept"))
		assert.Equal(t, "Basic test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "test-agent", r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Get(context.Background(), "/test-path")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPost_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"name":"test"}`, string(body))
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Post(context.Background(), "/items", []byte(`{"name":"test"}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestPut_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Put(context.Background(), "/items/1", []byte(`{}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Delete(context.Background(), "/items/1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDoRequest_Retry503(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	c := newTestClientWithRetry(t, server, 3)
	resp, err := c.Get(context.Background(), "/test")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, attempts)
}

func TestDoRequest_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	c := newTestClientWithRetry(t, server, 1)
	_, err := c.Get(context.Background(), "/test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries exceeded")
}

func TestDoRequest_ContextCanceled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	c := newTestClientWithRetry(t, server, 5)
	_, err := c.Get(ctx, "/test")
	require.Error(t, err)
}

func TestDoRequest_NilBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		assert.Empty(t, body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	resp, err := c.Get(context.Background(), "/test")
	require.NoError(t, err)
	resp.Body.Close()
}

// --- Tests for DecodeResponse ---

func TestDecodeResponse_Success(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(`{"id":"123","name":"test"}`)),
	}
	var result struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	err := DecodeResponse(resp, &result)
	require.NoError(t, err)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "test", result.Name)
}

func TestDecodeResponse_InvalidJSON(t *testing.T) {
	resp := &http.Response{
		Body: io.NopCloser(strings.NewReader(`not json`)),
	}
	var result map[string]interface{}
	err := DecodeResponse(resp, &result)
	require.Error(t, err)
}

// --- Tests for doRequest with body rewind on retry ---

func TestDoRequest_BodyRewindOnRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		body, _ := io.ReadAll(r.Body)
		assert.Equal(t, `{"data":"value"}`, string(body))
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClientWithRetry(t, server, 2)
	resp, err := c.Post(context.Background(), "/test", []byte(`{"data":"value"}`))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 2, attempts)
}
