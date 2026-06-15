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

func TestListContentFilterRules_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/content-filter-rules", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[ContentFilterRule]{
			Total: 1,
			Items: []ContentFilterRule{{ID: "cfr1", Name: "rule1", Risk: 3, Tags: []string{"a", "b"}}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	rules, err := c.ListContentFilterRules(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, rules, 1)
	assert.Equal(t, "cfr1", rules[0].ID)
	assert.Equal(t, 3, rules[0].Risk)
	assert.Equal(t, []string{"a", "b"}, rules[0].Tags)
}

func TestListContentFilterRules_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListContentFilterRules(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListContentFilterRules_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListContentFilterRules(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetContentFilterRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/content-filter-rules/cfr1", r.URL.Path)
		json.NewEncoder(w).Encode(ContentFilterRule{
			ID:          "cfr1",
			Name:        "rule1",
			Msg:         "blocked",
			Operand:     ".*",
			Category:    "cat",
			Subcategory: "sub",
			Risk:        5,
			Description: "desc",
			Tags:        []string{"x"},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	rule, err := c.GetContentFilterRule(context.Background(), "cfg1", "cfr1")
	require.NoError(t, err)
	assert.Equal(t, "cfr1", rule.ID)
	assert.Equal(t, "blocked", rule.Msg)
	assert.Equal(t, ".*", rule.Operand)
	assert.Equal(t, "cat", rule.Category)
	assert.Equal(t, "sub", rule.Subcategory)
	assert.Equal(t, 5, rule.Risk)
	assert.Equal(t, "desc", rule.Description)
	assert.Equal(t, []string{"x"}, rule.Tags)
}

func TestGetContentFilterRule_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetContentFilterRule(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetContentFilterRule_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetContentFilterRule(context.Background(), "cfg1", "cfr1")
	require.Error(t, err)
}

func TestCreateContentFilterRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/content-filter-rules/cfr1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var rule ContentFilterRule
		require.NoError(t, json.Unmarshal(body, &rule))
		assert.Equal(t, "cfr1", rule.ID)
		assert.Equal(t, 2, rule.Risk)
		assert.Equal(t, []string{"t1"}, rule.Tags)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateContentFilterRule(context.Background(), "cfg1", "cfr1", &ContentFilterRule{
		ID:          "cfr1",
		Name:        "rule1",
		Msg:         "blocked",
		Operand:     ".*",
		Category:    "cat",
		Subcategory: "sub",
		Risk:        2,
		Tags:        []string{"t1"},
	})
	require.NoError(t, err)
}

func TestCreateContentFilterRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateContentFilterRule(context.Background(), "cfg1", "cfr1", &ContentFilterRule{})
	require.Error(t, err)
}

func TestUpdateContentFilterRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/conf/cfg1/content-filter-rules/cfr1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateContentFilterRule(context.Background(), "cfg1", "cfr1", &ContentFilterRule{ID: "cfr1"})
	require.NoError(t, err)
}

func TestUpdateContentFilterRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateContentFilterRule(context.Background(), "cfg1", "cfr1", &ContentFilterRule{})
	require.Error(t, err)
}

func TestDeleteContentFilterRule_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/content-filter-rules/cfr1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteContentFilterRule(context.Background(), "cfg1", "cfr1")
	require.NoError(t, err)
}

func TestDeleteContentFilterRule_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteContentFilterRule(context.Background(), "cfg1", "cfr1")
	require.Error(t, err)
}
