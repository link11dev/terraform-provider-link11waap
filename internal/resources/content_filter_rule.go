package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &ContentFilterRuleResource{}
	_ resource.ResourceWithImportState = &ContentFilterRuleResource{}
)

// ContentFilterRuleResource implements the content filter rule resource.
type ContentFilterRuleResource struct {
	client *client.Client
}

// ContentFilterRuleResourceModel describes the content filter rule resource data model.
type ContentFilterRuleResourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Msg         types.String `tfsdk:"msg"`
	Operand     types.String `tfsdk:"operand"`
	Category    types.String `tfsdk:"category"`
	Subcategory types.String `tfsdk:"subcategory"`
	Risk        types.Int64  `tfsdk:"risk"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
}

// NewContentFilterRuleResource creates a new content filter rule resource instance.
func NewContentFilterRuleResource() resource.Resource {
	return &ContentFilterRuleResource{}
}

// Metadata returns the resource type name.
func (r *ContentFilterRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_filter_rule"
}

// Schema defines the schema for the content filter rule resource.
func (r *ContentFilterRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Content Filter Rule in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the content filter rule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the content filter rule.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"msg": schema.StringAttribute{
				Description: "Log message for this rule.",
				Required:    true,
			},
			"operand": schema.StringAttribute{
				Description: "Matching domain(s) regex.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"category": schema.StringAttribute{
				Description: "Category of the rule.",
				Required:    true,
			},
			"subcategory": schema.StringAttribute{
				Description: "Subcategory of the rule.",
				Required:    true,
			},
			"risk": schema.Int64Attribute{
				Description: "Risk level of this rule, between 1 (lowest risk) and 5 (highest risk).",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 5),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the content filter rule.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Description: "List of tags to apply.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *ContentFilterRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new content filter rule resource.
func (r *ContentFilterRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContentFilterRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateIDNoDash())

	rule := r.buildContentFilterRule(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateContentFilterRule(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Content Filter Rule",
			"Could not create content filter rule: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the content filter rule resource.
func (r *ContentFilterRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContentFilterRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetContentFilterRule(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Content Filter Rule",
			"Could not read content filter rule: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(rule.Name)
	state.Msg = types.StringValue(rule.Msg)
	state.Operand = types.StringValue(rule.Operand)
	state.Category = types.StringValue(rule.Category)
	state.Subcategory = types.StringValue(rule.Subcategory)
	state.Risk = types.Int64Value(int64(rule.Risk))
	state.Description = types.StringValue(rule.Description)

	if len(rule.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, rule.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the content filter rule resource.
func (r *ContentFilterRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContentFilterRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := r.buildContentFilterRule(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateContentFilterRule(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), rule)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Content Filter Rule",
			"Could not update content filter rule: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the content filter rule resource.
func (r *ContentFilterRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContentFilterRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteContentFilterRule(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Content Filter Rule",
			"Could not delete content filter rule: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing content filter rule resource.
func (r *ContentFilterRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/content_filter_rule_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildContentFilterRule maps the plan model into the API client struct.
func (r *ContentFilterRuleResource) buildContentFilterRule(ctx context.Context, plan *ContentFilterRuleResourceModel, diags *diag.Diagnostics) *client.ContentFilterRule {
	rule := &client.ContentFilterRule{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Msg:         plan.Msg.ValueString(),
		Operand:     plan.Operand.ValueString(),
		Category:    plan.Category.ValueString(),
		Subcategory: plan.Subcategory.ValueString(),
		Risk:        int(plan.Risk.ValueInt64()),
		Description: plan.Description.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &rule.Tags, false)...)
	}

	return rule
}
