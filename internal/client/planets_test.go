package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPlanet_ListResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/planets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode(ListResponse[Planet]{
			Total: 1,
			Items: []Planet{
				{
					ID:   "__default__",
					Name: "__default__",
					TrustedNets: []TrustedNet{
						{Source: "ip", Address: "1.2.3.4", Comment: "office"},
					},
				},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	planet, err := c.GetPlanet(context.Background(), "cfg1", "__default__")
	require.NoError(t, err)
	require.NotNil(t, planet)
	assert.Equal(t, "__default__", planet.ID)
	assert.Equal(t, "__default__", planet.Name)
	require.Len(t, planet.TrustedNets, 1)
	assert.Equal(t, "1.2.3.4", planet.TrustedNets[0].Address)
}

func TestGetPlanet_PlainObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/planets", r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		_ = json.NewEncoder(w).Encode(Planet{
			ID:   "__default__",
			Name: "__default__",
			TrustedNets: []TrustedNet{
				{Source: "ip", Address: "1.2.3.4", Comment: "office"},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	planet, err := c.GetPlanet(context.Background(), "cfg1", "__default__")
	require.NoError(t, err)
	require.NotNil(t, planet)
	assert.Equal(t, "__default__", planet.ID)
	assert.Equal(t, "__default__", planet.Name)
	require.Len(t, planet.TrustedNets, 1)
	assert.Equal(t, "1.2.3.4", planet.TrustedNets[0].Address)
}

func TestGetPlanet_NotFoundInList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(ListResponse[Planet]{
			Total: 1,
			Items: []Planet{
				{ID: "other", Name: "other"},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetPlanet(context.Background(), "cfg1", "__default__")
	require.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestGetPlanet_NotFoundPlainObject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(Planet{
			ID:   "other",
			Name: "other",
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetPlanet(context.Background(), "cfg1", "__default__")
	require.Error(t, err)
	assert.True(t, IsNotFoundError(err))
}

func TestGetPlanet_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"code":500,"message":"boom"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetPlanet(context.Background(), "cfg1", "__default__")
	require.Error(t, err)
}
