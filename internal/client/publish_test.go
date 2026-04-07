package client

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublish_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/tools/publish/cfg1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var buckets []PublishBucket
		require.NoError(t, json.Unmarshal(body, &buckets))
		require.Len(t, buckets, 1)
		assert.Equal(t, "bucket1", buckets[0].Name)
		assert.Equal(t, "https://bucket1.example.com", buckets[0].URL)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.Publish(context.Background(), "cfg1", []PublishBucket{
		{Name: "bucket1", URL: "https://bucket1.example.com"},
	})
	require.NoError(t, err)
}

func TestPublish_EmptyBuckets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var buckets []PublishBucket
		require.NoError(t, json.Unmarshal(body, &buckets))
		assert.Empty(t, buckets)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.Publish(context.Background(), "cfg1", []PublishBucket{})
	require.NoError(t, err)
}

func TestPublish_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"message":"publish in progress","detail":"busy"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.Publish(context.Background(), "cfg1", []PublishBucket{})
	require.Error(t, err)
}

func TestPublish_InternalServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"failed"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.Publish(context.Background(), "cfg1", []PublishBucket{{Name: "b1", URL: "http://b1"}})
	require.Error(t, err)
}
