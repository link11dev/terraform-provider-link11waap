package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &SecurityPoliciesDataSource{}

// SecurityPoliciesDataSource defines the data source for listing security policies in a configuration
type SecurityPoliciesDataSource struct {
	client *client.Client
}

// SecurityPoliciesDataSourceModel represents the data model for the security policies data source
type SecurityPoliciesDataSourceModel struct {
	ConfigID         types.String              `tfsdk:"config_id"`
	SecurityPolicies []SecurityPolicyDataModel `tfsdk:"security_policies"`
}

// SecurityPolicyDataModel represents a security policy in the data source
type SecurityPolicyDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
}

// NewSecurityPoliciesDataSource creates a new instance of the security policies data source
func NewSecurityPoliciesDataSource() datasource.DataSource {
	return &SecurityPoliciesDataSource{}
}

// Metadata returns the data source type name
func (d *SecurityPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_policies"
}

// Schema defines the schema for the security policies data source
func (d *SecurityPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all security policies in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"security_policies": schema.ListNestedAttribute{
				Description: "List of security policies.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

// Configure sets the client for the data source
func (d *SecurityPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the list of security policies for the given configuration and sets the state
func (d *SecurityPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityPoliciesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.client.ListSecurityPolicies(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Security Policies",
			"Could not read security policies: "+err.Error(),
		)
		return
	}

	data.SecurityPolicies = make([]SecurityPolicyDataModel, len(policies))
	for i, sp := range policies {
		var tagsList types.List
		if sp.Tags != nil {
			tl, diags := types.ListValueFrom(ctx, types.StringType, sp.Tags)
			resp.Diagnostics.Append(diags...)
			tagsList = tl
		} else {
			tagsList = types.ListNull(types.StringType)
		}

		data.SecurityPolicies[i] = SecurityPolicyDataModel{
			ID:          types.StringValue(sp.ID),
			Name:        types.StringValue(sp.Name),
			Description: types.StringValue(sp.Description),
			Tags:        tagsList,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
