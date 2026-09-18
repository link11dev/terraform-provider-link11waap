package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// dynamicRuleSchemaVersion is the current schema version of the dynamic rule
// resource.
//
// Version 1 changed include/exclude from a set of objects (schema.SetNestedBlock)
// to a single object (schema.SingleNestedBlock) so that Terraform diffs the nested
// 'tags' list element by element. The configuration syntax is unchanged; only the
// stored state type differs, which is what dynamicRuleStateUpgraderV0 migrates.
const dynamicRuleSchemaVersion int64 = 1

// dynamicRuleModelV0 is the resource model as it was written by schema version 0.
type dynamicRuleModelV0 struct {
	ConfigID           types.String `tfsdk:"config_id"`
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Threshold          types.Int64  `tfsdk:"threshold"`
	Timeframe          types.Int64  `tfsdk:"timeframe"`
	TTL                types.Int64  `tfsdk:"ttl"`
	Active             types.Bool   `tfsdk:"active"`
	OffloadIPFiltering types.Bool   `tfsdk:"offload_ip_filtering"`
	Target             types.String `tfsdk:"target"`
	Action             types.String `tfsdk:"action"`
	Tags               types.List   `tfsdk:"tags"`
	Include            types.Set    `tfsdk:"include"`
	Exclude            types.Set    `tfsdk:"exclude"`
}

// dynamicRuleSchemaV0 describes state written by schema version 0 of this resource.
//
// FROZEN: this must keep describing what older provider releases wrote, so it is
// never updated alongside the current schema. Only the attribute and block types
// matter here; validators, defaults and plan modifiers are irrelevant when
// decoding prior state and are therefore omitted.
func dynamicRuleSchemaV0() *schema.Schema {
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
			"config_id":            schema.StringAttribute{Required: true},
			"id":                   schema.StringAttribute{Computed: true},
			"name":                 schema.StringAttribute{Required: true},
			"description":          schema.StringAttribute{Optional: true, Computed: true},
			"threshold":            schema.Int64Attribute{Required: true},
			"timeframe":            schema.Int64Attribute{Required: true},
			"ttl":                  schema.Int64Attribute{Required: true},
			"active":               schema.BoolAttribute{Required: true},
			"offload_ip_filtering": schema.BoolAttribute{Required: true},
			"target":               schema.StringAttribute{Required: true},
			"action":               schema.StringAttribute{Required: true},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"include": tagFilterV0,
			"exclude": tagFilterV0,
		},
	}
}

// UpgradeState migrates state written by schema version 0 to the current version.
func (r *DynamicRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema:   dynamicRuleSchemaV0(),
			StateUpgrader: dynamicRuleStateUpgraderV0,
		},
	}
}

// dynamicRuleStateUpgraderV0 rewrites the include/exclude sets as single objects
// and carries every other attribute over unchanged, so that upgrading the provider
// produces no plan diff.
func dynamicRuleStateUpgraderV0(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
	var old dynamicRuleModelV0
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

	resp.Diagnostics.Append(resp.State.Set(ctx, DynamicRuleResourceModel{
		ConfigID:           old.ConfigID,
		ID:                 old.ID,
		Name:               old.Name,
		Description:        old.Description,
		Threshold:          old.Threshold,
		Timeframe:          old.Timeframe,
		TTL:                old.TTL,
		Active:             old.Active,
		OffloadIPFiltering: old.OffloadIPFiltering,
		Target:             old.Target,
		Action:             old.Action,
		Tags:               old.Tags,
		Include:            include,
		Exclude:            exclude,
	})...)
}
