package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildConfig builds a tfsdk.Config from a resource schema and values.
func buildConfig(ctx context.Context, t *testing.T, r resource.Resource, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	tfType := sResp.Schema.Type().TerraformType(ctx)
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}

	fullValues := make(map[string]tftypes.Value)
	for attrName, attrType := range objType.AttributeTypes {
		fullValues[attrName] = tftypes.NewValue(attrType, nil)
	}
	for k, v := range values {
		fullValues[k] = v
	}

	raw := tftypes.NewValue(tfType, fullValues)

	return tfsdk.Config{
		Schema: sResp.Schema,
		Raw:    raw,
	}
}

// --- SecurityPolicy ValidateConfig Tests ---

func TestSecurityPolicyResource_ValidateConfig_NoSession(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		// session is empty (no blocks)
		"session": tftypes.NewValue(
			tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
			}}},
			[]tftypes.Value{},
		),
		"session_ids": tftypes.NewValue(
			tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
			}}},
			[]tftypes.Value{},
		),
		"map": tftypes.NewValue(
			tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String, "name": tftypes.String, "match": tftypes.String,
				"acl_profile": tftypes.String, "acl_profile_active": tftypes.Bool,
				"content_filter_profile": tftypes.String, "content_filter_profile_active": tftypes.Bool,
				"backend_service": tftypes.String, "description": tftypes.String,
				"rate_limit_rules": tftypes.List{ElementType: tftypes.String},
				"edge_functions":   tftypes.List{ElementType: tftypes.String},
			}}},
			[]tftypes.Value{},
		),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected error when no session blocks are specified")
}

func TestSecurityPolicyResource_ValidateConfig_ValidOneSession(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sessionBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"session": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{
				tftypes.NewValue(sessionBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"session_ids": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{},
		),
		"map": tftypes.NewValue(
			tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String, "name": tftypes.String, "match": tftypes.String,
				"acl_profile": tftypes.String, "acl_profile_active": tftypes.Bool,
				"content_filter_profile": tftypes.String, "content_filter_profile_active": tftypes.Bool,
				"backend_service": tftypes.String, "description": tftypes.String,
				"rate_limit_rules": tftypes.List{ElementType: tftypes.String},
				"edge_functions":   tftypes.List{ElementType: tftypes.String},
			}}},
			[]tftypes.Value{},
		),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "expected no errors for valid config with one session block: %v", resp.Diagnostics)
}

func TestSecurityPolicyResource_ValidateConfig_TwoFieldsInSession(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sessionBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"session": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{
				tftypes.NewValue(sessionBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, "user_id"), // two fields set
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"session_ids": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{},
		),
		"map": tftypes.NewValue(
			tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String, "name": tftypes.String, "match": tftypes.String,
				"acl_profile": tftypes.String, "acl_profile_active": tftypes.Bool,
				"content_filter_profile": tftypes.String, "content_filter_profile_active": tftypes.Bool,
				"backend_service": tftypes.String, "description": tftypes.String,
				"rate_limit_rules": tftypes.List{ElementType: tftypes.String},
				"edge_functions":   tftypes.List{ElementType: tftypes.String},
			}}},
			[]tftypes.Value{},
		),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected error when two fields are set in session block")
}

func TestSecurityPolicyResource_ValidateConfig_InvalidSessionIDs(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sessionBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"session": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{
				tftypes.NewValue(sessionBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"session_ids": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{
				tftypes.NewValue(sessionBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, "q"), // two fields set
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"map": tftypes.NewValue(
			tftypes.Set{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{
				"id": tftypes.String, "name": tftypes.String, "match": tftypes.String,
				"acl_profile": tftypes.String, "acl_profile_active": tftypes.Bool,
				"content_filter_profile": tftypes.String, "content_filter_profile_active": tftypes.Bool,
				"backend_service": tftypes.String, "description": tftypes.String,
				"rate_limit_rules": tftypes.List{ElementType: tftypes.String},
				"edge_functions":   tftypes.List{ElementType: tftypes.String},
			}}},
			[]tftypes.Value{},
		),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected error when session_ids has two fields set")
}

// --- SecurityPolicy ModifyPlan Tests ---

func TestSecurityPolicyResource_ModifyPlan_NullPlan(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	nullRaw := tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil)

	req := resource.ModifyPlanRequest{
		Plan:  tfsdk.Plan{Schema: sResp.Schema, Raw: nullRaw},
		State: tfsdk.State{Schema: sResp.Schema, Raw: nullRaw},
	}
	resp := &resource.ModifyPlanResponse{
		Plan: tfsdk.Plan{Schema: sResp.Schema, Raw: nullRaw},
	}

	r.ModifyPlan(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSecurityPolicyResource_ModifyPlan_NullState(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	nullRaw := tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil)
	plan := buildTerraformPlan(ctx, t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
	})

	req := resource.ModifyPlanRequest{
		Plan:  plan,
		State: tfsdk.State{Schema: sResp.Schema, Raw: nullRaw},
	}
	resp := &resource.ModifyPlanResponse{
		Plan: tfsdk.Plan{Schema: plan.Schema, Raw: plan.Raw.Copy()},
	}

	r.ModifyPlan(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestSecurityPolicyResource_ModifyPlan_BothExist(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sessionBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}

	mapBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"id": tftypes.String, "name": tftypes.String, "match": tftypes.String,
		"acl_profile": tftypes.String, "acl_profile_active": tftypes.Bool,
		"content_filter_profile": tftypes.String, "content_filter_profile_active": tftypes.Bool,
		"backend_service": tftypes.String, "description": tftypes.String,
		"rate_limit_rules": tftypes.List{ElementType: tftypes.String},
		"edge_functions":   tftypes.List{ElementType: tftypes.String},
	}}

	commonVals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "sp1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"tags":        tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
		"session": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{
				tftypes.NewValue(sessionBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"session_ids": tftypes.NewValue(
			tftypes.List{ElementType: sessionBlockType},
			[]tftypes.Value{},
		),
		"map": tftypes.NewValue(
			tftypes.Set{ElementType: mapBlockType},
			[]tftypes.Value{
				tftypes.NewValue(mapBlockType, map[string]tftypes.Value{
					"id":                            tftypes.NewValue(tftypes.String, "entry1"),
					"name":                          tftypes.NewValue(tftypes.String, "API"),
					"match":                         tftypes.NewValue(tftypes.String, "/api"),
					"acl_profile":                   tftypes.NewValue(tftypes.String, "acl1"),
					"acl_profile_active":            tftypes.NewValue(tftypes.Bool, true),
					"content_filter_profile":        tftypes.NewValue(tftypes.String, "cf1"),
					"content_filter_profile_active": tftypes.NewValue(tftypes.Bool, true),
					"backend_service":               tftypes.NewValue(tftypes.String, "be1"),
					"description":                   tftypes.NewValue(tftypes.String, ""),
					"rate_limit_rules":              tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
					"edge_functions":                tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
				}),
			},
		),
	}

	state := buildTerraformState(ctx, t, r, commonVals)
	plan := buildTerraformPlan(ctx, t, r, commonVals)

	req := resource.ModifyPlanRequest{
		Plan:  plan,
		State: state,
	}
	resp := &resource.ModifyPlanResponse{
		Plan: tfsdk.Plan{Schema: plan.Schema, Raw: plan.Raw.Copy()},
	}

	r.ModifyPlan(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "expected no errors: %v", resp.Diagnostics)
}

// --- RateLimitRule ValidateConfig Tests ---

func TestRateLimitRuleResource_ValidateConfig_NoKeys(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	keyBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}
	tagFilterBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"name":          tftypes.NewValue(tftypes.String, "test"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyBlockType},
			[]tftypes.Value{}, // empty - no keys
		),
		"include": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
		"exclude": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected error when no key blocks are specified")
}

func TestRateLimitRuleResource_ValidateConfig_ValidKey(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	keyBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}
	tagFilterBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"name":          tftypes.NewValue(tftypes.String, "test"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyBlockType},
			[]tftypes.Value{
				tftypes.NewValue(keyBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"include": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
		"exclude": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "expected no errors for valid config: %v", resp.Diagnostics)
}

func TestRateLimitRuleResource_ValidateConfig_TwoFieldsInKey(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	keyBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}
	tagFilterBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"name":          tftypes.NewValue(tftypes.String, "test"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyBlockType},
			[]tftypes.Value{
				tftypes.NewValue(keyBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, "ip"),
					"args":    tftypes.NewValue(tftypes.String, "user_id"), // two fields set
					"plugins": tftypes.NewValue(tftypes.String, nil),
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"include": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
		"exclude": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected error when two fields are set in key block")
}

func TestRateLimitRuleResource_ValidateConfig_InvalidPlugins(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	keyBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}
	tagFilterBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"name":          tftypes.NewValue(tftypes.String, "test"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyBlockType},
			[]tftypes.Value{
				tftypes.NewValue(keyBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, nil),
					"args":    tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, "invalid-no-jwt-prefix"), // should start with "jwt."
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"include": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
		"exclude": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected error for plugins value not starting with 'jwt.'")
}

func TestRateLimitRuleResource_ValidateConfig_ValidPlugins(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	keyBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String, "cookies": tftypes.String, "headers": tftypes.String,
	}}
	tagFilterBlockType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"name":          tftypes.NewValue(tftypes.String, "test"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyBlockType},
			[]tftypes.Value{
				tftypes.NewValue(keyBlockType, map[string]tftypes.Value{
					"attrs":   tftypes.NewValue(tftypes.String, nil),
					"args":    tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, "jwt.sub"), // valid
					"cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"include": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
		"exclude": tftypes.NewValue(tftypes.Set{ElementType: tagFilterBlockType}, []tftypes.Value{}),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "expected no errors for valid plugins with jwt prefix: %v", resp.Diagnostics)
}

// --- Publish Delete and getBuckets Tests ---

func TestPublishResource_Delete_Invoked(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerRegionsResource_Delete_Invoked(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := resource.DeleteRequest{}
	resp := &resource.DeleteResponse{}

	r.Delete(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
}

func TestPublishResource_GetBuckets_WithBuckets(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()
	var diags diag.Diagnostics

	bucketObjType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"url":  types.StringType,
		},
	}

	bucket1, _ := types.ObjectValue(
		map[string]attr.Type{
			"name": types.StringType,
			"url":  types.StringType,
		},
		map[string]attr.Value{
			"name": types.StringValue("bucket1"),
			"url":  types.StringValue("https://b1.example.com"),
		},
	)
	bucket2, _ := types.ObjectValue(
		map[string]attr.Type{
			"name": types.StringType,
			"url":  types.StringType,
		},
		map[string]attr.Value{
			"name": types.StringValue("bucket2"),
			"url":  types.StringValue("https://b2.example.com"),
		},
	)

	bucketList, _ := types.ListValue(bucketObjType, []attr.Value{bucket1, bucket2})

	model := &PublishResourceModel{
		Buckets: bucketList,
	}

	buckets := r.getBuckets(ctx, model, &diags)

	require.False(t, diags.HasError(), "unexpected errors: %v", diags)
	require.Len(t, buckets, 2)
	assert.Equal(t, "bucket1", buckets[0].Name)
	assert.Equal(t, "https://b1.example.com", buckets[0].URL)
	assert.Equal(t, "bucket2", buckets[1].Name)
	assert.Equal(t, "https://b2.example.com", buckets[1].URL)
}
