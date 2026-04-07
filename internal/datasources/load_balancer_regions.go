package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &LoadBalancerRegionsDataSource{}

// LoadBalancerRegionsDataSource defines the data source for retrieving load balancer regions.
type LoadBalancerRegionsDataSource struct {
	client *client.Client
}

// LoadBalancerRegionsDataSourceModel describes the data model for the load balancer regions data source.
type LoadBalancerRegionsDataSourceModel struct {
	ConfigID  types.String                  `tfsdk:"config_id"`
	CityCodes types.Map                     `tfsdk:"city_codes"`
	LBs       []LoadBalancerRegionDataModel `tfsdk:"lbs"`
}

// LoadBalancerRegionDataModel represents a single load balancer region in the data source.
type LoadBalancerRegionDataModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Regions         types.Map    `tfsdk:"regions"`
	UpstreamRegions types.List   `tfsdk:"upstream_regions"`
}

// NewLoadBalancerRegionsDataSource creates a new load balancer regions data source instance.
func NewLoadBalancerRegionsDataSource() datasource.DataSource {
	return &LoadBalancerRegionsDataSource{}
}

// Metadata returns the data source type name.
func (d *LoadBalancerRegionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_regions"
}

// Schema defines the schema for the load balancer regions data source.
func (d *LoadBalancerRegionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves load balancer region configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"city_codes": schema.MapAttribute{
				Description: "Map of city codes to names.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"lbs": schema.ListNestedAttribute{
				Description: "List of load balancer region configurations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":               schema.StringAttribute{Computed: true},
						"name":             schema.StringAttribute{Computed: true},
						"regions":          schema.MapAttribute{Computed: true, ElementType: types.StringType},
						"upstream_regions": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *LoadBalancerRegionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the load balancer regions data source.
func (d *LoadBalancerRegionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LoadBalancerRegionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	lbRegions, err := d.client.GetLoadBalancerRegions(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancer Regions",
			"Could not read load balancer regions: "+err.Error(),
		)
		return
	}

	// Convert city codes map
	cityCodesMapValues := make(map[string]attr.Value)
	for k, v := range lbRegions.CityCodes {
		cityCodesMapValues[k] = types.StringValue(v)
	}
	cityCodesMap, diags := types.MapValue(types.StringType, cityCodesMapValues)
	resp.Diagnostics.Append(diags...)
	data.CityCodes = cityCodesMap

	// Convert LBs
	data.LBs = make([]LoadBalancerRegionDataModel, len(lbRegions.LBs))
	for i, lb := range lbRegions.LBs {
		// Convert regions map
		regionsMapValues := make(map[string]attr.Value)
		for k, v := range lb.Regions {
			regionsMapValues[k] = types.StringValue(v)
		}
		regionsMap, diags := types.MapValue(types.StringType, regionsMapValues)
		resp.Diagnostics.Append(diags...)

		upstreamRegions, diags := types.ListValueFrom(ctx, types.StringType, lb.UpstreamRegions)
		resp.Diagnostics.Append(diags...)

		data.LBs[i] = LoadBalancerRegionDataModel{
			ID:              types.StringValue(lb.ID),
			Name:            types.StringValue(lb.Name),
			Regions:         regionsMap,
			UpstreamRegions: upstreamRegions,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
