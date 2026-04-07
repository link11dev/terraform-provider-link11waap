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

func TestListCertificates_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/certificates", r.URL.Path)
		json.NewEncoder(w).Encode(ListResponse[Certificate]{
			Total: 1,
			Items: []Certificate{{ID: "cert1", Name: "my-cert"}},
		})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	certs, err := c.ListCertificates(context.Background(), "cfg1")
	require.NoError(t, err)
	require.Len(t, certs, 1)
	assert.Equal(t, "cert1", certs[0].ID)
}

func TestListCertificates_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListCertificates(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestListCertificates_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.ListCertificates(context.Background(), "cfg1")
	require.Error(t, err)
}

func TestGetCertificate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/conf/cfg1/certificates/cert1", r.URL.Path)
		json.NewEncoder(w).Encode(Certificate{ID: "cert1", Name: "my-cert"})
	}))
	defer server.Close()

	c := newTestClient(t, server)
	cert, err := c.GetCertificate(context.Background(), "cfg1", "cert1")
	require.NoError(t, err)
	assert.Equal(t, "cert1", cert.ID)
}

func TestGetCertificate_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetCertificate(context.Background(), "cfg1", "missing")
	require.Error(t, err)
}

func TestGetCertificate_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`bad`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	_, err := c.GetCertificate(context.Background(), "cfg1", "cert1")
	require.Error(t, err)
}

func TestCreateCertificate_Success_NoDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/conf/cfg1/certificates/cert1", r.URL.Path)
		assert.Empty(t, r.URL.Query().Get("domains"))
		body, _ := io.ReadAll(r.Body)
		var req CertificateCreateRequest
		require.NoError(t, json.Unmarshal(body, &req))
		assert.Equal(t, "cert1", req.ID)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateCertificate(context.Background(), "cfg1", "cert1", &CertificateCreateRequest{ID: "cert1"}, nil)
	require.NoError(t, err)
}

func TestCreateCertificate_Success_WithDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		domains := r.URL.Query()["domains"]
		assert.Contains(t, domains, "example.com")
		assert.Contains(t, domains, "test.com")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateCertificate(context.Background(), "cfg1", "cert1",
		&CertificateCreateRequest{ID: "cert1"},
		[]string{"example.com", "test.com"})
	require.NoError(t, err)
}

func TestCreateCertificate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"invalid cert"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.CreateCertificate(context.Background(), "cfg1", "cert1", &CertificateCreateRequest{}, nil)
	require.Error(t, err)
}

func TestUpdateCertificate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "true", r.URL.Query().Get("le_auto_renew"))
		assert.Equal(t, "false", r.URL.Query().Get("le_auto_replace"))
		assert.Empty(t, r.URL.Query().Get("replace_cert_id"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateCertificate(context.Background(), "cfg1", "cert1", true, false, "")
	require.NoError(t, err)
}

func TestUpdateCertificate_WithReplaceCertID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "cert2", r.URL.Query().Get("replace_cert_id"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateCertificate(context.Background(), "cfg1", "cert1", true, true, "cert2")
	require.NoError(t, err)
}

func TestUpdateCertificate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"code":400,"message":"bad"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.UpdateCertificate(context.Background(), "cfg1", "cert1", false, false, "")
	require.Error(t, err)
}

func TestDeleteCertificate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/conf/cfg1/certificates/cert1", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteCertificate(context.Background(), "cfg1", "cert1")
	require.NoError(t, err)
}

func TestDeleteCertificate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"code":500,"message":"error"}`))
	}))
	defer server.Close()

	c := newTestClient(t, server)
	err := c.DeleteCertificate(context.Background(), "cfg1", "cert1")
	require.Error(t, err)
}
