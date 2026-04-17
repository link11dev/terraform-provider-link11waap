// Package datasources contains the data source implementations for the Link11 WAAP Terraform provider.
package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &PlanetTrustedNetsDataSource{}

// PlanetTrustedNetsDataSource defines the data source for reading planet trusted nets.
type PlanetTrustedNetsDataSource struct {
	client *client.Client
}

// PlanetTrustedNetsDataSourceModel describes the data model.
type PlanetTrustedNetsDataSourceModel struct {
	ConfigID    types.String          `tfsdk:"config_id"`
	ID          types.String          `tfsdk:"id"`
	Name        types.String          `tfsdk:"name"`
	TrustedNets []TrustedNetDataModel `tfsdk:"trusted_nets"`
}

// TrustedNetDataModel represents a single trusted net entry in the data source.
type TrustedNetDataModel struct {
	Source  types.String `tfsdk:"source"`
	Address types.String `tfsdk:"address"`
	GfID    types.String `tfsdk:"gf_id"`
	Comment types.String `tfsdk:"comment"`
}

// NewPlanetTrustedNetsDataSource creates a new data source instance.
func NewPlanetTrustedNetsDataSource() datasource.DataSource {
	return &PlanetTrustedNetsDataSource{}
}

// Metadata returns the data source type name.
func (d *PlanetTrustedNetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_planet_trusted_nets"
}

// Schema defines the schema for the data source.
func (d *PlanetTrustedNetsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Reads the trusted networks list from the planet.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The planet entry ID (always __default__).",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The planet entry name (always __default__).",
				Computed:    true,
			},
			"trusted_nets": schema.ListNestedAttribute{
				Description: "List of trusted network entries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "Source type: 'ip' or 'global_filter'.",
							Computed:    true,
						},
						"address": schema.StringAttribute{
							Description: "IP address or CIDR block.",
							Computed:    true,
						},
						"gf_id": schema.StringAttribute{
							Description: "Global filter ID.",
							Computed:    true,
						},
						"comment": schema.StringAttribute{
							Description: "Human-readable comment.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *PlanetTrustedNetsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the planet trusted nets data source.
func (d *PlanetTrustedNetsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PlanetTrustedNetsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planet, err := d.client.GetPlanet(ctx, data.ConfigID.ValueString(), "__default__")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Planet Trusted Nets",
			"Could not read planet trusted nets: "+err.Error(),
		)
		return
	}

	data.ID = types.StringValue(planet.ID)
	data.Name = types.StringValue(planet.Name)

	data.TrustedNets = make([]TrustedNetDataModel, len(planet.TrustedNets))
	for i, tn := range planet.TrustedNets {
		data.TrustedNets[i] = TrustedNetDataModel{
			Source:  types.StringValue(tn.Source),
			Address: types.StringValue(tn.Address),
			GfID:    types.StringValue(tn.GfID),
			Comment: types.StringValue(tn.Comment),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
