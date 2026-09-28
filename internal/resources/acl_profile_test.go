package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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

	result := stringSliceToList(ctx, []string{"a", "b", "c"}, types.ListNull(types.StringType), resp)

	assert.False(t, result.IsNull())
	assert.False(t, result.IsUnknown())

	var elements []string
	result.ElementsAs(ctx, &elements, false)
	assert.Equal(t, []string{"a", "b", "c"}, elements)
}

func TestStringSliceToList_Empty(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, []string{}, types.ListNull(types.StringType), resp)

	assert.True(t, result.IsNull())
}

func TestStringSliceToList_Nil(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, nil, types.ListNull(types.StringType), resp)

	assert.True(t, result.IsNull())
}

func TestStringSliceToList_SingleElement(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, []string{"only"}, types.ListNull(types.StringType), resp)

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
		Action:      types.StringValue("703c7a701c2e"),
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
	assert.Equal(t, "703c7a701c2e", model.Action.ValueString())
	assert.True(t, model.Tags.IsNull())
}

// --- empty list refresh behaviour ---

// TestStringSliceToList_EmptyMirrorsPriorEmpty covers the perpetual-diff case:
// a configuration that set `allow_bot = []` stores an empty list, and a refresh
// must not rewrite it to null, or every subsequent plan re-adds the empty list.
func TestStringSliceToList_EmptyMirrorsPriorEmpty(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	prior := types.ListValueMust(types.StringType, []attr.Value{})

	result := stringSliceToList(ctx, nil, prior, resp)

	assert.False(t, result.IsNull(), "an explicitly empty list must survive refresh")
	assert.Empty(t, result.Elements())
}

// TestStringSliceToList_EmptyKeepsPriorNull is the counterpart: an attribute that
// was never set stays null rather than gaining an empty list.
func TestStringSliceToList_EmptyKeepsPriorNull(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	result := stringSliceToList(ctx, nil, types.ListNull(types.StringType), resp)

	assert.True(t, result.IsNull())
}

// TestStringSliceToList_EmptyAfterPopulatedIsDrift makes sure real drift is still
// reported: a list that had elements and now comes back empty becomes null rather
// than silently keeping the old value.
func TestStringSliceToList_EmptyAfterPopulatedIsDrift(t *testing.T) {
	ctx := context.Background()
	resp := &readResp{}

	prior := types.ListValueMust(types.StringType, []attr.Value{types.StringValue("tag-a")})

	result := stringSliceToList(ctx, nil, prior, resp)

	assert.True(t, result.IsNull(), "removal of all elements upstream must surface as drift")
}
