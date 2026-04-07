package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewACLProfileResource(t *testing.T) {
	r := NewACLProfileResource()
	require.NotNil(t, r)
	_, ok := r.(*ACLProfileResource)
	assert.True(t, ok)
}

func TestACLProfileResource_Metadata(t *testing.T) {
	r := &ACLProfileResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_acl_profile", resp.TypeName)
}

func TestACLProfileResource_Schema(t *testing.T) {
	r := &ACLProfileResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "tags",
		"action", "allow", "allow_bot", "deny", "deny_bot",
		"force_deny", "passthrough",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestACLProfileResource_Configure_NilProvider(t *testing.T) {
	r := &ACLProfileResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestACLProfileResource_ImportState_Valid(t *testing.T) {
	r := &ACLProfileResource{}
	resp := testImportState(t, r, "config123/acl456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestACLProfileResource_ImportState_Invalid(t *testing.T) {
	r := &ACLProfileResource{}
	resp := testImportState(t, r, "invalidformat")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestACLProfileResource_ImportState_TooManyParts(t *testing.T) {
	r := &ACLProfileResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestStringSliceToList_NonEmpty(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, []string{"a", "b", "c"}, resp)

	assert.False(t, result.IsNull())
	assert.False(t, result.IsUnknown())

	var elements []string
	result.ElementsAs(ctx, &elements, false)
	assert.Equal(t, []string{"a", "b", "c"}, elements)
}

func TestStringSliceToList_Empty(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, []string{}, resp)

	assert.True(t, result.IsNull())
}

func TestStringSliceToList_Nil(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, nil, resp)

	assert.True(t, result.IsNull())
}

func TestStringSliceToList_SingleElement(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, []string{"only"}, resp)

	assert.False(t, result.IsNull())

	var elements []string
	result.ElementsAs(ctx, &elements, false)
	assert.Equal(t, []string{"only"}, elements)
}

// ACLProfileResourceModel field tests
func TestACLProfileResourceModel_FieldTypes(t *testing.T) {
	model := ACLProfileResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("id1"),
		Name:        types.StringValue("test-acl"),
		Description: types.StringValue("desc"),
		Action:      types.StringValue("action-acl-block"),
		Tags:        types.ListNull(types.StringType),
		Allow:       types.ListNull(types.StringType),
		AllowBot:    types.ListNull(types.StringType),
		Deny:        types.ListNull(types.StringType),
		DenyBot:     types.ListNull(types.StringType),
		ForceDeny:   types.ListNull(types.StringType),
		Passthrough: types.ListNull(types.StringType),
	}

	assert.Equal(t, "cfg1", model.ConfigID.ValueString())
	assert.Equal(t, "id1", model.ID.ValueString())
	assert.Equal(t, "test-acl", model.Name.ValueString())
	assert.Equal(t, "desc", model.Description.ValueString())
	assert.Equal(t, "action-acl-block", model.Action.ValueString())
	assert.True(t, model.Tags.IsNull())
}
