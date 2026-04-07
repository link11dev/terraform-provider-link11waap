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

var _ datasource.DataSource = &MobileApplicationGroupsDataSource{}

// MobileApplicationGroupsDataSource defines the data source implementation
// for listing mobile application groups.
type MobileApplicationGroupsDataSource struct {
	client *client.Client
}

// MobileApplicationGroupsDataSourceModel defines the data model for the mobile
// application groups data source.
type MobileApplicationGroupsDataSourceModel struct {
	ConfigID                types.String                      `tfsdk:"config_id"`
	MobileApplicationGroups []MobileApplicationGroupDataModel `tfsdk:"mobile_application_groups"`
}

// MobileApplicationGroupDataModel defines the data model for a single mobile
type MobileApplicationGroupDataModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	UIDHeader    types.String `tfsdk:"uid_header"`
	Grace        types.String `tfsdk:"grace"`
	ActiveConfig types.Set    `tfsdk:"active_config"`
	Signatures   types.Set    `tfsdk:"signatures"`
}

// ActiveConfigDataModel defines the data model for the active configuration
// of a mobile application group.
type ActiveConfigDataModel struct {
	Active types.Bool   `tfsdk:"active"`
	JSON   types.String `tfsdk:"json"`
	Name   types.String `tfsdk:"name"`
}

// SignatureDataModel defines the data model for a signature associated with a
// mobile application group.
type SignatureDataModel struct {
	Active types.Bool   `tfsdk:"active"`
	Hash   types.String `tfsdk:"hash"`
	Name   types.String `tfsdk:"name"`
}

// NewMobileApplicationGroupsDataSource is a helper function to instantiate the
// mobile application groups data source.
func NewMobileApplicationGroupsDataSource() datasource.DataSource {
	return &MobileApplicationGroupsDataSource{}
}

// Metadata returns the data source type name.
func (d *MobileApplicationGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mobile_application_groups"
}

// Schema defines the schema for the mobile application groups data source.
func (d *MobileApplicationGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all mobile application groups in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"mobile_application_groups": schema.ListNestedAttribute{
				Description: "List of mobile application groups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.StringAttribute{Computed: true},
						"name":        schema.StringAttribute{Computed: true},
						"description": schema.StringAttribute{Computed: true},
						"uid_header":  schema.StringAttribute{Computed: true},
						"grace":       schema.StringAttribute{Computed: true},
						"active_config": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"active": schema.BoolAttribute{Computed: true},
									"json":   schema.StringAttribute{Computed: true},
									"name":   schema.StringAttribute{Computed: true},
								},
							},
						},
						"signatures": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"active": schema.BoolAttribute{Computed: true},
									"hash":   schema.StringAttribute{Computed: true},
									"name":   schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure is called by the framework to set up the data source with
// the provider's configured client.
func (d *MobileApplicationGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the list of mobile application groups for the specified configuration
// and sets the data source state accordingly.
func (d *MobileApplicationGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MobileApplicationGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := d.client.ListMobileApplicationGroups(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Mobile Application Groups",
			"Could not read mobile application groups: "+err.Error(),
		)
		return
	}

	data.MobileApplicationGroups = make([]MobileApplicationGroupDataModel, len(groups))
	for i, mag := range groups {
		model := MobileApplicationGroupDataModel{
			ID:          types.StringValue(mag.ID),
			Name:        types.StringValue(mag.Name),
			Description: types.StringValue(mag.Description),
			UIDHeader:   types.StringValue(mag.UIDHeader),
			Grace:       types.StringValue(mag.Grace),
		}

		// ActiveConfig
		if mag.ActiveConfig != nil {
			acModels := make([]ActiveConfigDataModel, len(mag.ActiveConfig))
			for j, ac := range mag.ActiveConfig {
				acModels[j] = ActiveConfigDataModel{
					Active: types.BoolValue(ac.Active),
					JSON:   types.StringValue(ac.JSON),
					Name:   types.StringValue(ac.Name),
				}
			}
			acSet, diags := types.SetValueFrom(ctx, types.ObjectType{
				AttrTypes: dsActiveConfigAttrTypes(),
			}, acModels)
			resp.Diagnostics.Append(diags...)
			model.ActiveConfig = acSet
		} else {
			model.ActiveConfig = types.SetNull(types.ObjectType{AttrTypes: dsActiveConfigAttrTypes()})
		}

		// Signatures
		if mag.Signatures != nil {
			sigModels := make([]SignatureDataModel, len(mag.Signatures))
			for j, sig := range mag.Signatures {
				sigModels[j] = SignatureDataModel{
					Active: types.BoolValue(sig.Active),
					Hash:   types.StringValue(sig.Hash),
					Name:   types.StringValue(sig.Name),
				}
			}
			sigSet, diags := types.SetValueFrom(ctx, types.ObjectType{
				AttrTypes: dsSignatureAttrTypes(),
			}, sigModels)
			resp.Diagnostics.Append(diags...)
			model.Signatures = sigSet
		} else {
			model.Signatures = types.SetNull(types.ObjectType{AttrTypes: dsSignatureAttrTypes()})
		}

		data.MobileApplicationGroups[i] = model
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func dsActiveConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"active": types.BoolType,
		"json":   types.StringType,
		"name":   types.StringType,
	}
}

func dsSignatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"active": types.BoolType,
		"hash":   types.StringType,
		"name":   types.StringType,
	}
}
