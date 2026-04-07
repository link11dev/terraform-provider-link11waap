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

func TestEdgeFunctionsDataSource_Metadata(t *testing.T) {
	d := NewEdgeFunctionsDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_edge_functions", resp.TypeName)
}

func TestEdgeFunctionsDataSource_Schema(t *testing.T) {
	d := NewEdgeFunctionsDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "edge_functions")
}

func TestEdgeFunctionsDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewEdgeFunctionsDataSource())
}

func TestEdgeFunctionsDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewEdgeFunctionsDataSource())
}

func TestEdgeFunctionsDataSource_Read_Success(t *testing.T) {
	d := NewEdgeFunctionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.EdgeFunction]{
			Total: 2,
			Items: []client.EdgeFunction{
				{ID: "ef1", Name: "func1", Description: "First", Code: "print('hello')", Phase: "request"},
				{ID: "ef2", Name: "func2", Description: "Second", Code: "return 200", Phase: "response"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestEdgeFunctionsDataSource_Read_Empty(t *testing.T) {
	d := NewEdgeFunctionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.EdgeFunction]{
			Total: 0,
			Items: []client.EdgeFunction{},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestEdgeFunctionsDataSource_Read_APIError(t *testing.T) {
	d := NewEdgeFunctionsDataSource()
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
