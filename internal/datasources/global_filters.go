// Package datasources contains the data source implementations for the Link11 WAAP Terraform provider.
package datasources

import (
	"context"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &GlobalFiltersDataSource{}

// GlobalFiltersDataSource defines the data source for listing global filters.
type GlobalFiltersDataSource struct {
	client *client.Client
}

// GlobalFiltersDataSourceModel describes the data model for the global filters data source.
type GlobalFiltersDataSourceModel struct {
	ConfigID      types.String              `tfsdk:"config_id"`
	GlobalFilters []GlobalFilterDataModel `tfsdk:"global_filters"`
}

// GlobalFilterDataModel represents a single global filter in the data source.
type GlobalFilterDataModel struct {
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

// NewGlobalFiltersDataSource creates a new global filters data source instance.
func NewGlobalFiltersDataSource() datasource.DataSource {
	return &GlobalFiltersDataSource{}
}

// Metadata returns the data source type name.
func (d *GlobalFiltersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_filters"
}

// Schema defines the schema for the global filters data source.
func (d *GlobalFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all global filters in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"global_filters": schema.ListNestedAttribute{
				Description: "List of global filters.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"source":      schema.StringAttribute{Computed: true},
						"mdate":       schema.StringAttribute{Computed: true},
						"active":      schema.BoolAttribute{Computed: true},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"action":      schema.StringAttribute{Computed: true},
						"rule":        schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *GlobalFiltersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the global filters data source.
func (d *GlobalFiltersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GlobalFiltersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filters, err := d.client.ListGlobalFilters(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Global Filters",
			"Could not read global filters: "+err.Error(),
		)
		return
	}

	data.GlobalFilters = make([]GlobalFilterDataModel, len(filters))
	for i, f := range filters {
		model := GlobalFilterDataModel{
			ID:          types.StringValue(f.ID),
			Name:        types.StringValue(f.Name),
			Description: types.StringValue(f.Description),
			Source:      types.StringValue(f.Source),
			Mdate:       types.StringValue(f.Mdate),
			Active:      types.BoolValue(f.Active),
		}

		model.Tags = dsStringSliceToList(ctx, f.Tags, resp)

		// Action: interface{} -> string
		if f.Action != nil {
			if actionStr, ok := f.Action.(string); ok {
				model.Action = types.StringValue(actionStr)
			} else {
				actionBytes, marshalErr := json.Marshal(f.Action)
				if marshalErr != nil {
					resp.Diagnostics.AddError("Error Marshaling Action", marshalErr.Error())
					return
				}
				model.Action = types.StringValue(string(actionBytes))
			}
		} else {
			model.Action = types.StringValue("")
		}

		// Rule: interface{} -> JSON string
		if f.Rule != nil {
			ruleBytes, marshalErr := json.Marshal(f.Rule)
			if marshalErr != nil {
				resp.Diagnostics.AddError("Error Marshaling Rule", marshalErr.Error())
				return
			}
			model.Rule = types.StringValue(string(ruleBytes))
		} else {
			model.Rule = types.StringValue("{}")
		}

		data.GlobalFilters[i] = model
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
