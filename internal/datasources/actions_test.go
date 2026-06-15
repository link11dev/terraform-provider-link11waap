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

func TestActionsDataSource_Metadata(t *testing.T) {
	d := NewActionsDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_actions", resp.TypeName)
}

func TestActionsDataSource_Schema(t *testing.T) {
	d := NewActionsDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "actions")
}

func TestActionsDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewActionsDataSource())
}

func TestActionsDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewActionsDataSource())
}

func TestActionsDataSource_Read_ListAll(t *testing.T) {
	status := 403
	d := NewActionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Action]{
			Total: 2,
			Items: []client.Action{
				{ID: "a1", Name: "act1", Type: "block", Description: "First", Tags: []string{"a"}, Params: &client.ActionParams{Content: "no", Status: &status, Headers: map[string]string{"X-A": "1"}}},
				{ID: "a2", Name: "act2", Type: "monitor", Description: "Second"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestActionsDataSource_Read_ByID(t *testing.T) {
	d := NewActionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4.3/conf/cfg1/actions/a1", r.URL.Path)
		json.NewEncoder(w).Encode(client.Action{ID: "a1", Name: "act1", Type: "block"})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "a1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestActionsDataSource_Read_ByName(t *testing.T) {
	d := NewActionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Action]{
			Total: 2,
			Items: []client.Action{
				{ID: "a1", Name: "act1", Type: "block"},
				{ID: "a2", Name: "act2", Type: "monitor"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "act2"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestActionsDataSource_Read_ByName_NotFound(t *testing.T) {
	d := NewActionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Action]{
			Total: 1,
			Items: []client.Action{{ID: "a1", Name: "act1", Type: "block"}},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "missing"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestActionsDataSource_Read_Empty(t *testing.T) {
	d := NewActionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.Action]{
			Total: 0,
			Items: []client.Action{},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestActionsDataSource_Read_APIError(t *testing.T) {
	d := NewActionsDataSource()
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

func TestActionsDataSource_Read_ByID_APIError(t *testing.T) {
	d := NewActionsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
