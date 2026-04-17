package datasources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlanetTrustedNetsDataSource(t *testing.T) {
	d := NewPlanetTrustedNetsDataSource()
	require.NotNil(t, d)
	_, ok := d.(*PlanetTrustedNetsDataSource)
	assert.True(t, ok)
}

func TestPlanetTrustedNetsDataSource_Metadata(t *testing.T) {
	d := NewPlanetTrustedNetsDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_planet_trusted_nets", resp.TypeName)
}

func TestPlanetTrustedNetsDataSource_Schema(t *testing.T) {
	d := NewPlanetTrustedNetsDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)

	for _, name := range []string{"config_id", "id", "name", "trusted_nets"} {
		_, ok := resp.Schema.Attributes[name]
		assert.True(t, ok, "expected attribute %q", name)
	}

	trusted, ok := resp.Schema.Attributes["trusted_nets"].(schema.ListNestedAttribute)
	require.True(t, ok, "trusted_nets should be ListNestedAttribute")
	for _, name := range []string{"source", "address", "gf_id", "comment"} {
		_, ok := trusted.NestedObject.Attributes[name]
		assert.True(t, ok, "expected nested attribute %q", name)
	}
}

func TestPlanetTrustedNetsDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewPlanetTrustedNetsDataSource())
}

func TestPlanetTrustedNetsDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewPlanetTrustedNetsDataSource())
}

func TestPlanetTrustedNetsDataSource_Read_Success(t *testing.T) {
	d := NewPlanetTrustedNetsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(client.ListResponse[client.Planet]{
			Total: 1,
			Items: []client.Planet{
				{
					ID:   "__default__",
					Name: "__default__",
					TrustedNets: []client.TrustedNet{
						{Source: "ip", Address: "1.2.3.4", Comment: "office"},
						{Source: "global_filter", GfID: "gf1", Comment: "gf"},
					},
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

func TestPlanetTrustedNetsDataSource_Read_EmptyTrustedNets(t *testing.T) {
	d := NewPlanetTrustedNetsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(client.ListResponse[client.Planet]{
			Total: 1,
			Items: []client.Planet{
				{ID: "__default__", Name: "__default__", TrustedNets: []client.TrustedNet{}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestPlanetTrustedNetsDataSource_Read_APIError(t *testing.T) {
	d := NewPlanetTrustedNetsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
