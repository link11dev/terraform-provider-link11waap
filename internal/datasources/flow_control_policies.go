package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &FlowControlPoliciesDataSource{}

// FlowControlPoliciesDataSource implements the data source for listing flow control policies
type FlowControlPoliciesDataSource struct {
	client *client.Client
}

// FlowControlPoliciesDataSourceModel describes the data source model for flow control policies
type FlowControlPoliciesDataSourceModel struct {
	ConfigID            types.String                 `tfsdk:"config_id"`
	FlowControlPolicies []FlowControlPolicyDataModel `tfsdk:"flow_control_policies"`
}

// FlowControlPolicyDataModel describes the data model for a single flow control policy
type FlowControlPolicyDataModel struct {
	ID          types.String               `tfsdk:"id"`
	Name        types.String               `tfsdk:"name"`
	Description types.String               `tfsdk:"description"`
	Active      types.Bool                 `tfsdk:"active"`
	Timeframe   types.Int64                `tfsdk:"timeframe"`
	Tags        types.List                 `tfsdk:"tags"`
	Include     types.List                 `tfsdk:"include"`
	Exclude     types.List                 `tfsdk:"exclude"`
	Key         []FlowControlKeyDataModel  `tfsdk:"key"`
	Steps       []FlowControlStepDataModel `tfsdk:"steps"`
}

// FlowControlKeyDataModel describes the data model for a flow control key entry
type FlowControlKeyDataModel struct {
	Attrs   types.String `tfsdk:"attrs"`
	Args    types.String `tfsdk:"args"`
	Plugins types.String `tfsdk:"plugins"`
	Cookies types.String `tfsdk:"cookies"`
	Headers types.String `tfsdk:"headers"`
}

// FlowControlStepDataModel describes the data model for a single flow step
type FlowControlStepDataModel struct {
	Method  types.String `tfsdk:"method"`
	URI     types.String `tfsdk:"uri"`
	Headers types.Map    `tfsdk:"headers"`
	Cookies types.Map    `tfsdk:"cookies"`
	Args    types.Map    `tfsdk:"args"`
	Plugins types.Map    `tfsdk:"plugins"`
}

// NewFlowControlPoliciesDataSource returns a new instance of the flow control policies data source
func NewFlowControlPoliciesDataSource() datasource.DataSource {
	return &FlowControlPoliciesDataSource{}
}

// Metadata returns the data source type name.
func (d *FlowControlPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flow_control_policies"
}

// Schema defines the schema for the flow control policies data source.
func (d *FlowControlPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all flow control policies in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"flow_control_policies": schema.ListNestedAttribute{
				Description: "List of flow control policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"active":      schema.BoolAttribute{Computed: true},
						"timeframe":   schema.Int64Attribute{Computed: true},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"include":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"exclude":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"key": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"attrs":   schema.StringAttribute{Computed: true},
									"args":    schema.StringAttribute{Computed: true},
									"plugins": schema.StringAttribute{Computed: true},
									"cookies": schema.StringAttribute{Computed: true},
									"headers": schema.StringAttribute{Computed: true},
								},
							},
						},
						"steps": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"method":  schema.StringAttribute{Computed: true},
									"uri":     schema.StringAttribute{Computed: true},
									"headers": schema.MapAttribute{Computed: true, ElementType: types.StringType},
									"cookies": schema.MapAttribute{Computed: true, ElementType: types.StringType},
									"args":    schema.MapAttribute{Computed: true, ElementType: types.StringType},
									"plugins": schema.MapAttribute{Computed: true, ElementType: types.StringType},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure initializes the data source with the provider client.
func (d *FlowControlPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the flow control policies for the specified configuration and sets the data source state.
func (d *FlowControlPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data FlowControlPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.client.ListFlowControlPolicies(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Flow Control Policies",
			"Could not read flow control policies: "+err.Error(),
		)
		return
	}

	data.FlowControlPolicies = make([]FlowControlPolicyDataModel, len(policies))
	for i, p := range policies {
		var tagsList types.List
		if p.Tags != nil {
			tl, diags := types.ListValueFrom(ctx, types.StringType, p.Tags)
			resp.Diagnostics.Append(diags...)
			tagsList = tl
		} else {
			tagsList = types.ListNull(types.StringType)
		}

		include := p.Include
		if include == nil {
			include = []string{}
		}
		includeList, diags := types.ListValueFrom(ctx, types.StringType, include)
		resp.Diagnostics.Append(diags...)

		exclude := p.Exclude
		if exclude == nil {
			exclude = []string{}
		}
		excludeList, diags := types.ListValueFrom(ctx, types.StringType, exclude)
		resp.Diagnostics.Append(diags...)

		data.FlowControlPolicies[i] = FlowControlPolicyDataModel{
			ID:          types.StringValue(p.ID),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			Active:      types.BoolValue(p.Active),
			Timeframe:   types.Int64Value(int64(p.Timeframe)),
			Tags:        tagsList,
			Include:     includeList,
			Exclude:     excludeList,
			Key:         flowControlKeysToData(p.Key),
			Steps:       flowControlStepsToData(ctx, p.Steps, &resp.Diagnostics),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// flowControlKeysToData converts []client.FlowControlKeyEntry to []FlowControlKeyDataModel.
func flowControlKeysToData(keys []client.FlowControlKeyEntry) []FlowControlKeyDataModel {
	result := make([]FlowControlKeyDataModel, 0, len(keys))
	for _, k := range keys {
		km := FlowControlKeyDataModel{
			Attrs:   types.StringNull(),
			Args:    types.StringNull(),
			Plugins: types.StringNull(),
			Cookies: types.StringNull(),
			Headers: types.StringNull(),
		}
		switch {
		case k.Attrs != nil:
			km.Attrs = types.StringValue(*k.Attrs)
		case k.Args != nil:
			km.Args = types.StringValue(*k.Args)
		case k.Plugins != nil:
			km.Plugins = types.StringValue(*k.Plugins)
		case k.Cookies != nil:
			km.Cookies = types.StringValue(*k.Cookies)
		case k.Headers != nil:
			km.Headers = types.StringValue(*k.Headers)
		}
		result = append(result, km)
	}
	return result
}

// flowControlStepsToData converts []client.FlowStepItem to []FlowControlStepDataModel.
func flowControlStepsToData(ctx context.Context, steps []client.FlowStepItem, diags *diag.Diagnostics) []FlowControlStepDataModel {
	result := make([]FlowControlStepDataModel, 0, len(steps))
	for _, s := range steps {
		sm := FlowControlStepDataModel{
			Method:  types.StringValue(s.Method),
			URI:     types.StringValue(s.URI),
			Headers: types.MapNull(types.StringType),
			Cookies: types.MapNull(types.StringType),
			Args:    types.MapNull(types.StringType),
			Plugins: types.MapNull(types.StringType),
		}
		if len(s.Headers) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Headers)
			diags.Append(d...)
			sm.Headers = m
		}
		if len(s.Cookies) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Cookies)
			diags.Append(d...)
			sm.Cookies = m
		}
		if len(s.Args) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Args)
			diags.Append(d...)
			sm.Args = m
		}
		if len(s.Plugins) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Plugins)
			diags.Append(d...)
			sm.Plugins = m
		}
		result = append(result, sm)
	}
	return result
}
