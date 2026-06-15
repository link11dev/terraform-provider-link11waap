package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &ActionResource{}
	_ resource.ResourceWithImportState = &ActionResource{}
)

// ActionResource implements the action resource.
type ActionResource struct {
	client *client.Client
}

// ActionResourceModel describes the action resource data model.
type ActionResourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Tags        types.List   `tfsdk:"tags"`
	Params      types.Object `tfsdk:"params"`
}

// actionParamsAttrTypes is the attr.Type map for the params object.
func actionParamsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"content": types.StringType,
		"status":  types.Int64Type,
		"headers": types.MapType{ElemType: types.StringType},
	}
}

// NewActionResource creates a new action resource instance.
func NewActionResource() resource.Resource {
	return &ActionResource{}
}

// Metadata returns the resource type name.
func (r *ActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_action"
}

// Schema defines the schema for the action resource.
func (r *ActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an Action in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the action.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the action.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the action.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"type": schema.StringAttribute{
				Description: "Action type. One of: skip, block, challenge, ichallenge, monitor.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("skip", "block", "challenge", "ichallenge", "monitor"),
				},
			},
			"tags": schema.ListAttribute{
				Description: "List of tags to apply.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"params": schema.SingleNestedAttribute{
				Description: "Action parameters.",
				Optional:    true,
				Computed:    true,
				Attributes: map[string]schema.Attribute{
					"content": schema.StringAttribute{
						Description: "Response body content.",
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
					},
					"status": schema.Int64Attribute{
						Description: "HTTP status code to return.",
						Optional:    true,
					},
					"headers": schema.MapAttribute{
						Description: "Map of response header name to value.",
						Optional:    true,
						ElementType: types.StringType,
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *ActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new action resource.
func (r *ActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateIDNoDash())

	a := r.buildAction(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateAction(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), a)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Action",
			"Could not create action: "+err.Error(),
		)
		return
	}

	r.flattenAction(ctx, a, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the action resource.
func (r *ActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	a, err := r.client.GetAction(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Action",
			"Could not read action: "+err.Error(),
		)
		return
	}

	r.flattenAction(ctx, a, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the action resource.
func (r *ActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	a := r.buildAction(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateAction(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), a)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Action",
			"Could not update action: "+err.Error(),
		)
		return
	}

	r.flattenAction(ctx, a, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the action resource.
func (r *ActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteAction(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Action",
			"Could not delete action: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing action resource.
func (r *ActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/action_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildAction maps the plan model into the API client struct.
func (r *ActionResource) buildAction(ctx context.Context, plan *ActionResourceModel, diags *diag.Diagnostics) *client.Action {
	a := &client.Action{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Type:        plan.Type.ValueString(),
	}

	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &a.Tags, false)...)
	}

	if !plan.Params.IsNull() && !plan.Params.IsUnknown() {
		var pm struct {
			Content types.String `tfsdk:"content"`
			Status  types.Int64  `tfsdk:"status"`
			Headers types.Map    `tfsdk:"headers"`
		}
		diags.Append(plan.Params.As(ctx, &pm, basetypes.ObjectAsOptions{})...)

		params := &client.ActionParams{Content: pm.Content.ValueString()}
		if !pm.Status.IsNull() && !pm.Status.IsUnknown() {
			s := int(pm.Status.ValueInt64())
			params.Status = &s
		}
		if !pm.Headers.IsNull() && !pm.Headers.IsUnknown() {
			diags.Append(pm.Headers.ElementsAs(ctx, &params.Headers, false)...)
		}
		a.Params = params
	}

	return a
}

// flattenAction maps the API client struct back into the resource state model.
func (r *ActionResource) flattenAction(ctx context.Context, a *client.Action, state *ActionResourceModel, diags *diag.Diagnostics) {
	state.Name = types.StringValue(a.Name)
	state.Description = types.StringValue(a.Description)
	state.Type = types.StringValue(a.Type)

	if len(a.Tags) > 0 {
		tagsList, d := types.ListValueFrom(ctx, types.StringType, a.Tags)
		diags.Append(d...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	if a.Params != nil {
		var headers types.Map
		if a.Params.Headers != nil {
			hv, d := types.MapValueFrom(ctx, types.StringType, a.Params.Headers)
			diags.Append(d...)
			headers = hv
		} else {
			headers = types.MapNull(types.StringType)
		}
		status := types.Int64Null()
		if a.Params.Status != nil {
			status = types.Int64Value(int64(*a.Params.Status))
		}
		obj, d := types.ObjectValue(actionParamsAttrTypes(), map[string]attr.Value{
			"content": types.StringValue(a.Params.Content),
			"status":  status,
			"headers": headers,
		})
		diags.Append(d...)
		state.Params = obj
	} else {
		state.Params = types.ObjectNull(actionParamsAttrTypes())
	}
}
