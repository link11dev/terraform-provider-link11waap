package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &LoadBalancersDataSource{}

// LoadBalancersDataSource defines the data source for listing load balancers in a configuration
type LoadBalancersDataSource struct {
	client *client.Client
}

// LoadBalancersDataSourceModel represents the data model for the load balancers data source
type LoadBalancersDataSourceModel struct {
	ConfigID      types.String            `tfsdk:"config_id"`
	LoadBalancers []LoadBalancerDataModel `tfsdk:"load_balancers"`
}

// LoadBalancerDataModel represents a load balancer in the data source
type LoadBalancerDataModel struct {
	Name               types.String `tfsdk:"name"`
	Provider           types.String `tfsdk:"provider"`
	Region             types.String `tfsdk:"region"`
	DNSName            types.String `tfsdk:"dns_name"`
	ListenerName       types.String `tfsdk:"listener_name"`
	ListenerPort       types.Int64  `tfsdk:"listener_port"`
	LoadBalancerType   types.String `tfsdk:"load_balancer_type"`
	MaxCertificates    types.Int64  `tfsdk:"max_certificates"`
	DefaultCertificate types.String `tfsdk:"default_certificate"`
	Certificates       types.List   `tfsdk:"certificates"`
}

// NewLoadBalancersDataSource creates a new instance of the load balancers data source
func NewLoadBalancersDataSource() datasource.DataSource {
	return &LoadBalancersDataSource{}
}

// Metadata returns the data source type name
func (d *LoadBalancersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancers"
}

// Schema defines the schema for the load balancers data source
func (d *LoadBalancersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all load balancers in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"load_balancers": schema.ListNestedAttribute{
				Description: "List of load balancers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":                schema.StringAttribute{Computed: true},
						"provider":            schema.StringAttribute{Computed: true},
						"region":              schema.StringAttribute{Computed: true},
						"dns_name":            schema.StringAttribute{Computed: true},
						"listener_name":       schema.StringAttribute{Computed: true},
						"listener_port":       schema.Int64Attribute{Computed: true},
						"load_balancer_type":  schema.StringAttribute{Computed: true},
						"max_certificates":    schema.Int64Attribute{Computed: true},
						"default_certificate": schema.StringAttribute{Computed: true},
						"certificates":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

// Configure sets the client for the data source
func (d *LoadBalancersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the list of load balancers for the given configuration and sets the state
func (d *LoadBalancersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LoadBalancersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	loadBalancers, err := d.client.ListLoadBalancers(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancers",
			"Could not read load balancers: "+err.Error(),
		)
		return
	}

	data.LoadBalancers = make([]LoadBalancerDataModel, len(loadBalancers))
	for i, lb := range loadBalancers {
		certs, diags := types.ListValueFrom(ctx, types.StringType, lb.Certificates)
		resp.Diagnostics.Append(diags...)

		data.LoadBalancers[i] = LoadBalancerDataModel{
			Name:               types.StringValue(lb.Name),
			Provider:           types.StringValue(lb.Provider),
			Region:             types.StringValue(lb.Region),
			DNSName:            types.StringValue(lb.DNSName),
			ListenerName:       types.StringValue(lb.ListenerName),
			ListenerPort:       types.Int64Value(int64(lb.ListenerPort)),
			LoadBalancerType:   types.StringValue(lb.LoadBalancerType),
			MaxCertificates:    types.Int64Value(int64(lb.MaxCertificates)),
			DefaultCertificate: types.StringValue(lb.DefaultCertificate),
			Certificates:       certs,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
