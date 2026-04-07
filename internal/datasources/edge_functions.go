package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &EdgeFunctionsDataSource{}

// EdgeFunctionsDataSource implements the data source for listing edge functions
type EdgeFunctionsDataSource struct {
	client *client.Client
}

// EdgeFunctionsDataSourceModel describes the data source model for edge functions
type EdgeFunctionsDataSourceModel struct {
	ConfigID      types.String            `tfsdk:"config_id"`
	EdgeFunctions []EdgeFunctionDataModel `tfsdk:"edge_functions"`
}

// EdgeFunctionDataModel describes the data model for a single edge function
type EdgeFunctionDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Code        types.String `tfsdk:"code"`
	Phase       types.String `tfsdk:"phase"`
}

// NewEdgeFunctionsDataSource returns a new instance of the edge functions data source
func NewEdgeFunctionsDataSource() datasource.DataSource {
	return &EdgeFunctionsDataSource{}
}

// Metadata returns the data source type name.
func (d *EdgeFunctionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_edge_functions"
}

// Schema defines the schema for the edge functions data source.
func (d *EdgeFunctionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all edge functions in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"edge_functions": schema.ListNestedAttribute{
				Description: "List of edge functions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Unique identifier."},
						"name":        schema.StringAttribute{Computed: true, Description: "Name of the edge function."},
						"description": schema.StringAttribute{Computed: true, Description: "Description."},
						"code":        schema.StringAttribute{Computed: true, Description: "Lua source code."},
						"phase":       schema.StringAttribute{Computed: true, Description: "Execution phase (request or response)."},
					},
				},
			},
		},
	}
}

// Configure initializes the data source with the provider client.
func (d *EdgeFunctionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the edge functions for the specified configuration.
func (d *EdgeFunctionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EdgeFunctionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	functions, err := d.client.ListEdgeFunctions(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Edge Functions",
			"Could not read edge functions: "+err.Error(),
		)
		return
	}

	data.EdgeFunctions = make([]EdgeFunctionDataModel, len(functions))
	for i, f := range functions {
		data.EdgeFunctions[i] = EdgeFunctionDataModel{
			ID:          types.StringValue(f.ID),
			Name:        types.StringValue(f.Name),
			Description: types.StringValue(f.Description),
			Code:        types.StringValue(f.Code),
			Phase:       types.StringValue(f.Phase),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
