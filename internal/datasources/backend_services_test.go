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

func TestBackendServicesDataSource_Metadata(t *testing.T) {
	d := NewBackendServicesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_backend_services", resp.TypeName)
}

func TestBackendServicesDataSource_Schema(t *testing.T) {
	d := NewBackendServicesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "backend_services")
}

func TestBackendServicesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewBackendServicesDataSource())
}

func TestBackendServicesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewBackendServicesDataSource())
}

func TestBackendServicesDataSource_Read_ListAll(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.BackendService]{
			Total: 1,
			Items: []client.BackendService{
				{
					ID:            "bs1",
					Name:          "test-bs",
					Description:   "Test backend",
					HTTP11:        true,
					TransportMode: "default",
					Sticky:        "none",
					LeastConn:     false,
					BackHosts: []client.BackendHost{
						{
							Host:         "backend.example.com",
							HTTPPorts:    []int{80},
							HTTPSPorts:   []int{443},
							Weight:       1,
							MaxFails:     3,
							FailTimeout:  10,
							Down:         false,
							MonitorState: "up",
							Backup:       false,
						},
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

func TestBackendServicesDataSource_Read_ByID(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// GetBackendService returns a single object
		json.NewEncoder(w).Encode(client.BackendService{
			ID:            "bs1",
			Name:          "test-bs",
			TransportMode: "default",
			Sticky:        "none",
			BackHosts:     nil,
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "bs1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestBackendServicesDataSource_Read_ByID_APIError(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "bs1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendServicesDataSource_Read_ByName(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.BackendService]{
			Total: 2,
			Items: []client.BackendService{
				{ID: "bs1", Name: "first-bs", TransportMode: "default", Sticky: "none"},
				{ID: "bs2", Name: "target-bs", TransportMode: "default", Sticky: "none", BackHosts: []client.BackendHost{
					{Host: "h1.example.com", HTTPPorts: []int{80, 8080}, HTTPSPorts: []int{443}, Weight: 1, MaxFails: 3, FailTimeout: 10},
				}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "target-bs"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestBackendServicesDataSource_Read_ByName_NotFound(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.BackendService]{
			Total: 1,
			Items: []client.BackendService{
				{ID: "bs1", Name: "other-bs", TransportMode: "default", Sticky: "none"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "missing-bs"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendServicesDataSource_Read_ByName_APIError(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test-bs"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendServicesDataSource_Read_ListAll_APIError(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBackendServicesDataSource_Read_NullBackHosts(t *testing.T) {
	d := NewBackendServicesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.BackendService]{
			Total: 1,
			Items: []client.BackendService{
				{ID: "bs1", Name: "test-bs", TransportMode: "default", Sticky: "none", BackHosts: nil},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestDsBackendHostAttrTypes(t *testing.T) {
	attrTypes := dsBackendHostAttrTypes()
	assert.Contains(t, attrTypes, "host")
	assert.Contains(t, attrTypes, "http_ports")
	assert.Contains(t, attrTypes, "https_ports")
	assert.Contains(t, attrTypes, "weight")
	assert.Contains(t, attrTypes, "max_fails")
	assert.Contains(t, attrTypes, "fail_timeout")
	assert.Contains(t, attrTypes, "down")
	assert.Contains(t, attrTypes, "monitor_state")
	assert.Contains(t, attrTypes, "backup")
}

func TestDsIntSliceToInt64(t *testing.T) {
	result := dsIntSliceToInt64([]int{80, 443, 8080})
	assert.Equal(t, []int64{80, 443, 8080}, result)
}

func TestDsIntSliceToInt64_Empty(t *testing.T) {
	result := dsIntSliceToInt64([]int{})
	assert.Equal(t, []int64{}, result)
}
