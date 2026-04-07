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

func TestListProxyTemplates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/proxy-templates", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[ProxyTemplate]{
			Total: 1,
			Items: []ProxyTemplate{{Name: "pt1", Description: "template1"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	templates, err := c.ListProxyTemplates(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, templates, 1)
	assert.Equal(t, "pt1", templates[0].Name)
}

func TestListProxyTemplates_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListProxyTemplates(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListProxyTemplates_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListProxyTemplates(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetProxyTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/proxy-templates/pt1", r.URL.Path)
		json.NewEncoder(w).Encode(ProxyTemplate{Name: "pt1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	pt, err := c.GetProxyTemplate(context.Background(), "cfg1", "pt1")
	require.NoError(t, err)
	assert.Equal(t, "pt1", pt.Name)
}

func TestGetProxyTemplate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetProxyTemplate(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetProxyTemplate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetProxyTemplate(context.Background(), "cfg1", "pt1")
	require.Error(t, err)
}

func TestCreateProxyTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var pt ProxyTemplate
		require.NoError(t, json.Unmarshal(body, &pt))
		assert.Equal(t, "pt1", pt.Name)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateProxyTemplate(context.Background(), "cfg1", "pt1", &ProxyTemplate{Name: "pt1"})
	require.NoError(t, err)
}

func TestCreateProxyTemplate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateProxyTemplate(context.Background(), "cfg1", "pt1", &ProxyTemplate{})
	require.Error(t, err)
}

func TestUpdateProxyTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateProxyTemplate(context.Background(), "cfg1", "pt1", &ProxyTemplate{Name: "pt1"})
	require.NoError(t, err)
}

func TestUpdateProxyTemplate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateProxyTemplate(context.Background(), "cfg1", "pt1", &ProxyTemplate{})
	require.Error(t, err)
}

func TestDeleteProxyTemplate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteProxyTemplate(context.Background(), "cfg1", "pt1")
	require.NoError(t, err)
}

func TestDeleteProxyTemplate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteProxyTemplate(context.Background(), "cfg1", "pt1")
	require.Error(t, err)
}
