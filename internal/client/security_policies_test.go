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

func TestListSecurityPolicies_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/security-policies", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[SecurityPolicy]{
			Total: 1,
			Items: []SecurityPolicy{{ID: "sp1", Name: "policy1"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	policies, err := c.ListSecurityPolicies(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, policies, 1)
	assert.Equal(t, "sp1", policies[0].ID)
}

func TestListSecurityPolicies_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListSecurityPolicies(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListSecurityPolicies_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListSecurityPolicies(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetSecurityPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/security-policies/sp1", r.URL.Path)
		json.NewEncoder(w).Encode(SecurityPolicy{ID: "sp1", Name: "policy1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	sp, err := c.GetSecurityPolicy(context.Background(), "cfg1", "sp1")
	require.NoError(t, err)
	assert.Equal(t, "sp1", sp.ID)
}

func TestGetSecurityPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetSecurityPolicy(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetSecurityPolicy_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetSecurityPolicy(context.Background(), "cfg1", "sp1")
	require.Error(t, err)
}

func TestCreateSecurityPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var sp SecurityPolicy
		require.NoError(t, json.Unmarshal(body, &sp))
		assert.Equal(t, "sp1", sp.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateSecurityPolicy(context.Background(), "cfg1", "sp1", &SecurityPolicy{ID: "sp1", Name: "policy1"})
	require.NoError(t, err)
}

func TestCreateSecurityPolicy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateSecurityPolicy(context.Background(), "cfg1", "sp1", &SecurityPolicy{})
	require.Error(t, err)
}

func TestUpdateSecurityPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateSecurityPolicy(context.Background(), "cfg1", "sp1", &SecurityPolicy{ID: "sp1"})
	require.NoError(t, err)
}

func TestUpdateSecurityPolicy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateSecurityPolicy(context.Background(), "cfg1", "sp1", &SecurityPolicy{})
	require.Error(t, err)
}

func TestDeleteSecurityPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteSecurityPolicy(context.Background(), "cfg1", "sp1")
	require.NoError(t, err)
}

func TestDeleteSecurityPolicy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteSecurityPolicy(context.Background(), "cfg1", "sp1")
	require.Error(t, err)
}
