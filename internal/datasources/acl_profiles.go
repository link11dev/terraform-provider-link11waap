// Package datasources contains the data source implementations for the Link11 WAAP Terraform provider.
package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ACLProfilesDataSource{}

// ACLProfilesDataSource defines the data source for listing ACL profiles.
type ACLProfilesDataSource struct {
	client *client.Client
}

// ACLProfilesDataSourceModel describes the data model for the ACL profiles data source.
type ACLProfilesDataSourceModel struct {
	ConfigID    types.String          `tfsdk:"config_id"`
	ACLProfiles []ACLProfileDataModel `tfsdk:"acl_profiles"`
}

// ACLProfileDataModel represents a single ACL profile in the data source.
type ACLProfileDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
	Action      types.String `tfsdk:"action"`
	Allow       types.List   `tfsdk:"allow"`
	AllowBot    types.List   `tfsdk:"allow_bot"`
	Deny        types.List   `tfsdk:"deny"`
	DenyBot     types.List   `tfsdk:"deny_bot"`
	ForceDeny   types.List   `tfsdk:"force_deny"`
	Passthrough types.List   `tfsdk:"passthrough"`
}

// NewACLProfilesDataSource creates a new ACL profiles data source instance.
func NewACLProfilesDataSource() datasource.DataSource {
	return &ACLProfilesDataSource{}
}

// Metadata returns the data source type name.
func (d *ACLProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_acl_profiles"
}

// Schema defines the schema for the ACL profiles data source.
func (d *ACLProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all ACL profiles in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"acl_profiles": schema.ListNestedAttribute{
				Description: "List of ACL profiles.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"tags":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"action":      schema.StringAttribute{Computed: true},
						"allow":       schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"allow_bot":   schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"deny":        schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"deny_bot":    schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"force_deny":  schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"passthrough": schema.ListAttribute{Computed: true, ElementType: types.StringType},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *ACLProfilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the ACL profiles data source.
func (d *ACLProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ACLProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	profiles, err := d.client.ListACLProfiles(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading ACL Profiles",
			"Could not read ACL profiles: "+err.Error(),
		)
		return
	}

	data.ACLProfiles = make([]ACLProfileDataModel, len(profiles))
	for i, p := range profiles {
		model := ACLProfileDataModel{
			ID:          types.StringValue(p.ID),
			Name:        types.StringValue(p.Name),
			Description: types.StringValue(p.Description),
			Action:      types.StringValue(p.Action),
		}

		model.Tags = dsStringSliceToList(ctx, p.Tags, resp)
		model.Allow = dsStringSliceToList(ctx, p.Allow, resp)
		model.AllowBot = dsStringSliceToList(ctx, p.AllowBot, resp)
		model.Deny = dsStringSliceToList(ctx, p.Deny, resp)
		model.DenyBot = dsStringSliceToList(ctx, p.DenyBot, resp)
		model.ForceDeny = dsStringSliceToList(ctx, p.ForceDeny, resp)
		model.Passthrough = dsStringSliceToList(ctx, p.Passthrough, resp)

		data.ACLProfiles[i] = model
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// dsStringSliceToList converts a []string to a types.List for data source state.
func dsStringSliceToList(ctx context.Context, slice []string, resp *datasource.ReadResponse) types.List {
	if slice != nil {
		list, diags := types.ListValueFrom(ctx, types.StringType, slice)
		resp.Diagnostics.Append(diags...)
		return list
	}
	return types.ListNull(types.StringType)
}
