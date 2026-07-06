package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &LoadBalancerRegionsResource{}
	_ resource.ResourceWithImportState = &LoadBalancerRegionsResource{}
)

// knownRegions is the list of all known load balancer region codes. Regions
// is modeled as a nested object attribute (one named, typed attribute per
// known code) rather than a free-form map so that: (1) unknown city codes
// are rejected by Terraform Core itself as unsupported arguments, and (2)
// each region can carry its own Optional+Computed default independently,
// which a flat map attribute cannot do (Terraform Core requires a planned
// map value to match the config value exactly once the config is non-null).
var knownRegions = []string{"ams", "ash", "ffm", "hkg", "lax", "lon", "nyc", "sgp", "stl"}

// automaticRegionValue is the default value applied to any region not
// explicitly set by the user.
const automaticRegionValue = "automatic"

// regionAttributeTypes returns the attribute type map for the regions
// nested object, one types.StringType entry per known region code.
func regionAttributeTypes() map[string]attr.Type {
	attrTypes := make(map[string]attr.Type, len(knownRegions))
	for _, region := range knownRegions {
		attrTypes[region] = types.StringType
	}
	return attrTypes
}

// regionSchemaAttributes builds the nested schema attributes for the
// regions object, one Optional+Computed string attribute per known region
// code, each defaulting to "automatic".
func regionSchemaAttributes() map[string]schema.Attribute {
	attrs := make(map[string]schema.Attribute, len(knownRegions))
	for _, region := range knownRegions {
		attrs[region] = schema.StringAttribute{
			Description: fmt.Sprintf("Region value for city code %q. Defaults to %q.", region, automaticRegionValue),
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(automaticRegionValue),
		}
	}
	return attrs
}

// defaultRegionsObject returns the regions object value used as the
// top-level default when the whole `regions` attribute is omitted from
// config, with every known region set to "automatic".
func defaultRegionsObject() types.Object {
	values := make(map[string]attr.Value, len(knownRegions))
	for _, region := range knownRegions {
		values[region] = types.StringValue(automaticRegionValue)
	}
	obj, _ := types.ObjectValue(regionAttributeTypes(), values)
	return obj
}

// regionsObjectToMap converts a regions object value into a plain map for
// sending to the API.
func regionsObjectToMap(obj types.Object) map[string]string {
	attrs := obj.Attributes()
	result := make(map[string]string, len(attrs))
	for region, v := range attrs {
		if s, ok := v.(types.String); ok && !s.IsNull() {
			result[region] = s.ValueString()
		}
	}
	return result
}

// regionsMapToObject converts the API's regions map into a regions object
// value, defaulting any known region missing from the API response to
// "automatic".
func regionsMapToObject(regionsMap map[string]string) (types.Object, diag.Diagnostics) {
	values := make(map[string]attr.Value, len(knownRegions))
	for _, region := range knownRegions {
		if v, ok := regionsMap[region]; ok {
			values[region] = types.StringValue(v)
		} else {
			values[region] = types.StringValue(automaticRegionValue)
		}
	}
	return types.ObjectValue(regionAttributeTypes(), values)
}

// LoadBalancerRegionsResource implements the load balancer regions resource.
type LoadBalancerRegionsResource struct {
	client *client.Client
}

// LoadBalancerRegionsResourceModel describes the load balancer regions resource data model.
type LoadBalancerRegionsResourceModel struct {
	ConfigID types.String `tfsdk:"config_id"`
	LBID     types.String `tfsdk:"lb_id"`
	Regions  types.Object `tfsdk:"regions"`
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
			"regions": schema.SingleNestedAttribute{
				Description: fmt.Sprintf(
					"Region values keyed by city code (%s). Any region not explicitly set defaults to %q.",
					strings.Join(knownRegions, ", "), automaticRegionValue,
				),
				Optional:   true,
				Computed:   true,
				Attributes: regionSchemaAttributes(),
				Default:    objectdefault.StaticValue(defaultRegionsObject()),
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

	updateReq := &client.LoadBalancerRegionsUpdateRequest{
		LBs: []client.LoadBalancerRegionUpdate{
			{
				ID:      plan.LBID.ValueString(),
				Regions: regionsObjectToMap(plan.Regions),
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
			regionsObj, diags := regionsMapToObject(lb.Regions)
			resp.Diagnostics.Append(diags...)
			plan.Regions = regionsObj
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

			regionsObj, diags := regionsMapToObject(lb.Regions)
			resp.Diagnostics.Append(diags...)
			state.Regions = regionsObj

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

	updateReq := &client.LoadBalancerRegionsUpdateRequest{
		LBs: []client.LoadBalancerRegionUpdate{
			{
				ID:      plan.LBID.ValueString(),
				Regions: regionsObjectToMap(plan.Regions),
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
			regionsObj, diags := regionsMapToObject(lb.Regions)
			resp.Diagnostics.Append(diags...)
			plan.Regions = regionsObj
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
