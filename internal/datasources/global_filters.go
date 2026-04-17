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
var _ datasource.DataSource = &GlobalFilterDataSource{}

// GlobalFiltersDataSource defines the data source for listing global filters.
type GlobalFiltersDataSource struct {
	client *client.Client
}

// GlobalFiltersDataSourceModel describes the data model for the global filters data source.
type GlobalFiltersDataSourceModel struct {
	ConfigID      types.String            `tfsdk:"config_id"`
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

		model.Action = types.StringValue(f.Action)

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

// GlobalFilterDataSource defines the data source for fetching a single global filter by name.
type GlobalFilterDataSource struct {
	client *client.Client
}

// GlobalFilterSingleDataSourceModel describes the data model for the singular global filter data source.
type GlobalFilterSingleDataSourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	Name        types.String `tfsdk:"name"`
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	Mdate       types.String `tfsdk:"mdate"`
	Active      types.Bool   `tfsdk:"active"`
	Tags        types.List   `tfsdk:"tags"`
	Action      types.String `tfsdk:"action"`
	Rule        types.String `tfsdk:"rule"`
}

// NewGlobalFilterDataSource creates a new singular global filter data source instance.
func NewGlobalFilterDataSource() datasource.DataSource {
	return &GlobalFilterDataSource{}
}

// Metadata returns the data source type name.
func (d *GlobalFilterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_filter"
}

// Schema defines the schema for the singular global filter data source.
func (d *GlobalFilterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single global filter by name from a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the global filter to look up.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier of the global filter.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the global filter.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source of the global filter.",
				Computed:    true,
			},
			"mdate": schema.StringAttribute{
				Description: "Last modification date.",
				Computed:    true,
			},
			"active": schema.BoolAttribute{
				Description: "Whether the filter is active.",
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "Tags associated with the filter.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "Action taken when the filter matches.",
				Computed:    true,
			},
			"rule": schema.StringAttribute{
				Description: "JSON-encoded rule definition.",
				Computed:    true,
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *GlobalFilterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the singular global filter data source by looking up a filter by name.
func (d *GlobalFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GlobalFilterSingleDataSourceModel
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

	var found *client.GlobalFilter
	for _, f := range filters {
		if f.Name == data.Name.ValueString() {
			found = &f
			break
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Global Filter Not Found",
			"No global filter with name \""+data.Name.ValueString()+"\" found in config "+data.ConfigID.ValueString(),
		)
		return
	}

	data.ID = types.StringValue(found.ID)
	data.Description = types.StringValue(found.Description)
	data.Source = types.StringValue(found.Source)
	data.Mdate = types.StringValue(found.Mdate)
	data.Active = types.BoolValue(found.Active)

	data.Tags = dsStringSliceToList(ctx, found.Tags, resp)

	data.Action = types.StringValue(found.Action)

	// Rule: interface{} -> JSON string
	if found.Rule != nil {
		ruleBytes, marshalErr := json.Marshal(found.Rule)
		if marshalErr != nil {
			resp.Diagnostics.AddError("Error Marshaling Rule", marshalErr.Error())
			return
		}
		data.Rule = types.StringValue(string(ruleBytes))
	} else {
		data.Rule = types.StringValue("{}")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
