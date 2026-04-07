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

func TestUsersDataSource_Metadata(t *testing.T) {
	d := NewUsersDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_users", resp.TypeName)
}

func TestUsersDataSource_Schema(t *testing.T) {
	d := NewUsersDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "users")
}

func TestUsersDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewUsersDataSource())
}

func TestUsersDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewUsersDataSource())
}

func TestUsersDataSource_Read_Success(t *testing.T) {
	d := NewUsersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]client.UserOrganization{
			{
				ID:   "org1",
				Name: "Org Inc",
				Users: []client.User{
					{
						ID:          "user1",
						ACL:         10,
						ContactName: "John Doe",
						Email:       "john@test.com",
						Mobile:      "+1234567890",
						OrgID:       "org1",
						OrgName:     "Org Inc",
					},
					{
						ID:          "user2",
						ACL:         5,
						ContactName: "Jane Doe",
						Email:       "jane@test.com",
						Mobile:      "+9876543210",
						OrgID:       "org1",
						OrgName:     "Org Inc",
					},
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestUsersDataSource_Read_MultipleOrgs(t *testing.T) {
	d := NewUsersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]client.UserOrganization{
			{
				ID:   "org1",
				Name: "Org One",
				Users: []client.User{
					{ID: "u1", ACL: 10, ContactName: "User 1", Email: "u1@test.com", OrgID: "org1", OrgName: "Org One"},
				},
			},
			{
				ID:   "org2",
				Name: "Org Two",
				Users: []client.User{
					{ID: "u2", ACL: 5, ContactName: "User 2", Email: "u2@test.com", OrgID: "org2", OrgName: "Org Two"},
					{ID: "u3", ACL: 1, ContactName: "User 3", Email: "u3@test.com", OrgID: "org2", OrgName: "Org Two"},
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestUsersDataSource_Read_EmptyOrgs(t *testing.T) {
	d := NewUsersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode([]client.UserOrganization{})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestUsersDataSource_Read_APIError(t *testing.T) {
	d := NewUsersDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{})
	assert.True(t, resp.Diagnostics.HasError())
}
