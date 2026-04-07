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

func TestLoadBalancersDataSource_Metadata(t *testing.T) {
	d := NewLoadBalancersDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_load_balancers", resp.TypeName)
}

func TestLoadBalancersDataSource_Schema(t *testing.T) {
	d := NewLoadBalancersDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "load_balancers")
}

func TestLoadBalancersDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewLoadBalancersDataSource())
}

func TestLoadBalancersDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewLoadBalancersDataSource())
}

func TestLoadBalancersDataSource_Read_Success(t *testing.T) {
	d := NewLoadBalancersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
			Total: 1,
			Items: []client.LoadBalancer{
				{
					Name:               "lb1",
					Provider:           "aws",
					Region:             "us-east-1",
					DNSName:            "lb1.example.com",
					ListenerName:       "listener1",
					ListenerPort:       443,
					LoadBalancerType:   "application",
					MaxCertificates:    25,
					DefaultCertificate: "cert1",
					Certificates:       []string{"cert1", "cert2"},
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

func TestLoadBalancersDataSource_Read_MultipleLBs(t *testing.T) {
	d := NewLoadBalancersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
			Total: 2,
			Items: []client.LoadBalancer{
				{Name: "lb1", Provider: "aws", Region: "us-east-1", Certificates: []string{"c1"}},
				{Name: "lb2", Provider: "gcp", Region: "us-central1", Certificates: []string{"c2", "c3"}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestLoadBalancersDataSource_Read_APIError(t *testing.T) {
	d := NewLoadBalancersDataSource()
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
