package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserResource(t *testing.T) {
	r := NewUserResource()
	require.NotNil(t, r)
	_, ok := r.(*UserResource)
	assert.True(t, ok)
}

func TestUserResource_Metadata(t *testing.T) {
	r := &UserResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_user", resp.TypeName)
}

func TestUserResource_Schema(t *testing.T) {
	r := &UserResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"id", "acl", "contact_name", "email", "mobile", "org_id", "org_name",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestUserResource_Configure_NilProvider(t *testing.T) {
	r := &UserResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestUserResourceModel_Fields(t *testing.T) {
	model := UserResourceModel{
		ID:          types.StringValue("user-1"),
		ACL:         types.Int64Value(10),
		ContactName: types.StringValue("John Doe"),
		Email:       types.StringValue("john@example.com"),
		Mobile:      types.StringValue("+1234567890"),
		OrgID:       types.StringValue("org-1"),
		OrgName:     types.StringValue("Test Org"),
	}

	assert.Equal(t, "user-1", model.ID.ValueString())
	assert.Equal(t, int64(10), model.ACL.ValueInt64())
	assert.Equal(t, "John Doe", model.ContactName.ValueString())
	assert.Equal(t, "john@example.com", model.Email.ValueString())
	assert.Equal(t, "+1234567890", model.Mobile.ValueString())
	assert.Equal(t, "org-1", model.OrgID.ValueString())
	assert.Equal(t, "Test Org", model.OrgName.ValueString())
}
