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

var _ datasource.DataSource = &RateLimitRulesDataSource{}

var dsTagFilterAttrTypes = map[string]attr.Type{
	"relation": types.StringType,
	"tags":     types.ListType{ElemType: types.StringType},
}

// RateLimitRulesDataSource implements the data source for listing rate limit rules
type RateLimitRulesDataSource struct {
	client *client.Client
}

// RateLimitRulesDataSourceModel describes the data source model for rate limit rules
type RateLimitRulesDataSourceModel struct {
	ConfigID       types.String             `tfsdk:"config_id"`
	RateLimitRules []RateLimitRuleDataModel `tfsdk:"rate_limit_rules"`
}

// RateLimitRuleDataModel describes the data model for a single rate limit rule
type RateLimitRuleDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Global      types.Bool   `tfsdk:"global"`
	Active      types.Bool   `tfsdk:"active"`
	Timeframe   types.Int64  `tfsdk:"timeframe"`
	Threshold   types.Int64  `tfsdk:"threshold"`
	TTL         types.Int64  `tfsdk:"ttl"`
	Action      types.String `tfsdk:"action"`
	Tags        types.List   `tfsdk:"tags"`
	Include     types.Object `tfsdk:"include"`
	Exclude     types.Object `tfsdk:"exclude"`
}

// NewRateLimitRulesDataSource returns a new instance of the rate limit rules data source
func NewRateLimitRulesDataSource() datasource.DataSource {
	return &RateLimitRulesDataSource{}
}

// Metadata returns the data source type name.
func (d *RateLimitRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rate_limit_rules"
}

// Schema defines the schema for the rate limit rules data source.
func (d *RateLimitRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all rate limit rules in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"rate_limit_rules": schema.ListNestedAttribute{
				Description: "List of rate limit rules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"global":      schema.BoolAttribute{Computed: true},
						"active":      schema.BoolAttribute{Computed: true},
						"timeframe":   schema.Int64Attribute{Computed: true},
						"threshold":   schema.Int64Attribute{Computed: true},
						"ttl":         schema.Int64Attribute{Computed: true},
						"action":      schema.StringAttribute{Computed: true},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
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
func (d *RateLimitRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the rate limit rules for the specified configuration and sets the data source state.
func (d *RateLimitRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data RateLimitRulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rules, err := d.client.ListRateLimitRules(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Rate Limit Rules",
			"Could not read rate limit rules: "+err.Error(),
		)
		return
	}

	data.RateLimitRules = make([]RateLimitRuleDataModel, len(rules))
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
		includeTagsList, d := types.ListValueFrom(ctx, types.StringType, includeTags)
		resp.Diagnostics.Append(d...)
		includeObj, d := types.ObjectValue(dsTagFilterAttrTypes, map[string]attr.Value{
			"relation": types.StringValue(r.Include.Relation),
			"tags":     includeTagsList,
		})
		resp.Diagnostics.Append(d...)

		// Exclude
		excludeTags := r.Exclude.Tags
		if excludeTags == nil {
			excludeTags = []string{}
		}
		excludeTagsList, d2 := types.ListValueFrom(ctx, types.StringType, excludeTags)
		resp.Diagnostics.Append(d2...)
		excludeObj, d2 := types.ObjectValue(dsTagFilterAttrTypes, map[string]attr.Value{
			"relation": types.StringValue(r.Exclude.Relation),
			"tags":     excludeTagsList,
		})
		resp.Diagnostics.Append(d2...)

		data.RateLimitRules[i] = RateLimitRuleDataModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			Description: types.StringValue(r.Description),
			Global:      types.BoolValue(r.Global),
			Active:      types.BoolValue(r.Active),
			Timeframe:   types.Int64Value(int64(r.Timeframe)),
			Threshold:   types.Int64Value(int64(r.Threshold)),
			TTL:         types.Int64Value(int64(r.TTL)),
			Action:      types.StringValue(r.Action),
			Tags:        tagsList,
			Include:     includeObj,
			Exclude:     excludeObj,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
