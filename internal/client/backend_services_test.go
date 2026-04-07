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

func TestListBackendServices_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/backend-services", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[BackendService]{
			Total: 1,
			Items: []BackendService{{ID: "bs1", Name: "backend1"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	services, err := c.ListBackendServices(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, services, 1)
	assert.Equal(t, "bs1", services[0].ID)
}

func TestListBackendServices_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListBackendServices(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListBackendServices_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListBackendServices(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetBackendService_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/backend-services/bs1", r.URL.Path)
		json.NewEncoder(w).Encode(BackendService{ID: "bs1", Name: "backend1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	bs, err := c.GetBackendService(context.Background(), "cfg1", "bs1")
	require.NoError(t, err)
	assert.Equal(t, "bs1", bs.ID)
}

func TestGetBackendService_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetBackendService(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetBackendService_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetBackendService(context.Background(), "cfg1", "bs1")
	require.Error(t, err)
}

func TestCreateBackendService_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var bs BackendService
		require.NoError(t, json.Unmarshal(body, &bs))
		assert.Equal(t, "bs1", bs.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateBackendService(context.Background(), "cfg1", "bs1", &BackendService{ID: "bs1"})
	require.NoError(t, err)
}

func TestCreateBackendService_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateBackendService(context.Background(), "cfg1", "bs1", &BackendService{})
	require.Error(t, err)
}

func TestUpdateBackendService_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateBackendService(context.Background(), "cfg1", "bs1", &BackendService{ID: "bs1"})
	require.NoError(t, err)
}

func TestUpdateBackendService_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateBackendService(context.Background(), "cfg1", "bs1", &BackendService{})
	require.Error(t, err)
}

func TestDeleteBackendService_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteBackendService(context.Background(), "cfg1", "bs1")
	require.NoError(t, err)
}

func TestDeleteBackendService_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteBackendService(context.Background(), "cfg1", "bs1")
	require.Error(t, err)
}
