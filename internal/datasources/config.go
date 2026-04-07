package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ConfigDataSource{}

// ConfigDataSource defines the data source for retrieving configuration information.
type ConfigDataSource struct {
	client *client.Client
}

// ConfigDataSourceModel describes the data model for the config data source.
type ConfigDataSourceModel struct {
	Description types.String `tfsdk:"description"`
	ID          types.String `tfsdk:"id"`
	Version     types.String `tfsdk:"version"`
	Date        types.String `tfsdk:"date"`
}

// NewConfigDataSource creates a new config data source instance.
func NewConfigDataSource() datasource.DataSource {
	return &ConfigDataSource{}
}

// Metadata returns the data source type name.
func (d *ConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

// Schema defines the schema for the config data source.
func (d *ConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves configuration information from Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Description: "Optional filter by description.",
				Optional:    true,
			},
			"id": schema.StringAttribute{
				Description: "Configuration ID.",
				Computed:    true,
			},
			"version": schema.StringAttribute{
				Description: "Configuration version.",
				Computed:    true,
			},
			"date": schema.StringAttribute{
				Description: "Last modification date.",
				Computed:    true,
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *ConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the config data source.
func (d *ConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ConfigDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	configs, err := d.client.ListConfigs(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Configurations",
			"Could not read configurations: "+err.Error(),
		)
		return
	}

	if len(configs) == 0 {
		resp.Diagnostics.AddError(
			"No Configurations Found",
			"No configurations found in the account.",
		)
		return
	}

	// Filter by description if provided
	var config *client.WAAPConfig
	if !data.Description.IsNull() {
		filterDesc := data.Description.ValueString()
		for i := range configs {
			if configs[i].Description == filterDesc {
				config = &configs[i]
				break
			}
		}
		if config == nil {
			resp.Diagnostics.AddError(
				"Configuration Not Found",
				"No configuration found with description: "+filterDesc,
			)
			return
		}
	} else {
		// Return first config if no filter
		config = &configs[0]
	}

	data.ID = types.StringValue(config.ID)
	data.Description = types.StringValue(config.Description)
	data.Version = types.StringValue(config.Version)
	if !config.Date.IsZero() {
		data.Date = types.StringValue(config.Date.Format("2006-01-02T15:04:05Z"))
	} else {
		data.Date = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
