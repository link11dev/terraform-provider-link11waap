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

func TestLoadBalancerRegionsDataSource_Metadata(t *testing.T) {
	d := NewLoadBalancerRegionsDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_load_balancer_regions", resp.TypeName)
}

func TestLoadBalancerRegionsDataSource_Schema(t *testing.T) {
	d := NewLoadBalancerRegionsDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "city_codes")
	assert.Contains(t, resp.Schema.Attributes, "lbs")
}

func TestLoadBalancerRegionsDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewLoadBalancerRegionsDataSource())
}

func TestLoadBalancerRegionsDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewLoadBalancerRegionsDataSource())
}

func TestLoadBalancerRegionsDataSource_Read_Success(t *testing.T) {
	d := NewLoadBalancerRegionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.LoadBalancerRegions{
			CityCodes: map[string]string{"NYC": "New York", "AMS": "Amsterdam"},
			LBs: []client.LoadBalancerRegion{
				{
					ID:              "lb1",
					Name:            "load-balancer-1",
					Regions:         map[string]string{"ams": "automatic", "nyc": "manual"},
					UpstreamRegions: []string{"ams", "nyc"},
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

func TestLoadBalancerRegionsDataSource_Read_MultipleLBs(t *testing.T) {
	d := NewLoadBalancerRegionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.LoadBalancerRegions{
			CityCodes: map[string]string{"NYC": "New York"},
			LBs: []client.LoadBalancerRegion{
				{ID: "lb1", Name: "lb-1", Regions: map[string]string{"ams": "auto"}, UpstreamRegions: []string{"ams"}},
				{ID: "lb2", Name: "lb-2", Regions: map[string]string{"nyc": "manual"}, UpstreamRegions: []string{"nyc"}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestLoadBalancerRegionsDataSource_Read_APIError(t *testing.T) {
	d := NewLoadBalancerRegionsDataSource()
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
