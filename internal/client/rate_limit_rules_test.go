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

func TestListRateLimitRules_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/rate-limit-rules", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[RateLimitRule]{
			Total: 1,
			Items: []RateLimitRule{{ID: "rl1", Name: "rate1", Timeframe: 60, Threshold: 100}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	rules, err := c.ListRateLimitRules(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, "rl1", rules[0].ID)
}

func TestListRateLimitRules_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListRateLimitRules(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListRateLimitRules_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListRateLimitRules(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetRateLimitRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/rate-limit-rules/rl1", r.URL.Path)
		json.NewEncoder(w).Encode(RateLimitRule{ID: "rl1", Name: "rate1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	rule, err := c.GetRateLimitRule(context.Background(), "cfg1", "rl1")
	require.NoError(t, err)
	assert.Equal(t, "rl1", rule.ID)
}

func TestGetRateLimitRule_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetRateLimitRule(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetRateLimitRule_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetRateLimitRule(context.Background(), "cfg1", "rl1")
	require.Error(t, err)
}

func TestCreateRateLimitRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var rule RateLimitRule
		require.NoError(t, json.Unmarshal(body, &rule))
		assert.Equal(t, "rl1", rule.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateRateLimitRule(context.Background(), "cfg1", "rl1", &RateLimitRule{ID: "rl1", Name: "rate1"})
	require.NoError(t, err)
}

func TestCreateRateLimitRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateRateLimitRule(context.Background(), "cfg1", "rl1", &RateLimitRule{})
	require.Error(t, err)
}

func TestUpdateRateLimitRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateRateLimitRule(context.Background(), "cfg1", "rl1", &RateLimitRule{ID: "rl1"})
	require.NoError(t, err)
}

func TestUpdateRateLimitRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateRateLimitRule(context.Background(), "cfg1", "rl1", &RateLimitRule{})
	require.Error(t, err)
}

func TestDeleteRateLimitRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteRateLimitRule(context.Background(), "cfg1", "rl1")
	require.NoError(t, err)
}

func TestDeleteRateLimitRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteRateLimitRule(context.Background(), "cfg1", "rl1")
	require.Error(t, err)
}
