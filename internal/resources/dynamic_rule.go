package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                   = &DynamicRuleResource{}
	_ resource.ResourceWithImportState    = &DynamicRuleResource{}
	_ resource.ResourceWithValidateConfig = &DynamicRuleResource{}
	_ resource.ResourceWithUpgradeState   = &DynamicRuleResource{}
)

// DynamicRuleResource implements the resource for managing a dynamic rule
type DynamicRuleResource struct {
	client *client.Client
}

// DynamicRuleResourceModel describes the resource model for a dynamic rule
type DynamicRuleResourceModel struct {
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
	Include            types.Object `tfsdk:"include"`
	Exclude            types.Object `tfsdk:"exclude"`
}

// NewDynamicRuleResource returns a new instance of the dynamic rule resource
func NewDynamicRuleResource() resource.Resource {
	return &DynamicRuleResource{}
}

// Metadata returns the resource type name
func (r *DynamicRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_rule"
}

// Schema defines the schema for the dynamic rule resource
func (r *DynamicRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Version:     dynamicRuleSchemaVersion,
		Description: "Manages a Dynamic Rule in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the dynamic rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the dynamic rule.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the dynamic rule.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"threshold": schema.Int64Attribute{
				Description: "Maximum number of matching requests allowed within the timeframe.",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"timeframe": schema.Int64Attribute{
				Description: "Time window in seconds for counting requests.",
				Required:    true,
				Validators:  []validator.Int64{int64validator.AtLeast(1)},
			},
			"ttl": schema.Int64Attribute{
				Description: "Time-to-live in seconds for the dynamic rule ban. Must be a positive multiple of 3600 (full hours).",
				Required:    true,
				Validators: []validator.Int64{
					FullHours(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the dynamic rule is active.",
				Required:    true,
			},
			"offload_ip_filtering": schema.BoolAttribute{
				Description: "Whether IP filtering for this rule is offloaded to the edge.",
				Required:    true,
			},
			"target": schema.StringAttribute{
				Description: "The request attribute the rule counts on. One of 'ip', 'asn', 'country', 'organization', " +
					"or one of 'arguments', 'headers', 'cookies' followed by an underscore and a name " +
					"(e.g. 'headers_x-test-header', 'cookies_some_cookie').",
				Required: true,
				Validators: []validator.String{
					DynamicRuleTarget(),
				},
			},
			"action": schema.StringAttribute{
				Description: "Action to take when the threshold is exceeded.",
				Required:    true,
			},
			"tags": schema.ListAttribute{
				Description: "List of tags associated with the dynamic rule.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"include": tagFilterBlockSchema(
				"Include filter: requests matching these tags are counted. Exactly one block is required.",
			),
			"exclude": tagFilterBlockSchema(
				"Exclude filter: requests matching these tags are excluded from counting. Exactly one block is required.",
			),
		},
	}
}

// ValidateConfig validates that exactly one 'include' block and exactly one 'exclude' block are specified.
func (r *DynamicRuleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config DynamicRuleResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	validateTagFilterBlock(ctx, config.Include, path.Root("include"), true, &resp.Diagnostics)
	validateTagFilterBlock(ctx, config.Exclude, path.Root("exclude"), true, &resp.Diagnostics)
}

// Configure sets the client for the resource
func (r *DynamicRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new dynamic rule using the API client and sets the resource state
func (r *DynamicRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DynamicRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	rule, diags := buildDynamicRuleAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateDynamicRule(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dynamic Rule",
			"Could not create dynamic rule: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read retrieves the dynamic rule from the API and updates the resource state. If the rule is not found, it removes the resource from state.
func (r *DynamicRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DynamicRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetDynamicRule(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Dynamic Rule",
			"Could not read dynamic rule: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(rule.Name)
	state.Description = types.StringValue(rule.Description)
	state.Threshold = types.Int64Value(int64(rule.Threshold))
	state.Timeframe = types.Int64Value(int64(rule.Timeframe))
	state.TTL = types.Int64Value(int64(rule.TTL))
	state.Active = types.BoolValue(rule.Active)
	state.OffloadIPFiltering = types.BoolValue(rule.OffloadIPFiltering)
	state.Target = types.StringValue(rule.Target)
	state.Action = types.StringValue(rule.Action)

	// Tags
	if len(rule.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, rule.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	// Include
	includeObj, diags := tagFilterToObject(ctx, rule.Include)
	resp.Diagnostics.Append(diags...)
	state.Include = includeObj

	// Exclude
	excludeObj, diags := tagFilterToObject(ctx, rule.Exclude)
	resp.Diagnostics.Append(diags...)
	state.Exclude = excludeObj

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the dynamic rule using the API client and sets the resource state
func (r *DynamicRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DynamicRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, diags := buildDynamicRuleAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateDynamicRule(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dynamic Rule",
			"Could not update dynamic rule: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the dynamic rule using the API client
func (r *DynamicRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DynamicRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDynamicRule(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dynamic Rule",
			"Could not delete dynamic rule: "+err.Error(),
		)
		return
	}
}

// ImportState imports the dynamic rule using the format "config_id/dynamic_rule_id"
func (r *DynamicRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/dynamic_rule_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func buildDynamicRuleAPIModel(ctx context.Context, plan *DynamicRuleResourceModel) (*client.DynamicRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	rule := &client.DynamicRule{
		ID:                 plan.ID.ValueString(),
		Name:               plan.Name.ValueString(),
		Description:        plan.Description.ValueString(),
		Threshold:          int(plan.Threshold.ValueInt64()),
		Timeframe:          int(plan.Timeframe.ValueInt64()),
		TTL:                int(plan.TTL.ValueInt64()),
		Active:             plan.Active.ValueBool(),
		OffloadIPFiltering: plan.OffloadIPFiltering.ValueBool(),
		Target:             plan.Target.ValueString(),
		Action:             plan.Action.ValueString(),
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &rule.Tags, false)...)
	}

	// Include. ValidateConfig cannot see values that are still unknown at validate
	// time (dynamic blocks), so "exactly one block" is enforced here as well.
	if plan.Include.IsNull() {
		diags.AddAttributeError(
			path.Root("include"),
			"Invalid include configuration",
			"Exactly one 'include' block must be specified.",
		)
	}
	includeFilter, d := extractTagFilter(ctx, plan.Include)
	diags.Append(d...)
	rule.Include = includeFilter

	// Exclude
	if plan.Exclude.IsNull() {
		diags.AddAttributeError(
			path.Root("exclude"),
			"Invalid exclude configuration",
			"Exactly one 'exclude' block must be specified.",
		)
	}
	excludeFilter, d := extractTagFilter(ctx, plan.Exclude)
	diags.Append(d...)
	rule.Exclude = excludeFilter

	return rule, diags
}
