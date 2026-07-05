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

var _ datasource.DataSource = &DynamicRulesDataSource{}

// DynamicRulesDataSource implements the data source for listing dynamic rules
type DynamicRulesDataSource struct {
	client *client.Client
}

// DynamicRulesDataSourceModel describes the data source model for dynamic rules
type DynamicRulesDataSourceModel struct {
	ConfigID     types.String           `tfsdk:"config_id"`
	DynamicRules []DynamicRuleDataModel `tfsdk:"dynamic_rules"`
}

// DynamicRuleDataModel describes the data model for a single dynamic rule
type DynamicRuleDataModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	Threshold          types.Int64  `tfsdk:"threshold"`
	Timeframe          types.Int64  `tfsdk:"timeframe"`
	TTL                types.Int64  `tfsdk:"ttl"`
	Active             types.Bool   `tfsdk:"active"`
	OffloadIPFiltering types.Bool   `tfsdk:"offload_ip_filtering"`
	Target             types.String `tfsdk:"target"`
	Action             types.String `tfsdk:"action"`
	Tags               types.List   `tfsdk:"tags"`
	Include            types.Object `tfsdk:"include"`
	Exclude            types.Object `tfsdk:"exclude"`
}

// NewDynamicRulesDataSource returns a new instance of the dynamic rules data source
func NewDynamicRulesDataSource() datasource.DataSource {
	return &DynamicRulesDataSource{}
}

// Metadata returns the data source type name.
func (d *DynamicRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamic_rules"
}

// Schema defines the schema for the dynamic rules data source.
func (d *DynamicRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all dynamic rules in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"dynamic_rules": schema.ListNestedAttribute{
				Description: "List of dynamic rules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                   schema.StringAttribute{Computed: true},
						"name":                 schema.StringAttribute{Computed: true},
						"description":          schema.StringAttribute{Computed: true},
						"threshold":            schema.Int64Attribute{Computed: true},
						"timeframe":            schema.Int64Attribute{Computed: true},
						"ttl":                  schema.Int64Attribute{Computed: true},
						"active":               schema.BoolAttribute{Computed: true},
						"offload_ip_filtering": schema.BoolAttribute{Computed: true},
						"target":               schema.StringAttribute{Computed: true},
						"action":               schema.StringAttribute{Computed: true},
						"tags":                 schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"include": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"relation": schema.StringAttribute{Computed: true},
								"tags":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
							},
						},
						"exclude": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"relation": schema.StringAttribute{Computed: true},
								"tags":     schema.ListAttribute{Computed: true, ElementType: types.StringType},
							},
						},
					},
				},
			},
		},
	}
}

// Configure initializes the data source with the provider client.
func (d *DynamicRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the dynamic rules for the specified configuration and sets the data source state.
func (d *DynamicRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DynamicRulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rules, err := d.client.ListDynamicRules(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dynamic Rules",
			"Could not read dynamic rules: "+err.Error(),
		)
		return
	}

	data.DynamicRules = make([]DynamicRuleDataModel, len(rules))
	for i, r := range rules {
		var tagsList types.List
		if r.Tags != nil {
			tl, diags := types.ListValueFrom(ctx, types.StringType, r.Tags)
			resp.Diagnostics.Append(diags...)
			tagsList = tl
		} else {
			tagsList = types.ListNull(types.StringType)
		}

		// Include
		includeTags := r.Include.Tags
		if includeTags == nil {
			includeTags = []string{}
		}
		includeTagsList, dg := types.ListValueFrom(ctx, types.StringType, includeTags)
		resp.Diagnostics.Append(dg...)
		includeObj, dg := types.ObjectValue(dsTagFilterAttrTypes, map[string]attr.Value{
			"relation": types.StringValue(r.Include.Relation),
			"tags":     includeTagsList,
		})
		resp.Diagnostics.Append(dg...)

		// Exclude
		excludeTags := r.Exclude.Tags
		if excludeTags == nil {
			excludeTags = []string{}
		}
		excludeTagsList, dg2 := types.ListValueFrom(ctx, types.StringType, excludeTags)
		resp.Diagnostics.Append(dg2...)
		excludeObj, dg2 := types.ObjectValue(dsTagFilterAttrTypes, map[string]attr.Value{
			"relation": types.StringValue(r.Exclude.Relation),
			"tags":     excludeTagsList,
		})
		resp.Diagnostics.Append(dg2...)

		data.DynamicRules[i] = DynamicRuleDataModel{
			ID:                 types.StringValue(r.ID),
			Name:               types.StringValue(r.Name),
			Description:        types.StringValue(r.Description),
			Threshold:          types.Int64Value(int64(r.Threshold)),
			Timeframe:          types.Int64Value(int64(r.Timeframe)),
			TTL:                types.Int64Value(int64(r.TTL)),
			Active:             types.BoolValue(r.Active),
			OffloadIPFiltering: types.BoolValue(r.OffloadIPFiltering),
			Target:             types.StringValue(r.Target),
			Action:             types.StringValue(r.Action),
			Tags:               tagsList,
			Include:            includeObj,
			Exclude:            excludeObj,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
