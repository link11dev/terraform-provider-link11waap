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

func TestGlobalFiltersDataSource_Metadata(t *testing.T) {
	d := NewGlobalFiltersDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_global_filters", resp.TypeName)
}

func TestGlobalFiltersDataSource_Schema(t *testing.T) {
	d := NewGlobalFiltersDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "global_filters")
}

func TestGlobalFiltersDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewGlobalFiltersDataSource())
}

func TestGlobalFiltersDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewGlobalFiltersDataSource())
}

func TestGlobalFiltersDataSource_Read_Success(t *testing.T) {
	d := NewGlobalFiltersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.GlobalFilter]{
			Total: 1,
			Items: []client.GlobalFilter{
				{
					ID:          "gf1",
					Name:        "test-filter",
					Description: "Test Filter",
					Source:      "https://example.com/filter.json",
					Mdate:       "2024-01-01",
					Active:      true,
					Tags:        []string{"tag1", "tag2"},
					Action:      "action-monitor",
					Rule:        map[string]interface{}{"relation": "OR", "entries": []interface{}{}},
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

func TestGlobalFiltersDataSource_Read_EmptySlices(t *testing.T) {
	d := NewGlobalFiltersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.GlobalFilter]{
			Total: 1,
			Items: []client.GlobalFilter{
				{
					ID:     "gf1",
					Name:   "test-filter",
					Source: "https://example.com/filter.json",
					Active: true,
					// nil tags - should produce null lists
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

func TestGlobalFiltersDataSource_Read_APIError(t *testing.T) {
	d := NewGlobalFiltersDataSource()
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

func TestGlobalFiltersDataSource_Read_MultipleFilters(t *testing.T) {
	d := NewGlobalFiltersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.GlobalFilter]{
			Total: 2,
			Items: []client.GlobalFilter{
				{ID: "gf1", Name: "first", Source: "https://example.com/1.json", Active: true, Tags: []string{"t1"}, Action: "action-monitor"},
				{ID: "gf2", Name: "second", Source: "https://example.com/2.json", Active: false, Action: "action-skip", Rule: map[string]interface{}{"key": "value"}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

// --- Singular GlobalFilterDataSource tests ---

func TestGlobalFilterDataSource_Metadata(t *testing.T) {
	d := NewGlobalFilterDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_global_filter", resp.TypeName)
}

func TestGlobalFilterDataSource_Schema(t *testing.T) {
	d := NewGlobalFilterDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "name")
}

func TestGlobalFilterDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewGlobalFilterDataSource())
}

func TestGlobalFilterDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewGlobalFilterDataSource())
}

func TestGlobalFilterDataSource_Read_Success(t *testing.T) {
	d := NewGlobalFilterDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.GlobalFilter]{
			Total: 2,
			Items: []client.GlobalFilter{
				{
					ID:          "gf1",
					Name:        "test-filter",
					Description: "Test Filter",
					Source:      "https://example.com/filter.json",
					Mdate:       "2024-01-01",
					Active:      true,
					Tags:        []string{"tag1", "tag2"},
					Action:      "action-monitor",
					Rule:        map[string]interface{}{"relation": "OR", "entries": []interface{}{}},
				},
				{
					ID:     "gf2",
					Name:   "other-filter",
					Source: "https://example.com/other.json",
					Active: false,
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test-filter"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestGlobalFilterDataSource_Read_NotFound(t *testing.T) {
	d := NewGlobalFilterDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.GlobalFilter]{
			Total: 1,
			Items: []client.GlobalFilter{
				{
					ID:     "gf1",
					Name:   "existing-filter",
					Source: "https://example.com/filter.json",
					Active: true,
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "nonexistent"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestGlobalFilterDataSource_Read_APIError(t *testing.T) {
	d := NewGlobalFilterDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test-filter"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
