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

func TestMobileApplicationGroupsDataSource_Metadata(t *testing.T) {
	d := NewMobileApplicationGroupsDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_mobile_application_groups", resp.TypeName)
}

func TestMobileApplicationGroupsDataSource_Schema(t *testing.T) {
	d := NewMobileApplicationGroupsDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "mobile_application_groups")
}

func TestMobileApplicationGroupsDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewMobileApplicationGroupsDataSource())
}

func TestMobileApplicationGroupsDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewMobileApplicationGroupsDataSource())
}

func TestMobileApplicationGroupsDataSource_Read_Success(t *testing.T) {
	d := NewMobileApplicationGroupsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.MobileApplicationGroup]{
			Total: 1,
			Items: []client.MobileApplicationGroup{
				{
					ID:          "mag1",
					Name:        "test-mag",
					Description: "Test MAG",
					UIDHeader:   "X-UID",
					Grace:       "5m",
					ActiveConfig: []client.ActiveConfig{
						{Active: true, JSON: "{}", Name: "config1"},
					},
					Signatures: []client.Signature{
						{Active: true, Hash: "abc123", Name: "sig1"},
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

func TestMobileApplicationGroupsDataSource_Read_NullActiveConfigAndSignatures(t *testing.T) {
	d := NewMobileApplicationGroupsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.MobileApplicationGroup]{
			Total: 1,
			Items: []client.MobileApplicationGroup{
				{
					ID:   "mag1",
					Name: "test-mag",
					// nil ActiveConfig and Signatures
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

func TestMobileApplicationGroupsDataSource_Read_APIError(t *testing.T) {
	d := NewMobileApplicationGroupsDataSource()
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

func TestMobileApplicationGroupsDataSource_Read_MultipleGroups(t *testing.T) {
	d := NewMobileApplicationGroupsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.MobileApplicationGroup]{
			Total: 2,
			Items: []client.MobileApplicationGroup{
				{ID: "mag1", Name: "first", ActiveConfig: []client.ActiveConfig{{Active: true, JSON: "{}", Name: "c1"}}, Signatures: []client.Signature{{Active: true, Hash: "h1", Name: "s1"}}},
				{ID: "mag2", Name: "second"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestDsActiveConfigAttrTypes(t *testing.T) {
	attrTypes := dsActiveConfigAttrTypes()
	assert.Contains(t, attrTypes, "active")
	assert.Contains(t, attrTypes, "json")
	assert.Contains(t, attrTypes, "name")
}

func TestDsSignatureAttrTypes(t *testing.T) {
	attrTypes := dsSignatureAttrTypes()
	assert.Contains(t, attrTypes, "active")
	assert.Contains(t, attrTypes, "hash")
	assert.Contains(t, attrTypes, "name")
}
