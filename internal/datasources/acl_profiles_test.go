package datasources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestACLProfilesDataSource_Metadata(t *testing.T) {
	d := NewACLProfilesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_acl_profiles", resp.TypeName)
}

func TestACLProfilesDataSource_Schema(t *testing.T) {
	d := NewACLProfilesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "acl_profiles")
}

func TestACLProfilesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewACLProfilesDataSource())
}

func TestACLProfilesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewACLProfilesDataSource())
}

func TestACLProfilesDataSource_Read_Success(t *testing.T) {
	d := NewACLProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ACLProfile]{
			Total: 1,
			Items: []client.ACLProfile{
				{
					ID:          "acl1",
					Name:        "test-acl",
					Description: "Test ACL",
					Action:      "deny",
					Tags:        []string{"tag1", "tag2"},
					Allow:       []string{"allow1"},
					AllowBot:    []string{"bot1"},
					Deny:        []string{"deny1"},
					DenyBot:     []string{"dbot1"},
					ForceDeny:   []string{"fd1"},
					Passthrough: []string{"pt1"},
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

func TestACLProfilesDataSource_Read_EmptySlices(t *testing.T) {
	d := NewACLProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ACLProfile]{
			Total: 1,
			Items: []client.ACLProfile{
				{
					ID:     "acl1",
					Name:   "test-acl",
					Action: "deny",
					// nil slices - should produce null lists
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

func TestACLProfilesDataSource_Read_APIError(t *testing.T) {
	d := NewACLProfilesDataSource()
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

func TestACLProfilesDataSource_Read_MultipleProfiles(t *testing.T) {
	d := NewACLProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ACLProfile]{
			Total: 2,
			Items: []client.ACLProfile{
				{ID: "acl1", Name: "first", Action: "allow", Tags: []string{"t1"}},
				{ID: "acl2", Name: "second", Action: "deny", Allow: []string{"a1"}, Deny: []string{"d1"}},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestDsStringSliceToList_NilSlice(t *testing.T) {
	ctx := context.Background()
	resp := &datasource.ReadResponse{}
	result := dsStringSliceToList(ctx, nil, resp)
	assert.True(t, result.IsNull())
	assert.False(t, resp.Diagnostics.HasError())
}

func TestDsStringSliceToList_NonNilSlice(t *testing.T) {
	ctx := context.Background()
	resp := &datasource.ReadResponse{}
	result := dsStringSliceToList(ctx, []string{"a", "b"}, resp)
	assert.False(t, result.IsNull())
	assert.False(t, resp.Diagnostics.HasError())
}

func TestDsStringSliceToList_EmptySlice(t *testing.T) {
	ctx := context.Background()
	resp := &datasource.ReadResponse{}
	result := dsStringSliceToList(ctx, []string{}, resp)
	assert.False(t, result.IsNull())
	assert.False(t, resp.Diagnostics.HasError())
}
