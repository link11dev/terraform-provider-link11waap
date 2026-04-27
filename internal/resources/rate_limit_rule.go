package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                   = &RateLimitRuleResource{}
	_ resource.ResourceWithImportState    = &RateLimitRuleResource{}
	_ resource.ResourceWithValidateConfig = &RateLimitRuleResource{}
)

// RateLimitRuleResource implements the resource for managing a rate limit rule
type RateLimitRuleResource struct {
	client *client.Client
}

// RateLimitRuleResourceModel describes the resource model for a rate limit rule
type RateLimitRuleResourceModel struct {
	ConfigID    types.String        `tfsdk:"config_id"`
	ID          types.String        `tfsdk:"id"`
	Name        types.String        `tfsdk:"name"`
	Description types.String        `tfsdk:"description"`
	Global      types.Bool          `tfsdk:"global"`
	Active      types.Bool          `tfsdk:"active"`
	Timeframe   types.Int64         `tfsdk:"timeframe"`
	Threshold   types.Int64         `tfsdk:"threshold"`
	TTL         types.Int64         `tfsdk:"ttl"`
	Action      types.String        `tfsdk:"action"`
	IsActionBan types.Bool          `tfsdk:"is_action_ban"`
	Tags        types.List          `tfsdk:"tags"`
	Key         []RateLimitKeyModel `tfsdk:"key"`
	Pairwith    types.String        `tfsdk:"pairwith"`
	Include     types.Set           `tfsdk:"include"`
	Exclude     types.Set           `tfsdk:"exclude"`
	// LastActivated types.Int64         `tfsdk:"last_activated"`
}

// RateLimitKeyModel describes the data model for a rate limit key entry
type RateLimitKeyModel struct {
	Attrs   types.String `tfsdk:"attrs"`
	Args    types.String `tfsdk:"args"`
	Plugins types.String `tfsdk:"plugins"`
	Cookies types.String `tfsdk:"cookies"`
	Headers types.String `tfsdk:"headers"`
}

// RateLimitTagFilterModel describes the data model for a rate limit tag filter (include/exclude)
type RateLimitTagFilterModel struct {
	Relation types.String `tfsdk:"relation"`
	Tags     types.List   `tfsdk:"tags"`
}

// tagFilterAttrTypes defines the attribute types for a tag filter object
var tagFilterAttrTypes = map[string]attr.Type{
	"relation": types.StringType,
	"tags":     types.ListType{ElemType: types.StringType},
}

// NewRateLimitRuleResource returns a new instance of the rate limit rule resource
func NewRateLimitRuleResource() resource.Resource {
	return &RateLimitRuleResource{}
}

// Metadata returns the resource type name
func (r *RateLimitRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rate_limit_rule"
}

// Schema defines the schema for the rate limit rule resource
func (r *RateLimitRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Rate Limit Rule in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the rate limit rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the rate limit rule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the rate limit rule.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"global": schema.BoolAttribute{
				Description: "Whether this is a global rate limit rule.",
				Required:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the rate limit rule is active.",
				Required:    true,
			},
			"timeframe": schema.Int64Attribute{
				Description: "Time window in seconds for counting requests.",
				Required:    true,
			},
			"threshold": schema.Int64Attribute{
				Description: "Maximum number of requests allowed within the timeframe.",
				Required:    true,
			},
			"ttl": schema.Int64Attribute{
				Description: "Time-to-live in seconds for the rate limit ban.",
				Optional:    true,
				Computed:    true,
			},
			"action": schema.StringAttribute{
				Description: "Action to take when the rate limit is exceeded.",
				Required:    true,
			},
			"is_action_ban": schema.BoolAttribute{
				Description: "Whether the action is a ban action.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"tags": schema.ListAttribute{
				Description: "List of tags associated with the rate limit rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"pairwith": schema.StringAttribute{
				Description: "Pair-with configuration as a JSON string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(`{"self":"self"}`),
			},
			// "last_activated": schema.Int64Attribute{
			// 	Description: "Unix timestamp of last activation.",
			// 	Computed:    true,
			// },
		},
		Blocks: map[string]schema.Block{
			"key": schema.ListNestedBlock{
				Description: "Rate limit key configuration. At least one block is required. Exactly one of attrs, args, plugins, cookies, or headers must be set per block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attrs": schema.StringAttribute{
							Description: "Rate limit by request attribute (e.g., 'session', 'ip', 'authority').",
							Optional:    true,
						},
						"args": schema.StringAttribute{
							Description: "Rate limit by query argument name.",
							Optional:    true,
						},
						"plugins": schema.StringAttribute{
							Description: "Rate limit by plugin data. Value must start with 'jwt.'.",
							Optional:    true,
						},
						"cookies": schema.StringAttribute{
							Description: "Rate limit by cookie name.",
							Optional:    true,
						},
						"headers": schema.StringAttribute{
							Description: "Rate limit by header name.",
							Optional:    true,
						},
					},
				},
			},
			"include": schema.SetNestedBlock{
				Description: "Include filter: requests matching these tags are counted.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"relation": schema.StringAttribute{
							Description: "Relation between tags. Valid values: OR, AND.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("OR", "AND"),
							},
						},
						"tags": schema.ListAttribute{
							Description: "List of tag identifiers.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"exclude": schema.SetNestedBlock{
				Description: "Exclude filter: requests matching these tags are excluded from counting.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"relation": schema.StringAttribute{
							Description: "Relation between tags. Valid values: OR, AND.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("OR", "AND"),
							},
						},
						"tags": schema.ListAttribute{
							Description: "List of tag identifiers.",
							Required:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure sets the client for the resource
func (r *RateLimitRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new rate limit rule using the API client and sets the resource state
func (r *RateLimitRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RateLimitRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	rule, diags := buildRateLimitRuleAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateRateLimitRule(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Rate Limit Rule",
			"Could not create rate limit rule: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read retrieves the rate limit rule from the API and updates the resource state. If the rule is not found, it removes the resource from state.
func (r *RateLimitRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RateLimitRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetRateLimitRule(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Rate Limit Rule",
			"Could not read rate limit rule: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(rule.Name)
	state.Description = types.StringValue(rule.Description)
	state.Global = types.BoolValue(rule.Global)
	state.Active = types.BoolValue(rule.Active)
	state.Timeframe = types.Int64Value(int64(rule.Timeframe))
	state.Threshold = types.Int64Value(int64(rule.Threshold))
	state.TTL = types.Int64Value(int64(rule.TTL))
	state.Action = types.StringValue(rule.Action)
	state.IsActionBan = types.BoolValue(rule.IsActionBan)
	// state.LastActivated = types.Int64Value(int64(rule.LastActivated))

	// Tags
	if len(rule.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, rule.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	// Key ([]interface{} -> []RateLimitKeyModel)
	state.Key = parseRateLimitKeys(rule.Key)

	// Pairwith (interface{} -> JSON string)
	if rule.Pairwith != nil {
		pairwithBytes, err := json.Marshal(rule.Pairwith)
		if err != nil {
			resp.Diagnostics.AddError("Error Marshaling Pairwith", err.Error())
			return
		}
		state.Pairwith = types.StringValue(string(pairwithBytes))
	} else {
		state.Pairwith = types.StringNull()
	}

	// Include
	includeSet, diags := tagFilterToSet(ctx, rule.Include)
	resp.Diagnostics.Append(diags...)
	state.Include = includeSet

	// Exclude
	excludeSet, diags := tagFilterToSet(ctx, rule.Exclude)
	resp.Diagnostics.Append(diags...)
	state.Exclude = excludeSet

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the rate limit rule using the API client and sets the resource state
func (r *RateLimitRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RateLimitRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, diags := buildRateLimitRuleAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateRateLimitRule(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Rate Limit Rule",
			"Could not update rate limit rule: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the rate limit rule using the API client
func (r *RateLimitRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RateLimitRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRateLimitRule(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Rate Limit Rule",
			"Could not delete rate limit rule: "+err.Error(),
		)
		return
	}
}

// ImportState imports the rate limit rule using the format "config_id/rate_limit_rule_id"
func (r *RateLimitRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/rate_limit_rule_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// ValidateConfig validates that at least one key block is specified and each block has exactly one field set.
func (r *RateLimitRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config RateLimitRuleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if len(config.Key) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Missing Required key Blocks",
			"At least one 'key' block must be specified.",
		)
		return
	}
	for i, k := range config.Key {
		setCount := 0
		if !k.Attrs.IsNull() && !k.Attrs.IsUnknown() {
			setCount++
		}
		if !k.Args.IsNull() && !k.Args.IsUnknown() {
			setCount++
		}
		if !k.Plugins.IsNull() && !k.Plugins.IsUnknown() {
			setCount++
			if !strings.HasPrefix(k.Plugins.ValueString(), "jwt.") {
				resp.Diagnostics.AddAttributeError(
					path.Root("key").AtListIndex(i).AtName("plugins"),
					"Invalid plugins value",
					"The 'plugins' value must start with 'jwt.'.",
				)
			}
		}
		if !k.Cookies.IsNull() && !k.Cookies.IsUnknown() {
			setCount++
		}
		if !k.Headers.IsNull() && !k.Headers.IsUnknown() {
			setCount++
		}
		if setCount != 1 {
			resp.Diagnostics.AddAttributeError(
				path.Root("key").AtListIndex(i),
				"Invalid key block",
				"Exactly one of attrs, args, plugins, cookies, or headers must be set in each 'key' block.",
			)
		}
	}
}

func buildRateLimitRuleAPIModel(ctx context.Context, plan *RateLimitRuleResourceModel) (*client.RateLimitRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := &client.RateLimitRule{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Global:      plan.Global.ValueBool(),
		Active:      plan.Active.ValueBool(),
		Timeframe:   int(plan.Timeframe.ValueInt64()),
		Threshold:   int(plan.Threshold.ValueInt64()),
		TTL:         int(plan.TTL.ValueInt64()),
		Action:      plan.Action.ValueString(),
		IsActionBan: plan.IsActionBan.ValueBool(),
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &rule.Tags, false)...)
	}

	// Key ([]RateLimitKeyModel -> []map[string]string)
	if len(plan.Key) > 0 {
		rule.Key = buildRateLimitKeys(plan.Key)
	}

	// Pairwith (JSON string -> interface{})
	if !plan.Pairwith.IsNull() && !plan.Pairwith.IsUnknown() {
		var pairwith interface{}
		if err := json.Unmarshal([]byte(plan.Pairwith.ValueString()), &pairwith); err == nil {
			rule.Pairwith = pairwith
		}
	} else {
		rule.Pairwith = map[string]string{"self": "self"}
	}

	// Include
	includeFilter, d := extractTagFilter(ctx, plan.Include)
	diags.Append(d...)
	rule.Include = includeFilter

	// Exclude
	excludeFilter, d := extractTagFilter(ctx, plan.Exclude)
	diags.Append(d...)
	rule.Exclude = excludeFilter

	return rule, diags
}

// parseRateLimitKeys converts the API interface{} key to []RateLimitKeyModel
func parseRateLimitKeys(raw interface{}) []RateLimitKeyModel {
	if raw == nil {
		return []RateLimitKeyModel{}
	}
	rawKeys, ok := raw.([]interface{})
	if !ok {
		return []RateLimitKeyModel{}
	}
	keys := make([]RateLimitKeyModel, 0, len(rawKeys))
	for _, rk := range rawKeys {
		m, ok := rk.(map[string]interface{})
		if !ok {
			continue
		}
		km := RateLimitKeyModel{
			Attrs:   types.StringNull(),
			Args:    types.StringNull(),
			Plugins: types.StringNull(),
			Cookies: types.StringNull(),
			Headers: types.StringNull(),
		}
		if v, ok := m["attrs"].(string); ok {
			km.Attrs = types.StringValue(v)
		} else if v, ok := m["args"].(string); ok {
			km.Args = types.StringValue(v)
		} else if v, ok := m["plugins"].(string); ok {
			km.Plugins = types.StringValue(v)
		} else if v, ok := m["cookies"].(string); ok {
			km.Cookies = types.StringValue(v)
		} else if v, ok := m["headers"].(string); ok {
			km.Headers = types.StringValue(v)
		}
		keys = append(keys, km)
	}
	return keys
}

// buildRateLimitKeys converts []RateLimitKeyModel to []map[string]string for the API
func buildRateLimitKeys(keys []RateLimitKeyModel) []map[string]string {
	result := make([]map[string]string, 0, len(keys))
	for _, k := range keys {
		if !k.Attrs.IsNull() && !k.Attrs.IsUnknown() {
			result = append(result, map[string]string{"attrs": k.Attrs.ValueString()})
		} else if !k.Args.IsNull() && !k.Args.IsUnknown() {
			result = append(result, map[string]string{"args": k.Args.ValueString()})
		} else if !k.Plugins.IsNull() && !k.Plugins.IsUnknown() {
			result = append(result, map[string]string{"plugins": k.Plugins.ValueString()})
		} else if !k.Cookies.IsNull() && !k.Cookies.IsUnknown() {
			result = append(result, map[string]string{"cookies": k.Cookies.ValueString()})
		} else if !k.Headers.IsNull() && !k.Headers.IsUnknown() {
			result = append(result, map[string]string{"headers": k.Headers.ValueString()})
		}
	}
	return result
}

// extractTagFilter converts a Terraform set to an API RateLimitTagFilter.
func extractTagFilter(ctx context.Context, set types.Set) (client.RateLimitTagFilter, diag.Diagnostics) {
	var models []RateLimitTagFilterModel
	diags := set.ElementsAs(ctx, &models, false)
	if diags.HasError() {
		return client.RateLimitTagFilter{}, diags
	}
	if len(models) == 0 {
		return client.RateLimitTagFilter{Relation: "OR", Tags: []string{}}, diags
	}
	model := models[0]
	var tags []string
	diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	return client.RateLimitTagFilter{
		Relation: model.Relation.ValueString(),
		Tags:     tags,
	}, diags
}

// tagFilterToSet converts an API RateLimitTagFilter to a Terraform set.
func tagFilterToSet(ctx context.Context, filter client.RateLimitTagFilter) (types.Set, diag.Diagnostics) {
	tags := filter.Tags
	if tags == nil {
		tags = []string{}
	}
	tagsList, diags := types.ListValueFrom(ctx, types.StringType, tags)
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: tagFilterAttrTypes}), diags
	}
	obj, d := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
		"relation": types.StringValue(filter.Relation),
		"tags":     tagsList,
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.SetNull(types.ObjectType{AttrTypes: tagFilterAttrTypes}), diags
	}
	set, d := types.SetValue(types.ObjectType{AttrTypes: tagFilterAttrTypes}, []attr.Value{obj})
	diags.Append(d...)
	return set, diags
}
