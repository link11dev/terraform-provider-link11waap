package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

// automaticRegionValue is the default value applied to any region not
// explicitly set by the user.
const automaticRegionValue = "automatic"

// validRegionValues returns the set of values a region entry may be set to:
// "automatic" or a redirect to another known city code.
func validRegionValues() []string {
	values := make([]string, 0, len(knownRegions)+1)
	values = append(values, automaticRegionValue)
	values = append(values, knownRegions...)
	return values
}

// regionsMapToStringMap converts the configured regions map attribute into
// a plain Go map for sending to the API. Returns an empty map if the
// attribute is null or unknown.
func regionsMapToStringMap(ctx context.Context, m types.Map) (map[string]string, diag.Diagnostics) {
	result := make(map[string]string, len(knownRegions))
	if m.IsNull() || m.IsUnknown() {
		return result, nil
	}
	diags := m.ElementsAs(ctx, &result, false)
	return result, diags
}

// fullRegionsMap returns a copy of configured with every known region
// present, defaulting any region not explicitly configured to "automatic".
// This gives the resource exclusive ownership of all known regions on the
// API side, even though Terraform state only tracks the keys the user
// configured (see refreshTrackedRegions).
func fullRegionsMap(configured map[string]string) map[string]string {
	result := make(map[string]string, len(knownRegions))
	for _, region := range knownRegions {
		if v, ok := configured[region]; ok {
			result[region] = v
		} else {
			result[region] = automaticRegionValue
		}
	}
	return result
}

// refreshTrackedRegions rebuilds the regions map value using only the keys
// already tracked in priorRegions, refreshed with their current values from
// the API. Keys the user never configured are intentionally left out of
// state so that regions outside the user's config don't produce a
// perpetual diff.
func refreshTrackedRegions(ctx context.Context, priorRegions types.Map, apiRegions map[string]string) (types.Map, diag.Diagnostics) {
	if priorRegions.IsNull() || priorRegions.IsUnknown() {
		return priorRegions, nil
	}
	tracked := make(map[string]string, len(priorRegions.Elements()))
	for region := range priorRegions.Elements() {
		if v, ok := apiRegions[region]; ok {
			tracked[region] = v
		} else {
			tracked[region] = automaticRegionValue
		}
	}
	return types.MapValueFrom(ctx, types.StringType, tracked)
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
				Description: fmt.Sprintf(
					"Region values keyed by city code (%s). Any region not explicitly set defaults to %q on the API side, but only the keys you configure here are tracked in state.",
					strings.Join(knownRegions, ", "), automaticRegionValue,
				),
				Optional:    true,
				ElementType: types.StringType,
				Validators: []validator.Map{
					mapvalidator.KeysAre(stringvalidator.OneOf(knownRegions...)),
					mapvalidator.ValueStringsAre(stringvalidator.OneOf(validRegionValues()...)),
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

	configuredRegions, diags := regionsMapToStringMap(ctx, plan.Regions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.LoadBalancerRegionsUpdateRequest{
		LBs: []client.LoadBalancerRegionUpdate{
			{
				ID:      plan.LBID.ValueString(),
				Regions: fullRegionsMap(configuredRegions),
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

	// Read back to get computed values. The regions attribute itself is
	// left untouched: since it isn't Computed, its final state value must
	// match the planned (configured) value exactly.
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

			regionsMap, diags := refreshTrackedRegions(ctx, state.Regions, lb.Regions)
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

	configuredRegions, diags := regionsMapToStringMap(ctx, plan.Regions)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &client.LoadBalancerRegionsUpdateRequest{
		LBs: []client.LoadBalancerRegionUpdate{
			{
				ID:      plan.LBID.ValueString(),
				Regions: fullRegionsMap(configuredRegions),
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

	// Read back to get computed values. The regions attribute itself is
	// left untouched: since it isn't Computed, its final state value must
	// match the planned (configured) value exactly.
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
