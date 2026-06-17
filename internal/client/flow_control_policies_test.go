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

func TestListFlowControlPolicies_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/flow-control-policies", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[FlowControl]{
			Total: 1,
			Items: []FlowControl{{ID: "fc1", Name: "flow1", Timeframe: 60}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	policies, err := c.ListFlowControlPolicies(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, policies, 1)
	assert.Equal(t, "fc1", policies[0].ID)
}

func TestListFlowControlPolicies_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListFlowControlPolicies(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListFlowControlPolicies_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListFlowControlPolicies(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetFlowControlPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/flow-control-policies/fc1", r.URL.Path)
		json.NewEncoder(w).Encode(FlowControl{ID: "fc1", Name: "flow1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	policy, err := c.GetFlowControlPolicy(context.Background(), "cfg1", "fc1")
	require.NoError(t, err)
	assert.Equal(t, "fc1", policy.ID)
}

func TestGetFlowControlPolicy_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetFlowControlPolicy(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetFlowControlPolicy_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetFlowControlPolicy(context.Background(), "cfg1", "fc1")
	require.Error(t, err)
}

func TestCreateFlowControlPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var policy FlowControl
		require.NoError(t, json.Unmarshal(body, &policy))
		assert.Equal(t, "fc1", policy.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateFlowControlPolicy(context.Background(), "cfg1", "fc1", &FlowControl{ID: "fc1", Name: "flow1"})
	require.NoError(t, err)
}

func TestCreateFlowControlPolicy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateFlowControlPolicy(context.Background(), "cfg1", "fc1", &FlowControl{})
	require.Error(t, err)
}

func TestUpdateFlowControlPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateFlowControlPolicy(context.Background(), "cfg1", "fc1", &FlowControl{ID: "fc1"})
	require.NoError(t, err)
}

func TestUpdateFlowControlPolicy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateFlowControlPolicy(context.Background(), "cfg1", "fc1", &FlowControl{})
	require.Error(t, err)
}

func TestDeleteFlowControlPolicy_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteFlowControlPolicy(context.Background(), "cfg1", "fc1")
	require.NoError(t, err)
}

func TestDeleteFlowControlPolicy_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteFlowControlPolicy(context.Background(), "cfg1", "fc1")
	require.Error(t, err)
}
