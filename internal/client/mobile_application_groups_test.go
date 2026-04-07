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

func TestListMobileApplicationGroups_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/mobile-application-groups", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[MobileApplicationGroup]{
			Total: 1,
			Items: []MobileApplicationGroup{{ID: "mag1", Name: "app-group1"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	groups, err := c.ListMobileApplicationGroups(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, groups, 1)
	assert.Equal(t, "mag1", groups[0].ID)
}

func TestListMobileApplicationGroups_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListMobileApplicationGroups(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListMobileApplicationGroups_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListMobileApplicationGroups(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetMobileApplicationGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/mobile-application-groups/mag1", r.URL.Path)
		json.NewEncoder(w).Encode(MobileApplicationGroup{ID: "mag1", Name: "app-group1"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	mag, err := c.GetMobileApplicationGroup(context.Background(), "cfg1", "mag1")
	require.NoError(t, err)
	assert.Equal(t, "mag1", mag.ID)
}

func TestGetMobileApplicationGroup_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetMobileApplicationGroup(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetMobileApplicationGroup_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetMobileApplicationGroup(context.Background(), "cfg1", "mag1")
	require.Error(t, err)
}

func TestCreateMobileApplicationGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		var mag MobileApplicationGroup
		require.NoError(t, json.Unmarshal(body, &mag))
		assert.Equal(t, "mag1", mag.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateMobileApplicationGroup(context.Background(), "cfg1", "mag1", &MobileApplicationGroup{ID: "mag1", Name: "app-group1"})
	require.NoError(t, err)
}

func TestCreateMobileApplicationGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateMobileApplicationGroup(context.Background(), "cfg1", "mag1", &MobileApplicationGroup{})
	require.Error(t, err)
}

func TestUpdateMobileApplicationGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateMobileApplicationGroup(context.Background(), "cfg1", "mag1", &MobileApplicationGroup{ID: "mag1"})
	require.NoError(t, err)
}

func TestUpdateMobileApplicationGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateMobileApplicationGroup(context.Background(), "cfg1", "mag1", &MobileApplicationGroup{})
	require.Error(t, err)
}

func TestDeleteMobileApplicationGroup_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteMobileApplicationGroup(context.Background(), "cfg1", "mag1")
	require.NoError(t, err)
}

func TestDeleteMobileApplicationGroup_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteMobileApplicationGroup(context.Background(), "cfg1", "mag1")
	require.Error(t, err)
}
