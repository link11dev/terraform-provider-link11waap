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

func TestListActions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/actions", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[Action]{
			Total: 1,
			Items: []Action{{ID: "a1", Name: "act1", Type: "block", Tags: []string{"a", "b"}}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	actions, err := c.ListActions(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, actions, 1)
	assert.Equal(t, "a1", actions[0].ID)
	assert.Equal(t, "block", actions[0].Type)
	assert.Equal(t, []string{"a", "b"}, actions[0].Tags)
}

func TestListActions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListActions(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListActions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListActions(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetAction_Success(t *testing.T) {
	status := 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/actions/a1", r.URL.Path)
		json.NewEncoder(w).Encode(Action{
			ID:          "a1",
			Name:        "act1",
			Description: "desc",
			Type:        "block",
			Tags:        []string{"x"},
			Params: &ActionParams{
				Content: "denied",
				Status:  &status,
				Headers: map[string]string{"X-Test": "1"},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	a, err := c.GetAction(context.Background(), "cfg1", "a1")
	require.NoError(t, err)
	assert.Equal(t, "a1", a.ID)
	assert.Equal(t, "block", a.Type)
	assert.Equal(t, "desc", a.Description)
	assert.Equal(t, []string{"x"}, a.Tags)
	require.NotNil(t, a.Params)
	assert.Equal(t, "denied", a.Params.Content)
	require.NotNil(t, a.Params.Status)
	assert.Equal(t, 403, *a.Params.Status)
	assert.Equal(t, "1", a.Params.Headers["X-Test"])
}

func TestGetAction_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetAction(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetAction_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetAction(context.Background(), "cfg1", "a1")
	require.Error(t, err)
}

func TestCreateAction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/actions/a1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var a Action
		require.NoError(t, json.Unmarshal(body, &a))
		assert.Equal(t, "a1", a.ID)
		assert.Equal(t, "monitor", a.Type)
		assert.Equal(t, []string{"t1"}, a.Tags)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"Successfully created entry"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateAction(context.Background(), "cfg1", "a1", &Action{
		ID:   "a1",
		Name: "act1",
		Type: "monitor",
		Tags: []string{"t1"},
	})
	require.NoError(t, err)
}

func TestCreateAction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateAction(context.Background(), "cfg1", "a1", &Action{})
	require.Error(t, err)
}

func TestUpdateAction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/conf/cfg1/actions/a1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateAction(context.Background(), "cfg1", "a1", &Action{ID: "a1"})
	require.NoError(t, err)
}

func TestUpdateAction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateAction(context.Background(), "cfg1", "a1", &Action{})
	require.Error(t, err)
}

func TestDeleteAction_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/actions/a1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteAction(context.Background(), "cfg1", "a1")
	require.NoError(t, err)
}

func TestDeleteAction_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteAction(context.Background(), "cfg1", "a1")
	require.Error(t, err)
}
