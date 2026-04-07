package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListConfigs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/configs", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		json.NewEncoder(w).Encode(ListResponse[WAAPConfig]{
			Total: 1,
			Items: []WAAPConfig{{ID: "cfg1", Description: "test config"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	configs, err := c.ListConfigs(context.Background())
	require.NoError(t, err)
	require.Len(t, configs, 1)
	assert.Equal(t, "cfg1", configs[0].ID)
}

func TestListConfigs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"server error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListConfigs(context.Background())
	require.Error(t, err)
}

func TestListConfigs_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListConfigs(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding response")
}

func TestGetConfig_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/configs/cfg1", r.URL.Path)
		json.NewEncoder(w).Encode(WAAPConfig{ID: "cfg1", Description: "test"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cfg, err := c.GetConfig(context.Background(), "cfg1")
	require.NoError(t, err)
	assert.Equal(t, "cfg1", cfg.ID)
}

func TestGetConfig_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetConfig(context.Background(), "missing")
	require.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestGetConfig_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetConfig(context.Background(), "cfg1")
	require.Error(t, err)
}
