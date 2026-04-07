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

func TestListACLProfiles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/acl-profiles", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		json.NewEncoder(w).Encode(ListResponse[ACLProfile]{
			Total: 1,
			Items: []ACLProfile{{ID: "acl1", Name: "test-acl"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	profiles, err := c.ListACLProfiles(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, "acl1", profiles[0].ID)
}

func TestListACLProfiles_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListACLProfiles(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListACLProfiles_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListACLProfiles(context.Background(), "cfg1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "decoding response")
}

func TestGetACLProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/acl-profiles/acl1", r.URL.Path)
		json.NewEncoder(w).Encode(ACLProfile{ID: "acl1", Name: "test"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	p, err := c.GetACLProfile(context.Background(), "cfg1", "acl1")
	require.NoError(t, err)
	assert.Equal(t, "acl1", p.ID)
}

func TestGetACLProfile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetACLProfile(context.Background(), "cfg1", "missing")
	require.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestGetACLProfile_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetACLProfile(context.Background(), "cfg1", "acl1")
	require.Error(t, err)
}

func TestCreateACLProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/acl-profiles/acl1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var p ACLProfile
		require.NoError(t, json.Unmarshal(body, &p))
		assert.Equal(t, "acl1", p.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateACLProfile(context.Background(), "cfg1", "acl1", &ACLProfile{ID: "acl1", Name: "test"})
	require.NoError(t, err)
}

func TestCreateACLProfile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateACLProfile(context.Background(), "cfg1", "acl1", &ACLProfile{})
	require.Error(t, err)
}

func TestUpdateACLProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/conf/cfg1/acl-profiles/acl1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateACLProfile(context.Background(), "cfg1", "acl1", &ACLProfile{ID: "acl1"})
	require.NoError(t, err)
}

func TestUpdateACLProfile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateACLProfile(context.Background(), "cfg1", "acl1", &ACLProfile{})
	require.Error(t, err)
}

func TestDeleteACLProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/acl-profiles/acl1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteACLProfile(context.Background(), "cfg1", "acl1")
	require.NoError(t, err)
}

func TestDeleteACLProfile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteACLProfile(context.Background(), "cfg1", "acl1")
	require.Error(t, err)
}
