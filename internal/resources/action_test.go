package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActionResource(t *testing.T) {
	r := NewActionResource()
	require.NotNil(t, r)
	_, ok := r.(*ActionResource)
	assert.True(t, ok)
}

func TestActionResource_Metadata(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_action", resp.TypeName)
}

func TestActionResource_Schema(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	s := resp.Schema
	assert.NotEmpty(t, s.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "type", "tags", "params",
	}
	for _, attr := range expectedAttrs {
		_, ok := s.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestActionResource_Schema_NameValidators(t *testing.T) {
	r := &ActionResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	name, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok, "name attribute should be StringAttribute")
	assert.True(t, name.Required)
	assert.NotEmpty(t, name.Validators, "name should reject empty strings via LengthAtLeast(1)")
}

func TestActionResource_Schema_TypeValidators(t *testing.T) {
	r := &ActionResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	typ, ok := resp.Schema.Attributes["type"].(schema.StringAttribute)
	require.True(t, ok, "type attribute should be StringAttribute")
	assert.True(t, typ.Required)
	assert.NotEmpty(t, typ.Validators, "type should have a OneOf enum validator")
}

func TestActionResource_Schema_ParamsNested(t *testing.T) {
	r := &ActionResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	params, ok := resp.Schema.Attributes["params"].(schema.SingleNestedAttribute)
	require.True(t, ok, "params attribute should be SingleNestedAttribute")
	assert.True(t, params.Optional)
	assert.True(t, params.Computed)
	for _, attr := range []string{"content", "status", "headers"} {
		_, ok := params.Attributes[attr]
		assert.True(t, ok, "expected params attribute %q", attr)
	}
}

func TestActionResource_Configure_NilProvider(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestActionResource_ImportState_Valid(t *testing.T) {
	r := &ActionResource{}
	resp := testImportState(t, r, "config123/a456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestActionResource_ImportState_Invalid(t *testing.T) {
	r := &ActionResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestActionResource_ImportState_TooManyParts(t *testing.T) {
	r := &ActionResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestActionParamsAttrTypes(t *testing.T) {
	m := actionParamsAttrTypes()
	for _, k := range []string{"content", "status", "headers"} {
		_, ok := m[k]
		assert.True(t, ok, "expected attr type %q", k)
	}
}

// --- buildAction unit tests ---

func TestActionResource_buildAction_NullTagsNullParams(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	plan := &ActionResourceModel{
		ID:          types.StringValue("id1"),
		ConfigID:    types.StringValue("cfg1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue("desc"),
		Type:        types.StringValue("block"),
		Tags:        types.ListNull(types.StringType),
		Params:      types.ObjectNull(actionParamsAttrTypes()),
	}

	var diags diag.Diagnostics
	a := r.buildAction(ctx, plan, &diags)

	require.False(t, diags.HasError())
	assert.Equal(t, "id1", a.ID)
	assert.Equal(t, "test", a.Name)
	assert.Equal(t, "desc", a.Description)
	assert.Equal(t, "block", a.Type)
	assert.Nil(t, a.Tags)
	assert.Nil(t, a.Params)
}

func TestActionResource_buildAction_WithTags(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	tags, d := types.ListValueFrom(ctx, types.StringType, []string{"env:prod", "v2"})
	require.False(t, d.HasError())

	plan := &ActionResourceModel{
		ID:     types.StringValue("id1"),
		Name:   types.StringValue("test"),
		Type:   types.StringValue("skip"),
		Tags:   tags,
		Params: types.ObjectNull(actionParamsAttrTypes()),
	}

	var diags diag.Diagnostics
	a := r.buildAction(ctx, plan, &diags)

	require.False(t, diags.HasError())
	assert.Equal(t, []string{"env:prod", "v2"}, a.Tags)
}

func TestActionResource_buildAction_WithFullParams(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	headersVal, d := types.MapValueFrom(ctx, types.StringType, map[string]string{"X-Custom": "header-value"})
	require.False(t, d.HasError())
	paramsObj, d2 := types.ObjectValue(actionParamsAttrTypes(), map[string]attr.Value{
		"content": types.StringValue("blocked"),
		"status":  types.Int64Value(403),
		"headers": headersVal,
	})
	require.False(t, d2.HasError())

	plan := &ActionResourceModel{
		ID:     types.StringValue("id1"),
		Name:   types.StringValue("test"),
		Type:   types.StringValue("block"),
		Tags:   types.ListNull(types.StringType),
		Params: paramsObj,
	}

	var diags diag.Diagnostics
	a := r.buildAction(ctx, plan, &diags)

	require.False(t, diags.HasError())
	require.NotNil(t, a.Params)
	assert.Equal(t, "blocked", a.Params.Content)
	require.NotNil(t, a.Params.Status)
	assert.Equal(t, 403, *a.Params.Status)
	assert.Equal(t, map[string]string{"X-Custom": "header-value"}, a.Params.Headers)
}

func TestActionResource_buildAction_ParamsNullStatusNullHeaders(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	paramsObj, d := types.ObjectValue(actionParamsAttrTypes(), map[string]attr.Value{
		"content": types.StringValue("denied"),
		"status":  types.Int64Null(),
		"headers": types.MapNull(types.StringType),
	})
	require.False(t, d.HasError())

	plan := &ActionResourceModel{
		ID:     types.StringValue("id1"),
		Name:   types.StringValue("test"),
		Type:   types.StringValue("block"),
		Tags:   types.ListNull(types.StringType),
		Params: paramsObj,
	}

	var diags diag.Diagnostics
	a := r.buildAction(ctx, plan, &diags)

	require.False(t, diags.HasError())
	require.NotNil(t, a.Params)
	assert.Equal(t, "denied", a.Params.Content)
	assert.Nil(t, a.Params.Status)
	assert.Nil(t, a.Params.Headers)
}

// --- flattenAction unit tests ---

func TestActionResource_flattenAction_NilParams(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	a := &client.Action{
		ID:          "id1",
		Name:        "test",
		Description: "desc",
		Type:        "monitor",
	}
	state := &ActionResourceModel{}
	var diags diag.Diagnostics

	r.flattenAction(ctx, a, state, &diags)

	require.False(t, diags.HasError())
	assert.Equal(t, "test", state.Name.ValueString())
	assert.Equal(t, "desc", state.Description.ValueString())
	assert.Equal(t, "monitor", state.Type.ValueString())
	assert.True(t, state.Tags.IsNull())
	assert.True(t, state.Params.IsNull())
}

func TestActionResource_flattenAction_EmptyTags(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	a := &client.Action{
		ID:   "id1",
		Name: "test",
		Type: "block",
		Tags: []string{},
	}
	state := &ActionResourceModel{}
	var diags diag.Diagnostics

	r.flattenAction(ctx, a, state, &diags)

	require.False(t, diags.HasError())
	assert.True(t, state.Tags.IsNull(), "empty tags slice should produce null list")
}

func TestActionResource_flattenAction_WithTags(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	a := &client.Action{
		ID:   "id1",
		Name: "test",
		Type: "block",
		Tags: []string{"tag1", "tag2"},
	}
	state := &ActionResourceModel{}
	var diags diag.Diagnostics

	r.flattenAction(ctx, a, state, &diags)

	require.False(t, diags.HasError())
	assert.False(t, state.Tags.IsNull())
	var tags []string
	diags.Append(state.Tags.ElementsAs(ctx, &tags, false)...)
	require.False(t, diags.HasError())
	assert.Equal(t, []string{"tag1", "tag2"}, tags)
}

func TestActionResource_flattenAction_WithFullParams(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	status := 403
	a := &client.Action{
		ID:   "id1",
		Name: "test",
		Type: "block",
		Params: &client.ActionParams{
			Content: "blocked",
			Status:  &status,
			Headers: map[string]string{"X-Custom": "value"},
		},
	}
	state := &ActionResourceModel{}
	var diags diag.Diagnostics

	r.flattenAction(ctx, a, state, &diags)

	require.False(t, diags.HasError())
	assert.False(t, state.Params.IsNull())

	var pm struct {
		Content types.String `tfsdk:"content"`
		Status  types.Int64  `tfsdk:"status"`
		Headers types.Map    `tfsdk:"headers"`
	}
	diags.Append(state.Params.As(ctx, &pm, basetypes.ObjectAsOptions{})...)
	require.False(t, diags.HasError())
	assert.Equal(t, "blocked", pm.Content.ValueString())
	assert.False(t, pm.Status.IsNull())
	assert.Equal(t, int64(403), pm.Status.ValueInt64())
	assert.False(t, pm.Headers.IsNull())
}

func TestActionResource_flattenAction_ParamsNilStatus(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	a := &client.Action{
		ID:   "id1",
		Name: "test",
		Type: "block",
		Params: &client.ActionParams{
			Content: "blocked",
			Status:  nil,
			Headers: map[string]string{"X-Custom": "value"},
		},
	}
	state := &ActionResourceModel{}
	var diags diag.Diagnostics

	r.flattenAction(ctx, a, state, &diags)

	require.False(t, diags.HasError())
	assert.False(t, state.Params.IsNull())
	attrs := state.Params.Attributes()
	statusVal, ok := attrs["status"]
	require.True(t, ok)
	assert.True(t, statusVal.IsNull(), "status should be null when pointer is nil")
}

func TestActionResource_flattenAction_ParamsNilHeaders(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	status := 200
	a := &client.Action{
		ID:   "id1",
		Name: "test",
		Type: "block",
		Params: &client.ActionParams{
			Content: "ok",
			Status:  &status,
			Headers: nil,
		},
	}
	state := &ActionResourceModel{}
	var diags diag.Diagnostics

	r.flattenAction(ctx, a, state, &diags)

	require.False(t, diags.HasError())
	assert.False(t, state.Params.IsNull())
	attrs := state.Params.Attributes()
	headersVal, ok := attrs["headers"]
	require.True(t, ok)
	assert.True(t, headersVal.IsNull(), "headers should be null when map is nil")
}
