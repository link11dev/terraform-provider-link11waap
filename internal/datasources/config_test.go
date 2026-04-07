package datasources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDataSource_Metadata(t *testing.T) {
	d := NewConfigDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_config", resp.TypeName)
}

func TestConfigDataSource_Schema(t *testing.T) {
	d := NewConfigDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "description")
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "version")
	assert.Contains(t, resp.Schema.Attributes, "date")
}

func TestConfigDataSource_Configure_InvalidType(t *testing.T) {
	d := NewConfigDataSource()
	testDSConfigureWithInvalidType(t, d)
}

func TestConfigDataSource_Configure_Nil(t *testing.T) {
	d := NewConfigDataSource()
	testDSConfigureWithNil(t, d)
}

func TestConfigDataSource_Read_Success(t *testing.T) {
	d := NewConfigDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.WAAPConfig]{
			Total: 1,
			Items: []client.WAAPConfig{
				{
					ID:          "cfg1",
					Description: "My Config",
					Version:     "v1",
					Date:        time.Date(2024, 6, 15, 10, 30, 0, 0, time.UTC),
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestConfigDataSource_Read_WithDescriptionFilter(t *testing.T) {
	d := NewConfigDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.WAAPConfig]{
			Total: 2,
			Items: []client.WAAPConfig{
				{ID: "cfg1", Description: "First Config", Version: "v1"},
				{ID: "cfg2", Description: "Second Config", Version: "v2"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"description": tftypes.NewValue(tftypes.String, "Second Config"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestConfigDataSource_Read_DescriptionFilterNotFound(t *testing.T) {
	d := NewConfigDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.WAAPConfig]{
			Total: 1,
			Items: []client.WAAPConfig{
				{ID: "cfg1", Description: "Some Config", Version: "v1"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"description": tftypes.NewValue(tftypes.String, "Not Existing"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestConfigDataSource_Read_NoConfigs(t *testing.T) {
	d := NewConfigDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.WAAPConfig]{
			Total: 0,
			Items: []client.WAAPConfig{},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestConfigDataSource_Read_APIError(t *testing.T) {
	d := NewConfigDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"internal error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestConfigDataSource_Read_ZeroDate(t *testing.T) {
	d := NewConfigDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.WAAPConfig]{
			Total: 1,
			Items: []client.WAAPConfig{
				{ID: "cfg1", Description: "Config", Version: "v1"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}
