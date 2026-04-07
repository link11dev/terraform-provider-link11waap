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

func TestSecurityPoliciesDataSource_Metadata(t *testing.T) {
	d := NewSecurityPoliciesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_security_policies", resp.TypeName)
}

func TestSecurityPoliciesDataSource_Schema(t *testing.T) {
	d := NewSecurityPoliciesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "security_policies")
}

func TestSecurityPoliciesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewSecurityPoliciesDataSource())
}

func TestSecurityPoliciesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewSecurityPoliciesDataSource())
}

func TestSecurityPoliciesDataSource_Read_Success(t *testing.T) {
	d := NewSecurityPoliciesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.SecurityPolicy]{
			Total: 1,
			Items: []client.SecurityPolicy{
				{
					ID:          "sp1",
					Name:        "test-policy",
					Description: "Test security policy",
					Tags:        []string{"tag1", "tag2"},
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

func TestSecurityPoliciesDataSource_Read_NilTags(t *testing.T) {
	d := NewSecurityPoliciesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.SecurityPolicy]{
			Total: 1,
			Items: []client.SecurityPolicy{
				{ID: "sp1", Name: "test-policy", Tags: nil},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestSecurityPoliciesDataSource_Read_APIError(t *testing.T) {
	d := NewSecurityPoliciesDataSource()
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

func TestSecurityPoliciesDataSource_Read_MultiplePolicies(t *testing.T) {
	d := NewSecurityPoliciesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.SecurityPolicy]{
			Total: 2,
			Items: []client.SecurityPolicy{
				{ID: "sp1", Name: "first", Tags: []string{"t1"}},
				{ID: "sp2", Name: "second", Tags: nil},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}
