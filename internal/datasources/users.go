package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &UsersDataSource{}

// UsersDataSource defines the data source for listing all users
type UsersDataSource struct {
	client *client.Client
}

// UsersDataSourceModel represents the data model for the users data source
type UsersDataSourceModel struct {
	Users []UserDataModel `tfsdk:"users"`
}

// UserDataModel represents a user in the data source
type UserDataModel struct {
	ID          types.String `tfsdk:"id"`
	ACL         types.Int64  `tfsdk:"acl"`
	ContactName types.String `tfsdk:"contact_name"`
	Email       types.String `tfsdk:"email"`
	Mobile      types.String `tfsdk:"mobile"`
	OrgID       types.String `tfsdk:"org_id"`
	OrgName     types.String `tfsdk:"org_name"`
}

// NewUsersDataSource creates a new instance of the users data source
func NewUsersDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

// Metadata returns the data source type name
func (d *UsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

// Schema defines the schema for the users data source
func (d *UsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all users across all organizations.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description: "List of users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":           schema.StringAttribute{Computed: true},
						"acl":          schema.Int64Attribute{Computed: true},
						"contact_name": schema.StringAttribute{Computed: true},
						"email":        schema.StringAttribute{Computed: true},
						"mobile":       schema.StringAttribute{Computed: true},
						"org_id":       schema.StringAttribute{Computed: true},
						"org_name":     schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

// Configure sets the client for the data source
func (d *UsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the list of users and sets the state
func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	orgs, err := d.client.ListUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Users",
			"Could not read users: "+err.Error(),
		)
		return
	}

	// Flatten organizations into a flat list of users
	var allUsers []UserDataModel
	for _, org := range orgs {
		for _, user := range org.Users {
			allUsers = append(allUsers, UserDataModel{
				ID:          types.StringValue(user.ID),
				ACL:         types.Int64Value(int64(user.ACL)),
				ContactName: types.StringValue(user.ContactName),
				Email:       types.StringValue(user.Email),
				Mobile:      types.StringValue(user.Mobile),
				OrgID:       types.StringValue(user.OrgID),
				OrgName:     types.StringValue(user.OrgName),
			})
		}
	}

	data.Users = allUsers

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
