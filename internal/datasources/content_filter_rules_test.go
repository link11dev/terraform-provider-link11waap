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

func TestContentFilterRulesDataSource_Metadata(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_content_filter_rules", resp.TypeName)
}

func TestContentFilterRulesDataSource_Schema(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "content_filter_rules")
}

func TestContentFilterRulesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewContentFilterRulesDataSource())
}

func TestContentFilterRulesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewContentFilterRulesDataSource())
}

func TestContentFilterRulesDataSource_Read_Success(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterRule]{
			Total: 2,
			Items: []client.ContentFilterRule{
				{ID: "cfr1", Name: "rule1", Msg: "m1", Operand: ".*", Category: "c", Subcategory: "s", Risk: 1, Description: "First", Tags: []string{"a"}},
				{ID: "cfr2", Name: "rule2", Msg: "m2", Operand: ".+", Category: "c", Subcategory: "s", Risk: 5, Description: "Second"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterRulesDataSource_Read_Empty(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterRule]{
			Total: 0,
			Items: []client.ContentFilterRule{},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterRulesDataSource_Read_APIError(t *testing.T) {
	d := NewContentFilterRulesDataSource()
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

func TestContentFilterRulesDataSource_Read_ByName(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterRule]{
			Total: 2,
			Items: []client.ContentFilterRule{
				{ID: "cfr1", Name: "Block Forbidden", Msg: "blocked", Operand: ".*", Category: "c", Subcategory: "s", Risk: 3, Tags: []string{"block"}},
				{ID: "cfr2", Name: "Other Rule", Msg: "other", Operand: ".+", Category: "c", Subcategory: "s", Risk: 1},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "Block Forbidden"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterRulesDataSource_Read_ByName_NotFound(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterRule]{
			Total: 1,
			Items: []client.ContentFilterRule{
				{ID: "cfr1", Name: "Other Rule", Msg: "other", Operand: ".+", Category: "c", Subcategory: "s", Risk: 1},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "Block Forbidden"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestContentFilterRulesDataSource_Read_ByName_APIError(t *testing.T) {
	d := NewContentFilterRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "Block Forbidden"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
