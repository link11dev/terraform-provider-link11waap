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
	_ resource.Resource                = &EdgeFunctionResource{}
	_ resource.ResourceWithImportState = &EdgeFunctionResource{}
)

// EdgeFunctionResource implements the edge function resource.
type EdgeFunctionResource struct {
	client *client.Client
}

// EdgeFunctionResourceModel describes the edge function resource data model.
type EdgeFunctionResourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Code        types.String `tfsdk:"code"`
	Phase       types.String `tfsdk:"phase"`
}

// NewEdgeFunctionResource creates a new edge function resource instance.
func NewEdgeFunctionResource() resource.Resource {
	return &EdgeFunctionResource{}
}

// Metadata returns the resource type name.
func (r *EdgeFunctionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_function"
}

// Schema defines the schema for the edge function resource.
func (r *EdgeFunctionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Edge Function in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the edge function.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the edge function.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the edge function.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"code": schema.StringAttribute{
				Description: "The Lua source code for the edge function.",
				Required:    true,
			},
			"phase": schema.StringAttribute{
				Description: "The phase at which the edge function executes. Valid values: request, response.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("request_post", "response_post", "request_pre", "response_pre"),
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *EdgeFunctionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new edge function resource.
func (r *EdgeFunctionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan EdgeFunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateIDNoDash())

	ef := &client.EdgeFunction{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Code:        plan.Code.ValueString(),
		Phase:       plan.Phase.ValueString(),
	}

	err := r.client.CreateEdgeFunction(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), ef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Edge Function",
			"Could not create edge function: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the edge function resource.
func (r *EdgeFunctionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state EdgeFunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ef, err := r.client.GetEdgeFunction(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Edge Function",
			"Could not read edge function: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(ef.Name)
	state.Description = types.StringValue(ef.Description)
	state.Code = types.StringValue(ef.Code)
	state.Phase = types.StringValue(ef.Phase)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the edge function resource.
func (r *EdgeFunctionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan EdgeFunctionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ef := &client.EdgeFunction{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Code:        plan.Code.ValueString(),
		Phase:       plan.Phase.ValueString(),
	}

	err := r.client.UpdateEdgeFunction(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), ef)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Edge Function",
			"Could not update edge function: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the edge function resource.
func (r *EdgeFunctionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state EdgeFunctionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteEdgeFunction(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Edge Function",
			"Could not delete edge function: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing edge function resource.
func (r *EdgeFunctionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/edge_function_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
