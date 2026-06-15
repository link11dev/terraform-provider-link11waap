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

func sampleContentFilterProfile() ContentFilterProfile {
	return ContentFilterProfile{
		ID:             "cfp1",
		Name:           "profile1",
		Description:    "desc",
		IgnoreAlphanum: true,
		MaskingSeed:    "seed",
		ContentType:    []string{"application/json"},
		GraphqlPath:    "$.query",
		IgnoreBody:     false,
		Active:         []string{"a"},
		Report:         []string{"r"},
		Ignore:         []string{"i"},
		Tags:           []string{"t1"},
		Action:         "act1",
		Args: ContentFilterProfileSection{
			MaxCount: 10, MaxLength: 100, EnableMaxCount: true, EnableMaxLength: true,
			Names: []ContentFilterEntryMatch{{Key: "k", Mask: true, Active: true}},
		},
		Headers:     ContentFilterProfileSection{MaxCount: 1, MaxLength: 1},
		Cookies:     ContentFilterProfileSection{MaxCount: 1, MaxLength: 1},
		Path:        ContentFilterProfileSection{MaxCount: 1, MaxLength: 1},
		URL:         ContentFilterURLSection{},
		AllSections: ContentFilterProfileSection{},
		Decoding:    ContentFilterDecoding{Base64: true},
	}
}

func TestListContentFilterProfiles_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/content-filter-profiles", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[ContentFilterProfile]{
			Total: 1,
			Items: []ContentFilterProfile{sampleContentFilterProfile()},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	profiles, err := c.ListContentFilterProfiles(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, profiles, 1)
	assert.Equal(t, "cfp1", profiles[0].ID)
	assert.True(t, profiles[0].IgnoreAlphanum)
	assert.Equal(t, 10, profiles[0].Args.MaxCount)
	require.Len(t, profiles[0].Args.Names, 1)
	assert.Equal(t, "k", profiles[0].Args.Names[0].Key)
}

func TestListContentFilterProfiles_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListContentFilterProfiles(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListContentFilterProfiles_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListContentFilterProfiles(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetContentFilterProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/content-filter-profiles/cfp1", r.URL.Path)
		json.NewEncoder(w).Encode(sampleContentFilterProfile())
	}))
	defer server.Close()

	c := newTestClient(t, server)
	p, err := c.GetContentFilterProfile(context.Background(), "cfg1", "cfp1")
	require.NoError(t, err)
	assert.Equal(t, "cfp1", p.ID)
	assert.Equal(t, "seed", p.MaskingSeed)
	assert.True(t, p.Decoding.Base64)
}

func TestGetContentFilterProfile_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetContentFilterProfile(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetContentFilterProfile_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetContentFilterProfile(context.Background(), "cfg1", "cfp1")
	require.Error(t, err)
}

func TestCreateContentFilterProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/content-filter-profiles/cfp1", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var p ContentFilterProfile
		require.NoError(t, json.Unmarshal(body, &p))
		assert.Equal(t, "cfp1", p.ID)
		assert.Equal(t, "seed", p.MaskingSeed)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	p := sampleContentFilterProfile()
	err := c.CreateContentFilterProfile(context.Background(), "cfg1", "cfp1", &p)
	require.NoError(t, err)
}

func TestCreateContentFilterProfile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateContentFilterProfile(context.Background(), "cfg1", "cfp1", &ContentFilterProfile{})
	require.Error(t, err)
}

func TestUpdateContentFilterProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/conf/cfg1/content-filter-profiles/cfp1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateContentFilterProfile(context.Background(), "cfg1", "cfp1", &ContentFilterProfile{ID: "cfp1"})
	require.NoError(t, err)
}

func TestUpdateContentFilterProfile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateContentFilterProfile(context.Background(), "cfg1", "cfp1", &ContentFilterProfile{})
	require.Error(t, err)
}

func TestDeleteContentFilterProfile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/content-filter-profiles/cfp1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteContentFilterProfile(context.Background(), "cfg1", "cfp1")
	require.NoError(t, err)
}

func TestDeleteContentFilterProfile_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteContentFilterProfile(context.Background(), "cfg1", "cfp1")
	require.Error(t, err)
}
