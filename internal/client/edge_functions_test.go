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

func TestListEdgeFunctions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/edge-functions", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[EdgeFunction]{
			Total: 1,
			Items: []EdgeFunction{{ID: "ef1", Name: "func1", Phase: "request"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	funcs, err := c.ListEdgeFunctions(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, funcs, 1)
	assert.Equal(t, "ef1", funcs[0].ID)
}

func TestListEdgeFunctions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListEdgeFunctions(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListEdgeFunctions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListEdgeFunctions(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetEdgeFunction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/edge-functions/ef1", r.URL.Path)
		json.NewEncoder(w).Encode(EdgeFunction{ID: "ef1", Name: "func1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	ef, err := c.GetEdgeFunction(context.Background(), "cfg1", "ef1")
	require.NoError(t, err)
	assert.Equal(t, "ef1", ef.ID)
}

func TestGetEdgeFunction_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetEdgeFunction(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetEdgeFunction_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetEdgeFunction(context.Background(), "cfg1", "ef1")
	require.Error(t, err)
}

func TestCreateEdgeFunction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var ef EdgeFunction
		require.NoError(t, json.Unmarshal(body, &ef))
		assert.Equal(t, "ef1", ef.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateEdgeFunction(context.Background(), "cfg1", "ef1", &EdgeFunction{ID: "ef1", Name: "func1", Code: "return true", Phase: "request"})
	require.NoError(t, err)
}

func TestCreateEdgeFunction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateEdgeFunction(context.Background(), "cfg1", "ef1", &EdgeFunction{})
	require.Error(t, err)
}

func TestUpdateEdgeFunction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateEdgeFunction(context.Background(), "cfg1", "ef1", &EdgeFunction{ID: "ef1"})
	require.NoError(t, err)
}

func TestUpdateEdgeFunction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateEdgeFunction(context.Background(), "cfg1", "ef1", &EdgeFunction{})
	require.Error(t, err)
}

func TestDeleteEdgeFunction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteEdgeFunction(context.Background(), "cfg1", "ef1")
	require.NoError(t, err)
}

func TestDeleteEdgeFunction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteEdgeFunction(context.Background(), "cfg1", "ef1")
	require.Error(t, err)
}
