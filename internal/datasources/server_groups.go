package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ServerGroupsDataSource{}

// ServerGroupsDataSource defines the data source for listing server groups in a configuration
type ServerGroupsDataSource struct {
	client *client.Client
}

// ServerGroupsDataSourceModel represents the data model for the server groups data source
type ServerGroupsDataSourceModel struct {
	ConfigID     types.String           `tfsdk:"config_id"`
	ServerGroups []ServerGroupDataModel `tfsdk:"server_groups"`
}

// ServerGroupDataModel represents a server group in the data source
type ServerGroupDataModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	ServerNames            types.List   `tfsdk:"server_names"`
	SecurityPolicy         types.String `tfsdk:"security_policy"`
	RoutingProfile         types.String `tfsdk:"routing_profile"`
	ProxyTemplate          types.String `tfsdk:"proxy_template"`
	ChallengeCookieDomain  types.String `tfsdk:"challenge_cookie_domain"`
	SSLCertificate         types.String `tfsdk:"ssl_certificate"`
	ClientCertificate      types.String `tfsdk:"client_certificate"`
	ClientCertificateMode  types.String `tfsdk:"client_certificate_mode"`
	MobileApplicationGroup types.String `tfsdk:"mobile_application_group"`
}

// NewServerGroupsDataSource creates a new instance of the server groups data source
func NewServerGroupsDataSource() datasource.DataSource {
	return &ServerGroupsDataSource{}
}

// Metadata returns the data source type name
func (d *ServerGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_groups"
}

// Schema defines the schema for the server groups data source
func (d *ServerGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all server groups in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"server_groups": schema.ListNestedAttribute{
				Description: "List of server groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{Computed: true},
						"name":                     schema.StringAttribute{Computed: true},
						"description":              schema.StringAttribute{Computed: true},
						"server_names":             schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"security_policy":          schema.StringAttribute{Computed: true},
						"routing_profile":          schema.StringAttribute{Computed: true},
						"proxy_template":           schema.StringAttribute{Computed: true},
						"challenge_cookie_domain":  schema.StringAttribute{Computed: true},
						"ssl_certificate":          schema.StringAttribute{Computed: true},
						"client_certificate":       schema.StringAttribute{Computed: true},
						"client_certificate_mode":  schema.StringAttribute{Computed: true},
						"mobile_application_group": schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

// Configure sets the client for the data source
func (d *ServerGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the list of server groups for the given configuration and sets the state
func (d *ServerGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ServerGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	serverGroups, err := d.client.ListServerGroups(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Server Groups",
			"Could not read server groups: "+err.Error(),
		)
		return
	}

	data.ServerGroups = make([]ServerGroupDataModel, len(serverGroups))
	for i, sg := range serverGroups {
		serverNames, diags := types.ListValueFrom(ctx, types.StringType, sg.ServerNames)
		resp.Diagnostics.Append(diags...)

		data.ServerGroups[i] = ServerGroupDataModel{
			ID:                     types.StringValue(sg.ID),
			Name:                   types.StringValue(sg.Name),
			Description:            types.StringValue(sg.Description),
			ServerNames:            serverNames,
			SecurityPolicy:         types.StringValue(sg.SecurityPolicy),
			RoutingProfile:         types.StringValue(sg.RoutingProfile),
			ProxyTemplate:          types.StringValue(sg.ProxyTemplate),
			ChallengeCookieDomain:  types.StringValue(sg.ChallengeCookieDomain),
			SSLCertificate:         types.StringValue(sg.SSLCertificate),
			ClientCertificate:      types.StringValue(sg.ClientCertificate),
			ClientCertificateMode:  types.StringValue(sg.ClientCertificateMode),
			MobileApplicationGroup: types.StringValue(sg.MobileApplicationGroup),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
