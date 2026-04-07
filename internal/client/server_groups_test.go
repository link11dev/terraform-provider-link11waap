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

func TestListServerGroups_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/server-groups", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[ServerGroup]{
			Total: 1,
			Items: []ServerGroup{{ID: "sg1", Name: "site1", ServerNames: []string{"example.com"}}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	groups, err := c.ListServerGroups(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "sg1", groups[0].ID)
}

func TestListServerGroups_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListServerGroups(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListServerGroups_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListServerGroups(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetServerGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/server-groups/sg1", r.URL.Path)
		json.NewEncoder(w).Encode(ServerGroup{ID: "sg1", Name: "site1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	sg, err := c.GetServerGroup(context.Background(), "cfg1", "sg1")
	require.NoError(t, err)
	assert.Equal(t, "sg1", sg.ID)
}

func TestGetServerGroup_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetServerGroup(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetServerGroup_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetServerGroup(context.Background(), "cfg1", "sg1")
	require.Error(t, err)
}

func TestCreateServerGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var sg ServerGroupCreateRequest
		require.NoError(t, json.Unmarshal(body, &sg))
		assert.Equal(t, "sg1", sg.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateServerGroup(context.Background(), "cfg1", "sg1", &ServerGroupCreateRequest{
		ID:          "sg1",
		Name:        "site1",
		ServerNames: []string{"example.com"},
	})
	require.NoError(t, err)
}

func TestCreateServerGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateServerGroup(context.Background(), "cfg1", "sg1", &ServerGroupCreateRequest{})
	require.Error(t, err)
}

func TestUpdateServerGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateServerGroup(context.Background(), "cfg1", "sg1", &ServerGroupCreateRequest{ID: "sg1"})
	require.NoError(t, err)
}

func TestUpdateServerGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateServerGroup(context.Background(), "cfg1", "sg1", &ServerGroupCreateRequest{})
	require.Error(t, err)
}

func TestDeleteServerGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteServerGroup(context.Background(), "cfg1", "sg1")
	require.NoError(t, err)
}

func TestDeleteServerGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteServerGroup(context.Background(), "cfg1", "sg1")
	require.Error(t, err)
}
