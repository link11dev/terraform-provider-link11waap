package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ActionsDataSource{}

// ActionsDataSource defines the data source for listing actions.
type ActionsDataSource struct {
	client *client.Client
}

// ActionsDataSourceModel describes the data model for the actions data source.
type ActionsDataSourceModel struct {
	ConfigID types.String      `tfsdk:"config_id"`
	ID       types.String      `tfsdk:"id"`
	Name     types.String      `tfsdk:"name"`
	Actions  []ActionDataModel `tfsdk:"actions"`
}

// ActionDataModel represents a single action in the data source.
type ActionDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Tags        types.List   `tfsdk:"tags"`
	Params      types.Object `tfsdk:"params"`
}

// NewActionsDataSource creates a new actions data source instance.
func NewActionsDataSource() datasource.DataSource {
	return &ActionsDataSource{}
}

// Metadata returns the data source type name.
func (d *ActionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_actions"
}

// Schema defines the schema for the actions data source.
func (d *ActionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists actions in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Action ID. If specified, only the action with this ID will be returned.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Action Name. If specified, only the action with this name will be returned.",
				Optional:    true,
			},
			"actions": schema.ListNestedAttribute{
				Description: "List of actions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Unique identifier."},
						"name":        schema.StringAttribute{Computed: true, Description: "Name of the action."},
						"description": schema.StringAttribute{Computed: true, Description: "Description."},
						"type":        schema.StringAttribute{Computed: true, Description: "Action type."},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of tags."},
						"params": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Action parameters.",
							Attributes: map[string]schema.Attribute{
								"content": schema.StringAttribute{Computed: true, Description: "Response body content."},
								"status":  schema.Int64Attribute{Computed: true, Description: "HTTP status code to return."},
								"headers": schema.MapAttribute{Computed: true, ElementType: types.StringType, Description: "Map of response header name to value."},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *ActionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the actions data source.
func (d *ActionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var actions []client.Action

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		a, err := d.client.GetAction(ctx, data.ConfigID.ValueString(), data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Action",
				"Could not read action with ID "+data.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		actions = []client.Action{*a}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		allActions, err := d.client.ListActions(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Actions",
				"Could not read actions: "+err.Error(),
			)
			return
		}
		for _, a := range allActions {
			if a.Name == data.Name.ValueString() {
				actions = []client.Action{a}
				break
			}
		}
		if len(actions) == 0 {
			resp.Diagnostics.AddError(
				"Action Not Found",
				"No action found with name: "+data.Name.ValueString(),
			)
			return
		}
	} else {
		var err error
		actions, err = d.client.ListActions(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Actions",
				"Could not read actions: "+err.Error(),
			)
			return
		}
	}

	data.Actions = make([]ActionDataModel, len(actions))
	for i := range actions {
		a := actions[i]

		var tagsList types.List
		if len(a.Tags) > 0 {
			tl, diags := types.ListValueFrom(ctx, types.StringType, a.Tags)
			resp.Diagnostics.Append(diags...)
			tagsList = tl
		} else {
			tagsList = types.ListNull(types.StringType)
		}

		params, diags := flattenActionParams(ctx, &a)
		resp.Diagnostics.Append(diags...)

		data.Actions[i] = ActionDataModel{
			ID:          types.StringValue(a.ID),
			Name:        types.StringValue(a.Name),
			Description: types.StringValue(a.Description),
			Type:        types.StringValue(a.Type),
			Tags:        tagsList,
			Params:      params,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// dsActionParamsAttrTypes returns the attribute type map for the action params object.
func dsActionParamsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"content": types.StringType,
		"status":  types.Int64Type,
		"headers": types.MapType{ElemType: types.StringType},
	}
}

// flattenActionParams converts the API params block into a Terraform object value.
func flattenActionParams(ctx context.Context, a *client.Action) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if a.Params == nil {
		return types.ObjectNull(dsActionParamsAttrTypes()), diags
	}

	var headers types.Map
	if a.Params.Headers != nil {
		hv, d := types.MapValueFrom(ctx, types.StringType, a.Params.Headers)
		diags.Append(d...)
		headers = hv
	} else {
		headers = types.MapNull(types.StringType)
	}

	status := types.Int64Null()
	if a.Params.Status != nil {
		status = types.Int64Value(int64(*a.Params.Status))
	}

	obj, d := types.ObjectValue(dsActionParamsAttrTypes(), map[string]attr.Value{
		"content": types.StringValue(a.Params.Content),
		"status":  status,
		"headers": headers,
	})
	diags.Append(d...)
	return obj, diags
}
