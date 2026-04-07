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

func TestListLoadBalancers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/load-balancers", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[LoadBalancer]{
			Total: 1,
			Items: []LoadBalancer{{Name: "lb1", Provider: "aws", Region: "us-east-1"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	lbs, err := c.ListLoadBalancers(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, lbs, 1)
	assert.Equal(t, "lb1", lbs[0].Name)
}

func TestListLoadBalancers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListLoadBalancers(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListLoadBalancers_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListLoadBalancers(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetLoadBalancerRegions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/load-balancers/regions", r.URL.Path)
		json.NewEncoder(w).Encode(LoadBalancerRegions{
			CityCodes: map[string]string{"NYC": "New York"},
			LBs: []LoadBalancerRegion{
				{ID: "lb1", Name: "lb1", Regions: map[string]string{"us-east-1": "NYC"}},
			},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	regions, err := c.GetLoadBalancerRegions(context.Background(), "cfg1")
	require.NoError(t, err)
	assert.Len(t, regions.LBs, 1)
	assert.Equal(t, "lb1", regions.LBs[0].ID)
}

func TestGetLoadBalancerRegions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetLoadBalancerRegions(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetLoadBalancerRegions_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetLoadBalancerRegions(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestUpdateLoadBalancerRegions_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/load-balancers/regions", r.URL.Path)
		body, _ := io.ReadAll(r.Body)
		var req LoadBalancerRegionsUpdateRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Len(t, req.LBs, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateLoadBalancerRegions(context.Background(), "cfg1", &LoadBalancerRegionsUpdateRequest{
		LBs: []LoadBalancerRegionUpdate{{ID: "lb1", Regions: map[string]string{"us-east-1": "NYC"}}},
	})
	require.NoError(t, err)
}

func TestUpdateLoadBalancerRegions_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateLoadBalancerRegions(context.Background(), "cfg1", &LoadBalancerRegionsUpdateRequest{})
	require.Error(t, err)
}

func TestAttachCertificateToLoadBalancer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Contains(t, r.URL.Path, "/conf/cfg1/load-balancers/lb1/certificates/cert1")
		assert.Equal(t, "aws", r.URL.Query().Get("provider"))
		assert.Equal(t, "us-east-1", r.URL.Query().Get("region"))
		assert.Equal(t, "listener1", r.URL.Query().Get("listener"))
		assert.Equal(t, "443", r.URL.Query().Get("listener-port"))
		assert.Equal(t, "true", r.URL.Query().Get("default"))
		assert.Equal(t, "false", r.URL.Query().Get("elbv2"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.AttachCertificateToLoadBalancer(context.Background(), "cfg1", "lb1", "cert1", AttachCertificateOptions{
		Provider:     "aws",
		Region:       "us-east-1",
		Listener:     "listener1",
		ListenerPort: 443,
		IsDefault:    true,
		ELBv2:        false,
	})
	require.NoError(t, err)
}

func TestAttachCertificateToLoadBalancer_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.AttachCertificateToLoadBalancer(context.Background(), "cfg1", "lb1", "cert1", AttachCertificateOptions{})
	require.Error(t, err)
}

func TestDetachCertificateFromLoadBalancer_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/load-balancers/lb1/certificates", r.URL.Path)
		assert.Equal(t, "aws", r.URL.Query().Get("provider"))
		assert.Equal(t, "us-east-1", r.URL.Query().Get("region"))
		assert.Equal(t, "cert1", r.URL.Query().Get("certificate-id"))
		assert.Equal(t, "listener1", r.URL.Query().Get("listener"))
		assert.Equal(t, "443", r.URL.Query().Get("listener-port"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DetachCertificateFromLoadBalancer(context.Background(), "cfg1", "lb1", DetachCertificateOptions{
		Provider:      "aws",
		Region:        "us-east-1",
		CertificateID: "cert1",
		Listener:      "listener1",
		ListenerPort:  "443",
		ELBv2:         false,
	})
	require.NoError(t, err)
}

func TestDetachCertificateFromLoadBalancer_MinimalOpts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "aws", r.URL.Query().Get("provider"))
		assert.Equal(t, "us-east-1", r.URL.Query().Get("region"))
		// These should not be present when empty
		assert.Empty(t, r.URL.Query().Get("certificate-id"))
		assert.Empty(t, r.URL.Query().Get("listener"))
		assert.Empty(t, r.URL.Query().Get("listener-port"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DetachCertificateFromLoadBalancer(context.Background(), "cfg1", "lb1", DetachCertificateOptions{
		Provider: "aws",
		Region:   "us-east-1",
	})
	require.NoError(t, err)
}

func TestDetachCertificateFromLoadBalancer_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DetachCertificateFromLoadBalancer(context.Background(), "cfg1", "lb1", DetachCertificateOptions{})
	require.Error(t, err)
}
