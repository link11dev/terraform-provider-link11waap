package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ContentFilterRulesDataSource{}

// ContentFilterRulesDataSource implements the data source for listing content filter rules.
type ContentFilterRulesDataSource struct {
	client *client.Client
}

// ContentFilterRulesDataSourceModel describes the data source model for content filter rules.
type ContentFilterRulesDataSourceModel struct {
	ConfigID           types.String                 `tfsdk:"config_id"`
	Name               types.String                 `tfsdk:"name"`
	ContentFilterRules []ContentFilterRuleDataModel `tfsdk:"content_filter_rules"`
}

// ContentFilterRuleDataModel describes the data model for a single content filter rule.
type ContentFilterRuleDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Msg         types.String `tfsdk:"msg"`
	Operand     types.String `tfsdk:"operand"`
	Category    types.String `tfsdk:"category"`
	Subcategory types.String `tfsdk:"subcategory"`
	Risk        types.Int64  `tfsdk:"risk"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
}

// NewContentFilterRulesDataSource returns a new instance of the content filter rules data source.
func NewContentFilterRulesDataSource() datasource.DataSource {
	return &ContentFilterRulesDataSource{}
}

// Metadata returns the data source type name.
func (d *ContentFilterRulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_filter_rules"
}

// Schema defines the schema for the content filter rules data source.
func (d *ContentFilterRulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all content filter rules in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Rule name. If specified, only the rule with this name will be returned.",
				Optional:    true,
			},
			"content_filter_rules": schema.ListNestedAttribute{
				Description: "List of content filter rules.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true, Description: "Unique identifier."},
						"name":        schema.StringAttribute{Computed: true, Description: "Name of the rule."},
						"msg":         schema.StringAttribute{Computed: true, Description: "Log message for this rule."},
						"operand":     schema.StringAttribute{Computed: true, Description: "Matching domain(s) regex."},
						"category":    schema.StringAttribute{Computed: true, Description: "Category of the rule."},
						"subcategory": schema.StringAttribute{Computed: true, Description: "Subcategory of the rule."},
						"risk":        schema.Int64Attribute{Computed: true, Description: "Risk level (1-5)."},
						"description": schema.StringAttribute{Computed: true, Description: "Description."},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of tags."},
					},
				},
			},
		},
	}
}

// Configure initializes the data source with the provider client.
func (d *ContentFilterRulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the content filter rules for the specified configuration.
func (d *ContentFilterRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContentFilterRulesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	allRules, err := d.client.ListContentFilterRules(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Content Filter Rules",
			"Could not read content filter rules: "+err.Error(),
		)
		return
	}

	rules := allRules
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		name := data.Name.ValueString()
		rules = nil
		for _, r := range allRules {
			if r.Name == name {
				rules = []client.ContentFilterRule{r}
				break
			}
		}
		if len(rules) == 0 {
			resp.Diagnostics.AddError(
				"Content Filter Rule Not Found",
				"No content filter rule found with name: "+name,
			)
			return
		}
	}

	data.ContentFilterRules = make([]ContentFilterRuleDataModel, len(rules))
	for i, r := range rules {
		var tagsList types.List
		if len(r.Tags) > 0 {
			tl, diags := types.ListValueFrom(ctx, types.StringType, r.Tags)
			resp.Diagnostics.Append(diags...)
			tagsList = tl
		} else {
			tagsList = types.ListNull(types.StringType)
		}

		data.ContentFilterRules[i] = ContentFilterRuleDataModel{
			ID:          types.StringValue(r.ID),
			Name:        types.StringValue(r.Name),
			Msg:         types.StringValue(r.Msg),
			Operand:     types.StringValue(r.Operand),
			Category:    types.StringValue(r.Category),
			Subcategory: types.StringValue(r.Subcategory),
			Risk:        types.Int64Value(int64(r.Risk)),
			Description: types.StringValue(r.Description),
			Tags:        tagsList,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
