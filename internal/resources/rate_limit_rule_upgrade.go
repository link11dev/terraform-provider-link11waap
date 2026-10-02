package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// rateLimitRuleSchemaVersion is the current schema version of the rate limit rule
// resource.
//
// Version 1 changed include/exclude from a set of objects (schema.SetNestedBlock)
// to a single object (schema.SingleNestedBlock) so that Terraform diffs the nested
// 'tags' list element by element. The configuration syntax is unchanged; only the
// stored state type differs, which is what rateLimitRuleStateUpgraderV0 migrates.
const rateLimitRuleSchemaVersion int64 = 1

// rateLimitRuleModelV0 is the resource model as it was written by schema version 0.
type rateLimitRuleModelV0 struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Global      types.Bool   `tfsdk:"global"`
	Active      types.Bool   `tfsdk:"active"`
	Timeframe   types.Int64  `tfsdk:"timeframe"`
	Threshold   types.Int64  `tfsdk:"threshold"`
	TTL         types.Int64  `tfsdk:"ttl"`
	Action      types.String `tfsdk:"action"`
	IsActionBan types.Bool   `tfsdk:"is_action_ban"`
	Tags        types.List   `tfsdk:"tags"`
	Key         types.List   `tfsdk:"key"`
	Pairwith    types.String `tfsdk:"pairwith"`
	Include     types.Set    `tfsdk:"include"`
	Exclude     types.Set    `tfsdk:"exclude"`
}

// rateLimitRuleSchemaV0 describes state written by schema version 0 of this resource.
//
// FROZEN: this must keep describing what older provider releases wrote, so it is
// never updated alongside the current schema. Only the attribute and block types
// matter here; validators, defaults and plan modifiers are irrelevant when
// decoding prior state and are therefore omitted.
func rateLimitRuleSchemaV0() *schema.Schema {
	tagFilterV0 := schema.SetNestedBlock{
		NestedObject: schema.NestedBlockObject{
			Attributes: map[string]schema.Attribute{
				"relation": schema.StringAttribute{Required: true},
				"tags": schema.ListAttribute{
					Required:    true,
					ElementType: types.StringType,
				},
			},
		},
	}

	return &schema.Schema{
		Version: 0,
		Attributes: map[string]schema.Attribute{
			"config_id":     schema.StringAttribute{Required: true},
			"id":            schema.StringAttribute{Computed: true},
			"name":          schema.StringAttribute{Required: true},
			"description":   schema.StringAttribute{Optional: true, Computed: true},
			"global":        schema.BoolAttribute{Required: true},
			"active":        schema.BoolAttribute{Required: true},
			"timeframe":     schema.Int64Attribute{Required: true},
			"threshold":     schema.Int64Attribute{Required: true},
			"ttl":           schema.Int64Attribute{Optional: true, Computed: true},
			"action":        schema.StringAttribute{Required: true},
			"is_action_ban": schema.BoolAttribute{Optional: true, Computed: true},
			"pairwith":      schema.StringAttribute{Optional: true, Computed: true},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"key": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attrs":   schema.StringAttribute{Optional: true},
						"args":    schema.StringAttribute{Optional: true},
						"plugins": schema.StringAttribute{Optional: true},
						"cookies": schema.StringAttribute{Optional: true},
						"headers": schema.StringAttribute{Optional: true},
					},
				},
			},
			"include": tagFilterV0,
			"exclude": tagFilterV0,
		},
	}
}

// UpgradeState migrates state written by schema version 0 to the current version.
func (r *RateLimitRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   rateLimitRuleSchemaV0(),
			StateUpgrader: rateLimitRuleStateUpgraderV0,
		},
	}
}

// rateLimitRuleStateUpgraderV0 rewrites the include/exclude sets as single objects
// and carries every other attribute over unchanged, so that upgrading the provider
// produces no plan diff.
func rateLimitRuleStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var old rateLimitRuleModelV0
	resp.Diagnostics.Append(req.State.Get(ctx, &old)...)
	if resp.Diagnostics.HasError() {
		return
	}

	include, diags := tagFilterSetToObject(ctx, old.Include, path.Root("include"))
	resp.Diagnostics.Append(diags...)
	exclude, diags := tagFilterSetToObject(ctx, old.Exclude, path.Root("exclude"))
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, RateLimitRuleResourceModel{
		ConfigID:    old.ConfigID,
		ID:          old.ID,
		Name:        old.Name,
		Description: old.Description,
		Global:      old.Global,
		Active:      old.Active,
		Timeframe:   old.Timeframe,
		Threshold:   old.Threshold,
		TTL:         old.TTL,
		Action:      old.Action,
		IsActionBan: old.IsActionBan,
		Tags:        old.Tags,
		Key:         old.Key,
		Pairwith:    old.Pairwith,
		Include:     include,
		Exclude:     exclude,
	})...)
}
