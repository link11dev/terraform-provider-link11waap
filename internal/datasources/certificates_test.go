package datasources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertificatesDataSource_Metadata(t *testing.T) {
	d := NewCertificatesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_certificates", resp.TypeName)
}

func TestCertificatesDataSource_Schema(t *testing.T) {
	d := NewCertificatesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "certificates")
}

func TestCertificatesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewCertificatesDataSource())
}

func TestCertificatesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewCertificatesDataSource())
}

func TestCertificatesDataSource_Read_ListAll(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Certificate]{
			Total: 1,
			Items: []client.Certificate{
				{
					ID:            "cert1",
					Name:          "test-cert",
					Subject:       "CN=test",
					Issuer:        "CN=issuer",
					SAN:           []string{"test.com", "*.test.com"},
					Expires:       "2025-12-31",
					Uploaded:      "2024-01-01",
					LEAutoRenew:   false,
					LEAutoReplace: false,
					Revoked:       false,
					Side:          "front",
					Links:         []client.ProviderLink{{Provider: "aws", Link: "https://aws.example.com/cert1", Region: "us-east-1"}},
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestCertificatesDataSource_Read_ByID(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.Certificate{
			ID:      "cert1",
			Name:    "test-cert",
			Subject: "CN=test",
			SAN:     []string{"test.com"},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cert1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestCertificatesDataSource_Read_ByID_APIError(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cert1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestCertificatesDataSource_Read_ByName(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Certificate]{
			Total: 2,
			Items: []client.Certificate{
				{ID: "cert1", Name: "first", SAN: []string{"a.com"}},
				{ID: "cert2", Name: "target-cert", SAN: []string{"b.com"},
					ProviderLinks: []client.ProviderLink{{Provider: "gcp", Link: "https://gcp.example.com/cert2", Region: "us-central1"}}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "target-cert"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestCertificatesDataSource_Read_ByName_NotFound(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Certificate]{
			Total: 1,
			Items: []client.Certificate{
				{ID: "cert1", Name: "other-cert", SAN: []string{"a.com"}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "missing-cert"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestCertificatesDataSource_Read_ByName_APIError(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "cert"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestCertificatesDataSource_Read_ListAll_APIError(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestCertificatesDataSource_Read_NoLinks(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Certificate]{
			Total: 1,
			Items: []client.Certificate{
				{ID: "cert1", Name: "test-cert", SAN: []string{"test.com"}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestCertificatesDataSource_Read_FallbackToProviderLinks(t *testing.T) {
	d := NewCertificatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Certificate]{
			Total: 1,
			Items: []client.Certificate{
				{
					ID:            "cert1",
					Name:          "test-cert",
					SAN:           []string{"test.com"},
					Links:         nil,
					ProviderLinks: []client.ProviderLink{{Provider: "aws", Link: "link1", Region: "us-east-1"}},
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}
