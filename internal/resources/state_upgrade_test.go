package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runStateUpgrade builds a prior state from priorSchema, fills in the supplied
// attribute values (everything else is null), and runs the resource's version 0
// state upgrader against it.
func runStateUpgrade(
	ctx context.Context,
	t *testing.T,
	r resource.ResourceWithUpgradeState,
	priorSchema *schema.Schema,
	values map[string]tftypes.Value,
) *resource.UpgradeStateResponse {
	t.Helper()

	priorType := priorSchema.Type().TerraformType(ctx)
	priorObjType, ok := priorType.(tftypes.Object)
	require.True(t, ok, "prior schema must produce an object type")

	attrs := make(map[string]tftypes.Value, len(priorObjType.AttributeTypes))
	for name, attrType := range priorObjType.AttributeTypes {
		attrs[name] = tftypes.NewValue(attrType, nil)
	}
	for name, value := range values {
		_, known := attrs[name]
		require.True(t, known, "attribute %q is not part of the prior schema", name)
		attrs[name] = value
	}

	sResp := schemaResp()
	r.Schema(ctx, schemaReq(), sResp)

	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	require.True(t, ok, "expected a state upgrader for schema version 0")
	require.NotNil(t, upgrader.PriorSchema, "the upgrader should declare its prior schema")

	req := resource.UpgradeStateRequest{
		State: &tfsdk.State{
			Schema: *upgrader.PriorSchema,
			Raw:    tftypes.NewValue(priorType, attrs),
		},
	}
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{
			Schema: sResp.Schema,
			Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
	upgrader.StateUpgrader(ctx, req, resp)
	return resp
}

// --- Dynamic rule ---

func TestDynamicRuleResource_UpgradeState_V0ToV1(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}
	objType := tagFilterObjType()

	resp := runStateUpgrade(ctx, t, r, dynamicRuleSchemaV0(), map[string]tftypes.Value{
		"config_id":            tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                   tftypes.NewValue(tftypes.String, "dr1"),
		"name":                 tftypes.NewValue(tftypes.String, "Burst Protection"),
		"description":          tftypes.NewValue(tftypes.String, "desc"),
		"threshold":            tftypes.NewValue(tftypes.Number, 100),
		"timeframe":            tftypes.NewValue(tftypes.Number, 60),
		"ttl":                  tftypes.NewValue(tftypes.Number, 3600),
		"active":               tftypes.NewValue(tftypes.Bool, true),
		"offload_ip_filtering": tftypes.NewValue(tftypes.Bool, false),
		"target":               tftypes.NewValue(tftypes.String, "ip"),
		"action":               tftypes.NewValue(tftypes.String, "action-monitor"),
		"tags": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "api"),
		}),
		"include": tagFilterSetValue(objType, tagFilterEntry("OR", "facebook")),
		"exclude": tagFilterSetValue(objType, tagFilterEntry("AND", "tor", "global-blacklist")),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diags: %v", resp.Diagnostics)

	var upgraded DynamicRuleResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())

	// Every attribute must survive the upgrade, otherwise practitioners see a
	// diff right after upgrading the provider.
	assert.Equal(t, "cfg1", upgraded.ConfigID.ValueString())
	assert.Equal(t, "dr1", upgraded.ID.ValueString())
	assert.Equal(t, "Burst Protection", upgraded.Name.ValueString())
	assert.Equal(t, "desc", upgraded.Description.ValueString())
	assert.Equal(t, int64(100), upgraded.Threshold.ValueInt64())
	assert.Equal(t, int64(60), upgraded.Timeframe.ValueInt64())
	assert.Equal(t, int64(3600), upgraded.TTL.ValueInt64())
	assert.True(t, upgraded.Active.ValueBool())
	assert.False(t, upgraded.OffloadIPFiltering.ValueBool())
	assert.Equal(t, "ip", upgraded.Target.ValueString())
	assert.Equal(t, "action-monitor", upgraded.Action.ValueString())

	var tags []string
	require.False(t, upgraded.Tags.ElementsAs(ctx, &tags, false).HasError())
	assert.Equal(t, []string{"api"}, tags)

	relation, filterTags := mustTagFilterAttrs(t, upgraded.Include)
	assert.Equal(t, "OR", relation)
	assert.Equal(t, []string{"facebook"}, filterTags)

	relation, filterTags = mustTagFilterAttrs(t, upgraded.Exclude)
	assert.Equal(t, "AND", relation)
	assert.Equal(t, []string{"tor", "global-blacklist"}, filterTags)
}

func TestDynamicRuleResource_UpgradeState_EmptySetBecomesNull(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}
	objType := tagFilterObjType()

	resp := runStateUpgrade(ctx, t, r, dynamicRuleSchemaV0(), map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, "dr1"),
		"include": tagFilterSetValue(objType),
		"exclude": tagFilterSetValue(objType),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diags: %v", resp.Diagnostics)

	var upgraded DynamicRuleResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())
	assert.True(t, upgraded.Include.IsNull(), "an empty set must upgrade to a null object")
	assert.True(t, upgraded.Exclude.IsNull(), "an empty set must upgrade to a null object")
}

func TestDynamicRuleResource_UpgradeState_MultipleBlocksWarn(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}
	objType := tagFilterObjType()

	resp := runStateUpgrade(ctx, t, r, dynamicRuleSchemaV0(), map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "dr1"),
		"include": tagFilterSetValue(objType,
			tagFilterEntry("OR", "facebook"),
			tagFilterEntry("AND", "tor"),
		),
		"exclude": tagFilterSetValue(objType, tagFilterEntry("OR")),
	})

	require.False(t, resp.Diagnostics.HasError(), "discarding extra blocks must not be fatal: %v", resp.Diagnostics)
	assert.Equal(t, 1, resp.Diagnostics.WarningsCount(), "expected one warning about the discarded block")

	var upgraded DynamicRuleResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())
	assert.False(t, upgraded.Include.IsNull(), "one of the blocks must be kept")
}

func TestDynamicRuleResource_UpgradeState_NullTagsNormalised(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}
	objType := tagFilterObjType()

	nullTags := map[string]tftypes.Value{
		"relation": tftypes.NewValue(tftypes.String, "OR"),
		"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	}

	resp := runStateUpgrade(ctx, t, r, dynamicRuleSchemaV0(), map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, "dr1"),
		"include": tagFilterSetValue(objType, nullTags),
		"exclude": tagFilterSetValue(objType, tagFilterEntry("OR")),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diags: %v", resp.Diagnostics)

	var upgraded DynamicRuleResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())

	relation, tags := mustTagFilterAttrs(t, upgraded.Include)
	assert.Equal(t, "OR", relation)
	assert.NotNil(t, tags)
	assert.Len(t, tags, 0, "a null tags list must be normalised to an empty list")
}

func TestDynamicRuleResource_UpgradeState_SchemaVersions(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}

	assert.Equal(t, int64(0), dynamicRuleSchemaV0().Version)
	upgraders := r.UpgradeState(ctx)
	assert.Len(t, upgraders, int(dynamicRuleSchemaVersion))
	for version := int64(0); version < dynamicRuleSchemaVersion; version++ {
		assert.Contains(t, upgraders, version, "missing upgrader for schema version %d", version)
	}
}

// --- Rate limit rule ---

func TestRateLimitRuleResource_UpgradeState_V0ToV1(t *testing.T) {
	ctx := context.Background()
	r := &RateLimitRuleResource{}
	objType := tagFilterObjType()

	keyObjType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String,
		"cookies": tftypes.String, "headers": tftypes.String,
	}}
	keyValue := tftypes.NewValue(tftypes.List{ElementType: keyObjType}, []tftypes.Value{
		tftypes.NewValue(keyObjType, map[string]tftypes.Value{
			"attrs":   tftypes.NewValue(tftypes.String, "session"),
			"args":    tftypes.NewValue(tftypes.String, nil),
			"plugins": tftypes.NewValue(tftypes.String, nil),
			"cookies": tftypes.NewValue(tftypes.String, nil),
			"headers": tftypes.NewValue(tftypes.String, nil),
		}),
	})

	resp := runStateUpgrade(ctx, t, r, rateLimitRuleSchemaV0(), map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"id":            tftypes.NewValue(tftypes.String, "rl1"),
		"name":          tftypes.NewValue(tftypes.String, "API Rate Limit"),
		"description":   tftypes.NewValue(tftypes.String, "desc"),
		"global":        tftypes.NewValue(tftypes.Bool, true),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"ttl":           tftypes.NewValue(tftypes.Number, 300),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, true),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
		"tags": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "api"),
		}),
		"key":     keyValue,
		"include": tagFilterSetValue(objType, tagFilterEntry("OR", "facebook")),
		"exclude": tagFilterSetValue(objType, tagFilterEntry("OR", "tor")),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diags: %v", resp.Diagnostics)

	var upgraded RateLimitRuleResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())

	assert.Equal(t, "cfg1", upgraded.ConfigID.ValueString())
	assert.Equal(t, "rl1", upgraded.ID.ValueString())
	assert.Equal(t, "API Rate Limit", upgraded.Name.ValueString())
	assert.Equal(t, "desc", upgraded.Description.ValueString())
	assert.True(t, upgraded.Global.ValueBool())
	assert.True(t, upgraded.Active.ValueBool())
	assert.Equal(t, int64(60), upgraded.Timeframe.ValueInt64())
	assert.Equal(t, int64(100), upgraded.Threshold.ValueInt64())
	assert.Equal(t, int64(300), upgraded.TTL.ValueInt64())
	assert.Equal(t, "action-monitor", upgraded.Action.ValueString())
	assert.True(t, upgraded.IsActionBan.ValueBool())
	assert.Equal(t, `{"self":"self"}`, upgraded.Pairwith.ValueString())

	var keys []RateLimitKeyModel
	require.False(t, upgraded.Key.ElementsAs(ctx, &keys, false).HasError())
	require.Len(t, keys, 1)
	assert.Equal(t, "session", keys[0].Attrs.ValueString())

	relation, filterTags := mustTagFilterAttrs(t, upgraded.Include)
	assert.Equal(t, "OR", relation)
	assert.Equal(t, []string{"facebook"}, filterTags)

	relation, filterTags = mustTagFilterAttrs(t, upgraded.Exclude)
	assert.Equal(t, "OR", relation)
	assert.Equal(t, []string{"tor"}, filterTags)
}

// A rate limit rule may legitimately omit both blocks; that must upgrade to a
// null object so the configuration and the state still agree.
func TestRateLimitRuleResource_UpgradeState_EmptySetBecomesNull(t *testing.T) {
	ctx := context.Background()
	r := &RateLimitRuleResource{}
	objType := tagFilterObjType()

	resp := runStateUpgrade(ctx, t, r, rateLimitRuleSchemaV0(), map[string]tftypes.Value{
		"id":      tftypes.NewValue(tftypes.String, "rl1"),
		"include": tagFilterSetValue(objType),
		"exclude": tagFilterSetValue(objType),
	})

	require.False(t, resp.Diagnostics.HasError(), "unexpected diags: %v", resp.Diagnostics)

	var upgraded RateLimitRuleResourceModel
	require.False(t, resp.State.Get(ctx, &upgraded).HasError())
	assert.True(t, upgraded.Include.IsNull())
	assert.True(t, upgraded.Exclude.IsNull())
}

func TestRateLimitRuleResource_UpgradeState_SchemaVersions(t *testing.T) {
	ctx := context.Background()
	r := &RateLimitRuleResource{}

	sResp := schemaResp()
	r.Schema(ctx, schemaReq(), sResp)
	assert.Equal(t, rateLimitRuleSchemaVersion, sResp.Schema.Version)
	assert.Equal(t, int64(0), rateLimitRuleSchemaV0().Version)

	for _, block := range []string{"include", "exclude"} {
		b, ok := sResp.Schema.Blocks[block]
		require.True(t, ok, "expected block %q", block)
		_, isSingle := b.(schema.SingleNestedBlock)
		assert.True(t, isSingle, "expected block %q to be a SingleNestedBlock, got %T", block, b)
	}
}

// tagFilterSetToObject is also reachable with a null or unknown set, which must
// not panic.
func TestTagFilterSetToObject_NullAndUnknown(t *testing.T) {
	ctx := context.Background()
	setType := types.ObjectType{AttrTypes: tagFilterAttrTypes}

	for name, set := range map[string]types.Set{
		"null":    types.SetNull(setType),
		"unknown": types.SetUnknown(setType),
	} {
		t.Run(name, func(t *testing.T) {
			obj, diags := tagFilterSetToObject(ctx, set, path.Root("include"))
			require.False(t, diags.HasError(), "unexpected diags: %v", diags)
			assert.True(t, obj.IsNull())
		})
	}
}
