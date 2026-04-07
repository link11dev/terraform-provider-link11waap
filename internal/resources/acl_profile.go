// Package resources contains the implementation of Terraform resources for Link11 WAAP.
package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &ACLProfileResource{}
	_ resource.ResourceWithImportState = &ACLProfileResource{}
)

// ACLProfileResource implements the ACL profile resource.
type ACLProfileResource struct {
	client *client.Client
}

// ACLProfileResourceModel describes the ACL profile resource data model.
type ACLProfileResourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Action      types.String `tfsdk:"action"`
	Allow       types.List   `tfsdk:"allow"`
	AllowBot    types.List   `tfsdk:"allow_bot"`
	Deny        types.List   `tfsdk:"deny"`
	DenyBot     types.List   `tfsdk:"deny_bot"`
	ForceDeny   types.List   `tfsdk:"force_deny"`
	Passthrough types.List   `tfsdk:"passthrough"`
}

// NewACLProfileResource creates a new ACL profile resource instance.
func NewACLProfileResource() resource.Resource {
	return &ACLProfileResource{}
}

// Metadata returns the resource type name.
func (r *ACLProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_profile"
}

// Schema defines the schema for the ACL profile resource.
func (r *ACLProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an ACL Profile in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the ACL profile.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the ACL profile.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the ACL profile.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Description: "List of tags associated with the ACL profile.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "Default action for the ACL profile. Allowed values: action-acl-block, action-waap-feed-block, action-https-redirect.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("action-acl-block", "action-waap-feed-block", "action-https-redirect"),
				},
			},
			"allow": schema.ListAttribute{
				Description: "List of tag identifiers to allow.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"allow_bot": schema.ListAttribute{
				Description: "List of tag identifiers to allow (bot).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"deny": schema.ListAttribute{
				Description: "List of tag identifiers to deny.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"deny_bot": schema.ListAttribute{
				Description: "List of tag identifiers to deny (bot).",
				Optional:    true,
				ElementType: types.StringType,
			},
			"force_deny": schema.ListAttribute{
				Description: "List of tag identifiers to force deny.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"passthrough": schema.ListAttribute{
				Description: "List of tag identifiers to pass through without inspection.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *ACLProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new ACL profile resource.
func (r *ACLProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ACLProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	profile := &client.ACLProfile{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Action:      plan.Action.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &profile.Tags, false)...)
	}
	if !plan.Allow.IsNull() && !plan.Allow.IsUnknown() {
		resp.Diagnostics.Append(plan.Allow.ElementsAs(ctx, &profile.Allow, false)...)
	}
	if !plan.AllowBot.IsNull() && !plan.AllowBot.IsUnknown() {
		resp.Diagnostics.Append(plan.AllowBot.ElementsAs(ctx, &profile.AllowBot, false)...)
	}
	if !plan.Deny.IsNull() && !plan.Deny.IsUnknown() {
		resp.Diagnostics.Append(plan.Deny.ElementsAs(ctx, &profile.Deny, false)...)
	}
	if !plan.DenyBot.IsNull() && !plan.DenyBot.IsUnknown() {
		resp.Diagnostics.Append(plan.DenyBot.ElementsAs(ctx, &profile.DenyBot, false)...)
	}
	if !plan.ForceDeny.IsNull() && !plan.ForceDeny.IsUnknown() {
		resp.Diagnostics.Append(plan.ForceDeny.ElementsAs(ctx, &profile.ForceDeny, false)...)
	}
	if !plan.Passthrough.IsNull() && !plan.Passthrough.IsUnknown() {
		resp.Diagnostics.Append(plan.Passthrough.ElementsAs(ctx, &profile.Passthrough, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateACLProfile(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), profile)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating ACL Profile",
			"Could not create ACL profile: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the ACL profile resource.
func (r *ACLProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ACLProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile, err := r.client.GetACLProfile(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading ACL Profile",
			"Could not read ACL profile: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(profile.Name)
	state.Description = types.StringValue(profile.Description)
	state.Action = types.StringValue(profile.Action)

	if len(profile.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, profile.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	state.Allow = stringSliceToList(ctx, profile.Allow, resp)
	state.AllowBot = stringSliceToList(ctx, profile.AllowBot, resp)
	state.Deny = stringSliceToList(ctx, profile.Deny, resp)
	state.DenyBot = stringSliceToList(ctx, profile.DenyBot, resp)
	state.ForceDeny = stringSliceToList(ctx, profile.ForceDeny, resp)
	state.Passthrough = stringSliceToList(ctx, profile.Passthrough, resp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// stringSliceToList converts a []string from the API to a types.List for Terraform state.
func stringSliceToList(ctx context.Context, slice []string, resp *resource.ReadResponse) types.List {
	if len(slice) > 0 {
		list, diags := types.ListValueFrom(ctx, types.StringType, slice)
		resp.Diagnostics.Append(diags...)
		return list
	}
	return types.ListNull(types.StringType)
}

// Update updates the ACL profile resource.
func (r *ACLProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ACLProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profile := &client.ACLProfile{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Action:      plan.Action.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &profile.Tags, false)...)
	}
	if !plan.Allow.IsNull() && !plan.Allow.IsUnknown() {
		resp.Diagnostics.Append(plan.Allow.ElementsAs(ctx, &profile.Allow, false)...)
	}
	if !plan.AllowBot.IsNull() && !plan.AllowBot.IsUnknown() {
		resp.Diagnostics.Append(plan.AllowBot.ElementsAs(ctx, &profile.AllowBot, false)...)
	}
	if !plan.Deny.IsNull() && !plan.Deny.IsUnknown() {
		resp.Diagnostics.Append(plan.Deny.ElementsAs(ctx, &profile.Deny, false)...)
	}
	if !plan.DenyBot.IsNull() && !plan.DenyBot.IsUnknown() {
		resp.Diagnostics.Append(plan.DenyBot.ElementsAs(ctx, &profile.DenyBot, false)...)
	}
	if !plan.ForceDeny.IsNull() && !plan.ForceDeny.IsUnknown() {
		resp.Diagnostics.Append(plan.ForceDeny.ElementsAs(ctx, &profile.ForceDeny, false)...)
	}
	if !plan.Passthrough.IsNull() && !plan.Passthrough.IsUnknown() {
		resp.Diagnostics.Append(plan.Passthrough.ElementsAs(ctx, &profile.Passthrough, false)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateACLProfile(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), profile)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating ACL Profile",
			"Could not update ACL profile: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the ACL profile resource.
func (r *ACLProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ACLProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteACLProfile(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting ACL Profile",
			"Could not delete ACL profile: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing ACL profile resource.
func (r *ACLProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/acl_profile_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
