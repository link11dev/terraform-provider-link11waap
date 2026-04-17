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

func TestListGlobalFilters_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/global-filters", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		json.NewEncoder(w).Encode(ListResponse[GlobalFilter]{
			Total: 1,
			Items: []GlobalFilter{{ID: "gf1", Name: "test-filter", Source: "https://example.com/filter.json", Active: true}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	filters, err := c.ListGlobalFilters(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, filters, 1)
	assert.Equal(t, "gf1", filters[0].ID)
}

func TestListGlobalFilters_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListGlobalFilters(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetGlobalFilter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/global-filters/gf1", r.URL.Path)
		json.NewEncoder(w).Encode(GlobalFilter{ID: "gf1", Name: "test", Source: "https://example.com/filter.json", Active: true})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	f, err := c.GetGlobalFilter(context.Background(), "cfg1", "gf1")
	require.NoError(t, err)
	assert.Equal(t, "gf1", f.ID)
}

func TestGetGlobalFilter_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetGlobalFilter(context.Background(), "cfg1", "missing")
	require.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestCreateGlobalFilter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/global-filters/gf1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var f GlobalFilter
		require.NoError(t, json.Unmarshal(body, &f))
		assert.Equal(t, "gf1", f.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateGlobalFilter(context.Background(), "cfg1", "gf1", &GlobalFilter{ID: "gf1", Name: "test", Source: "https://example.com/filter.json"})
	require.NoError(t, err)
}

func TestCreateGlobalFilter_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"code":422,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateGlobalFilter(context.Background(), "cfg1", "gf1", &GlobalFilter{})
	require.Error(t, err)
}

func TestUpdateGlobalFilter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/conf/cfg1/global-filters/gf1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateGlobalFilter(context.Background(), "cfg1", "gf1", &GlobalFilter{ID: "gf1"})
	require.NoError(t, err)
}

func TestUpdateGlobalFilter_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"code":422,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateGlobalFilter(context.Background(), "cfg1", "gf1", &GlobalFilter{})
	require.Error(t, err)
}

func TestDeleteGlobalFilter_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/global-filters/gf1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteGlobalFilter(context.Background(), "cfg1", "gf1")
	require.NoError(t, err)
}

func TestDeleteGlobalFilter_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteGlobalFilter(context.Background(), "cfg1", "gf1")
	require.Error(t, err)
}
