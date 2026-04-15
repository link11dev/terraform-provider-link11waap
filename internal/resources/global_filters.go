// Package resources contains the implementation of Terraform resources for Link11 WAAP.
package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &GlobalFilterResource{}
	_ resource.ResourceWithImportState = &GlobalFilterResource{}
)

// GlobalFilterResource implements the global filter resource.
type GlobalFilterResource struct {
	client *client.Client
}

// GlobalFilterResourceModel describes the global filter resource data model.
type GlobalFilterResourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	Mdate       types.String `tfsdk:"mdate"`
	Active      types.Bool   `tfsdk:"active"`
	Tags        types.List   `tfsdk:"tags"`
	Action      types.String `tfsdk:"action"`
	Rule        types.String `tfsdk:"rule"`
}

// NewGlobalFilterResource creates a new global filter resource instance.
func NewGlobalFilterResource() resource.Resource {
	return &GlobalFilterResource{}
}

// Metadata returns the resource type name.
func (r *GlobalFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_filter"
}

// Schema defines the schema for the global filter resource.
func (r *GlobalFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Global Filter in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the global filter.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the global filter.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the global filter.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"source": schema.StringAttribute{
				Description: "The source URL for the global filter.",
				Required:    true,
			},
			"mdate": schema.StringAttribute{
				Description: "The last modification date (server-managed).",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the global filter is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"tags": schema.ListAttribute{
				Description: "List of tags associated with the global filter.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "Action for the global filter. Allowed values: action-challenge, action-monitor, action-skip, action-global-filter-block, action-rate-limit-block, action-acl-block, action-contentfilter-block, action-dynamic-rule-block, action-waap-feed-block, action-https-redirect.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"",
						"action-challenge",
						"action-monitor",
						"action-skip",
						"action-global-filter-block",
						"action-rate-limit-block",
						"action-acl-block",
						"action-contentfilter-block",
						"action-dynamic-rule-block",
						"action-waap-feed-block",
						"action-https-redirect",
					),
				},
			},
			"rule": schema.StringAttribute{
				Description: "The rule definition as a JSON string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("{}"),
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *GlobalFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new global filter resource.
func (r *GlobalFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GlobalFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	filter, diags := buildGlobalFilterAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateGlobalFilter(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Global Filter",
			"Could not create global filter: "+err.Error(),
		)
		return
	}

	// mdate is server-managed, set empty after create
	plan.Mdate = types.StringValue("")

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the global filter resource.
func (r *GlobalFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, err := r.client.GetGlobalFilter(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Global Filter",
			"Could not read global filter: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(filter.Name)
	state.Description = types.StringValue(filter.Description)
	state.Source = types.StringValue(filter.Source)
	state.Mdate = types.StringValue(filter.Mdate)
	state.Active = types.BoolValue(filter.Active)

	// Tags
	if len(filter.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, filter.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	// Action: interface{} -> string
	if filter.Action != nil {
		if actionStr, ok := filter.Action.(string); ok {
			state.Action = types.StringValue(actionStr)
		} else {
			// Complex action type - marshal to JSON
			actionBytes, marshalErr := json.Marshal(filter.Action)
			if marshalErr != nil {
				resp.Diagnostics.AddError("Error Marshaling Action", marshalErr.Error())
				return
			}
			state.Action = types.StringValue(string(actionBytes))
		}
	} else {
		state.Action = types.StringValue("")
	}

	// Rule: interface{} -> JSON string
	if filter.Rule != nil {
		ruleBytes, marshalErr := json.Marshal(filter.Rule)
		if marshalErr != nil {
			resp.Diagnostics.AddError("Error Marshaling Rule", marshalErr.Error())
			return
		}
		state.Rule = types.StringValue(string(ruleBytes))
	} else {
		state.Rule = types.StringValue("{}")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the global filter resource.
func (r *GlobalFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GlobalFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateGlobalFilter(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Global Filter",
			"Could not update global filter: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the global filter resource.
func (r *GlobalFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlobalFilter(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Global Filter",
			"Could not delete global filter: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing global filter resource.
func (r *GlobalFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/global_filter_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildGlobalFilterAPIModel converts the Terraform resource model to the API model.
func buildGlobalFilterAPIModel(ctx context.Context, plan *GlobalFilterResourceModel) (*client.GlobalFilter, diag.Diagnostics) {
	var diags diag.Diagnostics

	filter := &client.GlobalFilter{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Source:      plan.Source.ValueString(),
		Active:      plan.Active.ValueBool(),
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &filter.Tags, false)...)
	}

	// Action: plain string
	actionStr := plan.Action.ValueString()
	if actionStr != "" {
		filter.Action = actionStr
	}

	// Rule: JSON string -> interface{}
	if !plan.Rule.IsNull() && !plan.Rule.IsUnknown() {
		ruleStr := plan.Rule.ValueString()
		if ruleStr != "" {
			var rule interface{}
			if err := json.Unmarshal([]byte(ruleStr), &rule); err != nil {
				diags.AddError("Invalid Rule JSON", fmt.Sprintf("Could not parse rule JSON: %s", err.Error()))
				return nil, diags
			}
			filter.Rule = rule
		}
	}

	return filter, diags
}
