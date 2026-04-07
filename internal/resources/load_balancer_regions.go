package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &LoadBalancerRegionsResource{}
	_ resource.ResourceWithImportState = &LoadBalancerRegionsResource{}
)

// knownRegions is the list of all known load balancer region codes.
var knownRegions = []string{"ams", "ash", "ffm", "hkg", "lax", "lon", "nyc", "sgp", "stl"}

// allRegionsDefaultModifier is a plan modifier that fills in missing region keys
// with "automatic" as the default value, so that user configs specifying a subset
// of regions do not produce perpetual diffs against the full API response.
type allRegionsDefaultModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m allRegionsDefaultModifier) Description(_ context.Context) string {
	return "Fills in missing region keys with the default value 'automatic'."
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m allRegionsDefaultModifier) MarkdownDescription(_ context.Context) string {
	return "Fills in missing region keys with the default value `automatic`."
}

// PlanModifyMap implements the plan modification logic for the regions map.
func (m allRegionsDefaultModifier) PlanModifyMap(_ context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	// If the plan value is null or unknown, do nothing.
	if req.PlanValue.IsNull() || req.PlanValue.IsUnknown() {
		return
	}

	elements := req.PlanValue.Elements()
	newElements := make(map[string]attr.Value, len(knownRegions))

	// Copy existing plan values.
	for k, v := range elements {
		newElements[k] = v
	}

	// Fill in missing known regions with "automatic".
	for _, region := range knownRegions {
		if _, exists := newElements[region]; !exists {
			newElements[region] = types.StringValue("automatic")
		}
	}

	newMap, diags := types.MapValue(types.StringType, newElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.PlanValue = newMap
}

// LoadBalancerRegionsResource implements the load balancer regions resource.
type LoadBalancerRegionsResource struct {
	client *client.Client
}

// LoadBalancerRegionsResourceModel describes the load balancer regions resource data model.
type LoadBalancerRegionsResourceModel struct {
	ConfigID types.String `tfsdk:"config_id"`
	LBID     types.String `tfsdk:"lb_id"`
	Regions  types.Map    `tfsdk:"regions"`
	// Computed
	Name            types.String `tfsdk:"name"`
	UpstreamRegions types.List   `tfsdk:"upstream_regions"`
}

// NewLoadBalancerRegionsResource creates a new load balancer regions resource instance.
func NewLoadBalancerRegionsResource() resource.Resource {
	return &LoadBalancerRegionsResource{}
}

// Metadata returns the resource type name.
func (r *LoadBalancerRegionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_regions"
}

// Schema defines the schema for the load balancer regions resource.
func (r *LoadBalancerRegionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages load balancer region configuration in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"lb_id": schema.StringAttribute{
				Description: "The load balancer ID.",
				Required:    true,
			},
			"regions": schema.MapAttribute{
				Description: "Map of city codes to region values. Missing keys default to 'automatic'.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.Map{
					allRegionsDefaultModifier{},
				},
			},
			"name": schema.StringAttribute{
				Description: "Load balancer name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"upstream_regions": schema.ListAttribute{
				Description: "List of upstream regions.",
				Computed:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *LoadBalancerRegionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new load balancer regions resource.
func (r *LoadBalancerRegionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadBalancerRegionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert regions map
	var regionsMap map[string]string
	resp.Diagnostics.Append(plan.Regions.ElementsAs(ctx, &regionsMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.LoadBalancerRegionsUpdateRequest{
		LBs: []client.LoadBalancerRegionUpdate{
			{
				ID:      plan.LBID.ValueString(),
				Regions: regionsMap,
			},
		},
	}

	err := r.client.UpdateLoadBalancerRegions(ctx, plan.ConfigID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Load Balancer Regions",
			"Could not update load balancer regions: "+err.Error(),
		)
		return
	}

	// Read back to get computed values
	lbRegions, err := r.client.GetLoadBalancerRegions(ctx, plan.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancer Regions",
			"Could not read load balancer regions after create: "+err.Error(),
		)
		return
	}

	// Find our LB in the response
	for _, lb := range lbRegions.LBs {
		if lb.ID == plan.LBID.ValueString() {
			plan.Name = types.StringValue(lb.Name)
			upstreamRegions, diags := types.ListValueFrom(ctx, types.StringType, lb.UpstreamRegions)
			resp.Diagnostics.Append(diags...)
			plan.UpstreamRegions = upstreamRegions
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the load balancer regions resource.
func (r *LoadBalancerRegionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadBalancerRegionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbRegions, err := r.client.GetLoadBalancerRegions(ctx, state.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancer Regions",
			"Could not read load balancer regions: "+err.Error(),
		)
		return
	}

	// Find our LB in the response
	found := false
	for _, lb := range lbRegions.LBs {
		if lb.ID == state.LBID.ValueString() {
			found = true
			state.Name = types.StringValue(lb.Name)

			// Convert regions map
			regionsMapValues := make(map[string]attr.Value)
			for k, v := range lb.Regions {
				regionsMapValues[k] = types.StringValue(v)
			}
			regionsMap, diags := types.MapValue(types.StringType, regionsMapValues)
			resp.Diagnostics.Append(diags...)
			state.Regions = regionsMap

			upstreamRegions, diags := types.ListValueFrom(ctx, types.StringType, lb.UpstreamRegions)
			resp.Diagnostics.Append(diags...)
			state.UpstreamRegions = upstreamRegions
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the load balancer regions resource.
func (r *LoadBalancerRegionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LoadBalancerRegionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert regions map
	var regionsMap map[string]string
	resp.Diagnostics.Append(plan.Regions.ElementsAs(ctx, &regionsMap, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.LoadBalancerRegionsUpdateRequest{
		LBs: []client.LoadBalancerRegionUpdate{
			{
				ID:      plan.LBID.ValueString(),
				Regions: regionsMap,
			},
		},
	}

	err := r.client.UpdateLoadBalancerRegions(ctx, plan.ConfigID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Load Balancer Regions",
			"Could not update load balancer regions: "+err.Error(),
		)
		return
	}

	// Read back to get computed values
	lbRegions, err := r.client.GetLoadBalancerRegions(ctx, plan.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancer Regions",
			"Could not read load balancer regions after update: "+err.Error(),
		)
		return
	}

	// Find our LB in the response
	for _, lb := range lbRegions.LBs {
		if lb.ID == plan.LBID.ValueString() {
			plan.Name = types.StringValue(lb.Name)
			upstreamRegions, diags := types.ListValueFrom(ctx, types.StringType, lb.UpstreamRegions)
			resp.Diagnostics.Append(diags...)
			plan.UpstreamRegions = upstreamRegions
			break
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the load balancer regions resource.
func (r *LoadBalancerRegionsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Load balancer regions can't really be deleted, just reset to empty
}

// ImportState imports an existing load balancer regions resource.
func (r *LoadBalancerRegionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: config_id/lb_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/lb_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lb_id"), parts[1])...)
}
