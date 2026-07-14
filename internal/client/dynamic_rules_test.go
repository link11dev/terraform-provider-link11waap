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

func TestListDynamicRules_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/dynamic-rules", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[DynamicRule]{
			Total: 1,
			Items: []DynamicRule{{ID: "dr1", Name: "dyn1", Timeframe: 60, Threshold: 100}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	rules, err := c.ListDynamicRules(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, "dr1", rules[0].ID)
}

func TestListDynamicRules_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListDynamicRules(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListDynamicRules_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListDynamicRules(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetDynamicRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/dynamic-rules/dr1", r.URL.Path)
		json.NewEncoder(w).Encode(DynamicRule{ID: "dr1", Name: "dyn1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	rule, err := c.GetDynamicRule(context.Background(), "cfg1", "dr1")
	require.NoError(t, err)
	assert.Equal(t, "dr1", rule.ID)
}

func TestGetDynamicRule_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetDynamicRule(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetDynamicRule_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetDynamicRule(context.Background(), "cfg1", "dr1")
	require.Error(t, err)
}

func TestCreateDynamicRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var rule DynamicRule
		require.NoError(t, json.Unmarshal(body, &rule))
		assert.Equal(t, "dr1", rule.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateDynamicRule(context.Background(), "cfg1", "dr1", &DynamicRule{ID: "dr1", Name: "dyn1"})
	require.NoError(t, err)
}

func TestCreateDynamicRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateDynamicRule(context.Background(), "cfg1", "dr1", &DynamicRule{})
	require.Error(t, err)
}

func TestUpdateDynamicRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateDynamicRule(context.Background(), "cfg1", "dr1", &DynamicRule{ID: "dr1"})
	require.NoError(t, err)
}

func TestUpdateDynamicRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateDynamicRule(context.Background(), "cfg1", "dr1", &DynamicRule{})
	require.Error(t, err)
}

func TestDeleteDynamicRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteDynamicRule(context.Background(), "cfg1", "dr1")
	require.NoError(t, err)
}

func TestDeleteDynamicRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteDynamicRule(context.Background(), "cfg1", "dr1")
	require.Error(t, err)
}
