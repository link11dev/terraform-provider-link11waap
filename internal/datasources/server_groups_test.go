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

func TestServerGroupsDataSource_Metadata(t *testing.T) {
	d := NewServerGroupsDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_server_groups", resp.TypeName)
}

func TestServerGroupsDataSource_Schema(t *testing.T) {
	d := NewServerGroupsDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "server_groups")
}

func TestServerGroupsDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewServerGroupsDataSource())
}

func TestServerGroupsDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewServerGroupsDataSource())
}

func TestServerGroupsDataSource_Read_Success(t *testing.T) {
	d := NewServerGroupsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ServerGroup]{
			Total: 1,
			Items: []client.ServerGroup{
				{
					ID:                     "sg1",
					Name:                   "site1",
					Description:            "Test server group",
					ServerNames:            []string{"example.com", "www.example.com"},
					SecurityPolicy:         "sp1",
					RoutingProfile:         "rp1",
					ProxyTemplate:          "pt1",
					ChallengeCookieDomain:  "example.com",
					SSLCertificate:         "cert1",
					ClientCertificate:      "ccert1",
					ClientCertificateMode:  "verify",
					MobileApplicationGroup: "mag1",
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

func TestServerGroupsDataSource_Read_MultipleGroups(t *testing.T) {
	d := NewServerGroupsDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ServerGroup]{
			Total: 2,
			Items: []client.ServerGroup{
				{ID: "sg1", Name: "site1", ServerNames: []string{"a.com"}, SecurityPolicy: "sp1", RoutingProfile: "rp1", ProxyTemplate: "pt1", ChallengeCookieDomain: "a.com"},
				{ID: "sg2", Name: "site2", ServerNames: []string{"b.com"}, SecurityPolicy: "sp2", RoutingProfile: "rp2", ProxyTemplate: "pt2", ChallengeCookieDomain: "b.com"},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestServerGroupsDataSource_Read_APIError(t *testing.T) {
	d := NewServerGroupsDataSource()
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
